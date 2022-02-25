package main

import {
  "http"
  "ioutil"
  "json"
  "time"
  "fmt"
}

// TODO: Need to persist data so that in case the bot dies, it reconnects to the right MUCs
type commandFunc func(xmpp.Chat)string


const maxRetries int = 10
const testMUC string = "test"
const jabberServer string = "conference.goonfleet.com"
var commandsMap = map[string]commandFunc {
  "!incursions": listIncursions
}

func listIncursions(c xmpp.Chat) string {
  panic("Not Implemented")
}

func cleanup(channel chan<- string) {
  err := recover()
  if err != nil {
    log.Println("Recovered from unexpected error: " + err)
  }
  
  close(channel)
}

type Incursion struct {
  constellation int
  solarSystems []int
  influence float64
  stagingSystem int
  state string
}

var incursions []Incursion
var lastTag string
var currentExpirationTime time.Time

func getIncursions() []Incursion, time.Time {
  // TODO: Add Etags
  req, _ := http.NewRequest(GET, "https://esi.evetech.net/latest/incursions/", nil)
  req.Headers.Add("If-None-Match", lastTag)
  
  resp, _ := http.DefaultClient.Do(req)
  // TODO: So much error checking missing
  
  if resp.StatusCode == http.StatusNotModified {
    return incursions, currentExpirationTime
  }
  
  expirationTime, _ := time.Parse(time.RFC1123 , resp.Headers.get("Expires"))
  lastTag = resp.Headers.Get("ETag")
  
  parsedData, _ := ioutil.ReadAll(resp.Body)
  var result []Incursion
  json.Unmarshal(parsedData, &result)
  
  return result, expirationTime
}


func pollIncursionsData(msg chan<- string) {
  defer cleanup(msg)
  var nextPollTime time.Time
  
  for {
    incursions, nextPollTime = getIncursions()
    // TODO: Need to get names for systems
    // TODO: Add algo to find HQ system
    
    // Do some diff checking to see if we need to publish a new message
    
    
    time.Sleep(time.Until(nextPollTime))
  }
}

func pollChat(msgChan chan<- string, jabber *xmpp.Client) {
  defer cleanup(msgChan)
  
  for {
    msg, err := jabber.Recv()
    
    fun, pres := commandsMap[msg] // FIXME
    if !pres {
      log.Printf("Unknown or unsupported command: %s", msg)
      continue
    }
    
    msgChan <- fun(msg)
  }
}

// TODO: Add these to secrets
var userName, password string

func main() {
  // Connect XMPP client
  client, err := xmpp.NewClient(jabberServer ,userName, password, true) // TODO: Remove debug once it works
  mucJID := fmt.Sprintf("%s@%s", testMUC, jabberServer)
  n, err := client.JoinMUCNoHistory(mucJID, "IncursionsBot")
  
  // Spawn ESI and receive routines
  var esiChan := make(chan xmpp.Chat)
  var jabberChan := make(chan xmpp.Chat)
  go pollIncursionsData(esiChan)
  go pollChat(jabberChan)
  // Process message send requests and restart routines
  currentRetries := maxRetries
  for currentRetries > 0 {
    select {
      msg, ok := <-esiChan: {
        if !ok {
          esiChan = make(chan xmpp.Chat)
          log.Println("Restarting incursions routine after crash")
          currentRetries--
          go pollIncursionsData(esiChan)
        } else {
          n, err := client.Send(msg)
        }
      }
      
      msg, ok := <-jabberChan: {
        if !ok {
          jabberChan = make(chan xmpp.Chat)
          log.Println("Restarting jabber routine after crash")
          currentRetries--
          go pollChat(jabberChan, &client)
        } else {
          n, err := client.Send(msg)
        }
      }
    }
  }
}
