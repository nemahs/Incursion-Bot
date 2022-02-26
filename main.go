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
// TODO: Fix all the MUC stuff to not use testMUC

const maxRetries int = 10
const testMUC string = "testbot"
const jabberServer string = "conference.goonfleet.com"
const homeSystem int = 30004759 // 1DQ1-A
const botNick string = "IncursionBot"
const commandPrefix byte = '!'

var commandsMap CommandMap
var incursions IncursionList

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
    responseText += fmt.Sprintf("%s - Influence: %.2f%% - Status: %s - %d jumps \n",
    incursion.ToString(),
    incursion.Influence * 100, // Convert to % for easier reading
    incursion.State,
    incursion.Distance)
  }

  response.Text = responseText

  return response
}

func newGroupMessage(muc string, text string) xmpp.Chat {
  return xmpp.Chat{
    Remote: fmt.Sprintf("%s@%s", muc, jabberServer),
    Type: "groupchat",
    Text: text,
  }
}

func pollIncursionsData(msgChan chan<- xmpp.Chat) {
  defer cleanup(msgChan)
  var nextPollTime time.Time
  firstRun := true
  
  for {
    var newIncursionList IncursionList
    var incursionResponses []IncursionResponse

    incursionResponses, nextPollTime = getIncursions()
    
    for _, incursionData := range incursionResponses {
      existingIncursion := incursions.find(incursionData.StagingID)

      if existingIncursion == nil {
        newIncursion := createNewIncursion(incursionData)

        // Make new incursion
        newIncursionList = append(newIncursionList, newIncursion)
        if !firstRun {
          msgText := fmt.Sprintf("New incursion detected in %s - %d jumps", newIncursion.ToString(), newIncursion.Distance)
          msgChan <- newGroupMessage(testMUC, msgText)
        }
      } else {
        // Update data and check if anything changed
        if updateIncursion(existingIncursion, incursionData) {
          msgText := fmt.Sprintf("%s changed state to %s", existingIncursion.ToString(), existingIncursion.State)
          msgChan <- newGroupMessage(testMUC, msgText)
        }

        newIncursionList = append(newIncursionList, *existingIncursion)
      }
    }

    log.Printf("Comparing %+v to %+v", incursions, newIncursionList)
    for _, existing := range incursions {
      if newIncursionList.find(existing.StagingID) == nil {
        msgText := fmt.Sprintf("Incursion in %s despawned", existing.ToString())
        msgChan <- newGroupMessage(testMUC, msgText)
      }
    }

    incursions = newIncursionList

    firstRun = false
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
    
    if len(chatMsg.Text) == 0 || chatMsg.Text[0] != commandPrefix {
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

func getUptime(msg xmpp.Chat) xmpp.Chat {
  currentUptime := time.Since(startTime)
  msgText := fmt.Sprintf("Bot has been up for: %s", currentUptime)

  return newGroupMessage(testMUC, msgText)
}


var startTime time.Time = time.Now()
func main() {
  commandsMap = NewCommandMap()
  commandsMap.AddCommand("!incursions", listIncursions, "Lists the current incursions")
  commandsMap.AddCommand("!uptime", getUptime, "Gets the current bot uptime")

  userName := flag.String("username", "", "Username for Jabber")
  password := flag.String("password", "", "Password for Jabber")
  flag.Parse()
  checkESI()

  // Connect XMPP client
  log.Println("Creating client...")
  client, err := xmpp.NewClientNoTLS(jabberServer, *userName, *password, false)

  if err != nil {
    log.Fatalln("Failed to init client", err)
  }

  mucJID := fmt.Sprintf("%s@%s", testMUC, jabberServer)
  _, err = client.JoinMUCNoHistory(mucJID, botNick)

  if err != nil { log.Println("Failed to join MUC", err) }

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

  log.Fatalln("Max retries reached, shutting down...")
}
