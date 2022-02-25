package main

import {
  "http"
  "ioutil"
  "json"
  "time"
}

const maxRetries := 10

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
func getIncursions() []Incursion, time.Time {
  // TODO: Add Etags
  resp := http.Get("https://esi.evetech.net/latest/incursions/")
  
  // TODO: So much error checking missing
  
  expirationTime := time.Parse(time.RFC1123 , resp.Headers.get("Expires"))
  
  parsedData := ioutil.ReadAll(resp.body)
  var result []Incursion
  json.Unmarshal(parsedData, &result)
  
  return result, expirationTime
}


func pollIncursionsData(msg chan<- string) {
  defer cleanup(msg)
  var nextPollTime time.Time
  
  for {
    incursions, nextPollTime = getIncursions()
    
    // Do some diff checking to see if we need to publish a new message
    
    
    time.Sleep(time.Until(nextPollTime))
  }
}

func pollChat(msg chan<- string) {
  defer cleanup(msg)
}

func main() {
  // Connect XMPP client
  // Spawn ESI and receive routines
  var esiChan := make(chan string)
  var jabberChan := make(chan string)
  go pollIncursionsData(esiChan)
  go pollChat(jabberChan)
  // Process message send requests and restart routines
  currentRetries := maxRetries
  for currentRetries > 0 {
    select {
      msg, ok := <-esiChan: {
        if !ok {
          esiChan = make(chan string)
          log.Println("Restarting incursions routine after crash")
          go pollIncursionsData(esiChan)
        } else {
          //Send msg here 
        }
      }
      
      msg, ok := <-jabberChan: {
        if !ok {
          jabberChan = make(chan string)
          log.Println("Restarting jabber routine after crash")
          go pollChat(jabberChan)
        } else {
          //Send msg here
        }
      }
    }
  }
}
