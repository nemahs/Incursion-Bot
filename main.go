package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mattn/go-xmpp"
)

// TODO: Need to persist data so that in case the bot dies, it reconnects to the right MUCs
// In addition, persistance can let the bot know how much time is left on a particular spawn

const maxRetries int = 10
const testMUC string = "testbot"
const jabberServer string = "conference.goonfleet.com"
const homeSystem int = 30004759


type Incursion struct {
  Constellation string
  HQSystem string
  Influence float64
  Region string
  State string
  Security string
  SecStatus float64
  Distance int
}

var commandsMap CommandMap

func cleanup(channel chan<- xmpp.Chat) {
  err := recover()
  if err != nil {
    log.Println("Recovered from unexpected error: ", err)
  }
  
  close(channel)
}


func listIncursions(c xmpp.Chat) xmpp.Chat {
  response := xmpp.Chat {
    Remote: parseMuc(c.Remote, jabberServer),
    Type: c.Type,
  }
  responseText := "\n"

  for _, incursion := range incursions {
    responseText += fmt.Sprintf("%s {%.2f} (%s - %s) - Influence: %.2f%% - Status: %s - %d jumps \n",
    incursion.HQSystem, // TODO: Actually make this the HQ system and not the staging
    incursion.SecStatus,
    incursion.Constellation,
    incursion.Region,
    incursion.Influence * 100, // Convert to % for easier reading
    incursion.State,
    incursion.Distance)
  }

  response.Text = responseText

  return response
}


var incursions []Incursion

func pollIncursionsData(msg chan<- xmpp.Chat) {
  defer cleanup(msg)
  var nextPollTime time.Time
  
  for {
    incursionData, nextPollTime = getIncursions()
    incursions = nil
    
    for _, incursion := range incursionData {
      stagingData := getSystemInfo(incursion.Staging_Solar_System_Id)


      if stagingData.Security_Class == HighSec {
        continue // No one cares about high sec
      }

      constData := getConstInfo(incursion.Constellation_Id)
      names := getNames([]int{constData.Region_Id, incursion.Staging_Solar_System_Id})
      distance := GetRouteLength(homeSystem, incursion.Staging_Solar_System_Id)

      newIncursion := Incursion{
        Constellation: constData.Name,
        HQSystem: names[incursion.Staging_Solar_System_Id],
        Influence: incursion.Influence,
        Region: names[constData.Region_Id],
        State: incursion.State,
        SecStatus: stagingData.Security_Status,
        Security: string(stagingData.Security_Class),
        Distance: distance,
      }

      log.Printf("Incursion: %+v", newIncursion)
      incursions = append(incursions, newIncursion)
    }
    
    // TODO: Do some diff checking to see if we need to publish a new message
    
    time.Sleep(time.Until(nextPollTime))
  }
}

func pollChat(msgChan chan<- xmpp.Chat, jabber *xmpp.Client) {
  defer cleanup(msgChan)
  
  for {
    msg, err := jabber.Recv()

    if err != nil {
      log.Println("Error encountered receiving message: ", err)
    }

    chatMsg, ok := msg.(xmpp.Chat)

    if !ok {
      // Not a chat message, we don't care about it
      continue
    }
    
    if len(chatMsg.Text) == 0 || chatMsg.Text[0] != '!' {
      //Not a command, ignore
      continue
    }

    function, present := commandsMap.GetFunction(chatMsg.Text)
    if !present {
      log.Printf("Unknown or unsupported command: %s", chatMsg.Text)
      continue
    }

    
    msgChan <- function(chatMsg)
  }
}

func main() {
  commandsMap = NewCommandMap()
  commandsMap.AddCommand("!incursions", listIncursions, "Lists the current incursions")

  userName := flag.String("username", "", "Username for Jabber")
  password := flag.String("password", "", "Password for Jabber")
  flag.Parse()

  // Connect XMPP client
  log.Println("Creating client...")
  client, err := xmpp.NewClientNoTLS(jabberServer, *userName, *password, false)

  if err != nil {
    log.Fatalln("Failed to init client", err)
  }

  mucJID := fmt.Sprintf("%s@%s", testMUC, jabberServer)
  _, err = client.JoinMUCNoHistory(mucJID, "IncursionsBot")
  
  log.Println("Results from JoinMUC:", err)

  // Spawn ESI and receive routines
  log.Println("Client created, starting routines...")
  esiChan := make(chan xmpp.Chat)
  jabberChan := make(chan xmpp.Chat)
  go pollIncursionsData(esiChan)
  go pollChat(jabberChan, client)
  // Process message send requests and restart routines
  currentRetries := maxRetries
  for currentRetries > 0 {
    select {
      case msg, ok := <-esiChan: {
        if !ok {
          esiChan = make(chan xmpp.Chat)
          log.Println("Restarting incursions routine after crash")
          currentRetries--
          go pollIncursionsData(esiChan)
        } else {
          _, err := client.Send(msg)

          if err != nil { log.Println(err) }
        }
      }
      
      case msg, ok := <-jabberChan: {
        if !ok {
          jabberChan = make(chan xmpp.Chat)
          log.Println("Restarting jabber routine after crash")
          currentRetries--
          go pollChat(jabberChan, client)
        } else {
          _, err := client.Send(msg)

          if err != nil { log.Println(err) }
        }
      }
    }
  }
}
