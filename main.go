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
const jabberServer string = "conference.goonfleet.com"
const homeSystem int = 30004759 // 1DQ1-A
const botNick string = "IncursionBot"
const commandPrefix byte = '!'

var commandsMap CommandMap
var incursions IncursionList
var jabberChannel *string

// Recovers from any uncaught panics so that the main thread
// can restart the routine.
func cleanup(channel chan<- xmpp.Chat) {
  err := recover()
  if err != nil {
    log.Println("Recovered from unexpected error: ", err)
  }
  
  close(channel)
}

// Periodically polls ESI to get incursion data, and notifies chat of any changes
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
        // No existing incursion found, make a new one
        newIncursion := createNewIncursion(incursionData)
        newIncursionList = append(newIncursionList, newIncursion)

        if !firstRun { // Don't want to spam chats with "NEW INCURSION" whenever the bot starts, so notifications are inhibited on the first run
          msgText := fmt.Sprintf("New incursion detected in %s - %d jumps", newIncursion.ToString(), newIncursion.Distance)
          msgChan <- newGroupMessage(*jabberChannel, msgText)
        }
      } else {
        // Update data and check if anything changed
        if updateIncursion(existingIncursion, incursionData) {
          msgText := fmt.Sprintf("%s changed state to %s", existingIncursion.ToString(), existingIncursion.State)
          msgChan <- newGroupMessage(*jabberChannel, msgText)
        }

        newIncursionList = append(newIncursionList, *existingIncursion)
      }
    }

    // Check if any incursions have despawned and report
    for _, existing := range incursions {
      if newIncursionList.find(existing.StagingID) == nil {
        msgText := fmt.Sprintf("Incursion in %s despawned", existing.ToString())
        msgChan <- newGroupMessage(*jabberChannel, msgText)
      }
    }

    incursions = newIncursionList
    firstRun = false
    time.Sleep(time.Until(nextPollTime))
  }
}

// Polls jabber and processes any commands received
func pollChat(msgChan chan<- xmpp.Chat, jabber *xmpp.Client) {
  defer cleanup(msgChan)
  
  for {
    msg, err := jabber.Recv()

    if err != nil {
      log.Println("Error encountered receiving message: ", err)
      continue
    }

    chatMsg, ok := msg.(xmpp.Chat)
    if !ok { continue } // Not a chat message, we don't care about it
    
    if len(chatMsg.Text) == 0 || chatMsg.Text[0] != commandPrefix {
      //Not a command, ignore
      continue
    }

    // Slice off the command prefix
    function, present := commandsMap.GetFunction(chatMsg.Text[1:])
    if !present {
      log.Printf("Unknown or unsupported command: %s", chatMsg.Text)
      continue
    }

    msgChan <- function(chatMsg)
  }
}

// ------------- COMMANDS --------------------

var startTime time.Time = time.Now()
// Respond with the amount of time the bot's been up
func getUptime(msg xmpp.Chat) xmpp.Chat {
  currentUptime := time.Since(startTime)
  msgText := fmt.Sprintf("Bot has been up for: %s", currentUptime)

  return createReply(msg, msgText)
}

func printESIStatus(msg xmpp.Chat) xmpp.Chat {
  log.Println("Checking ESI status...")
  log.Println(msg.Remote)
  var status string
  if checkESI() { status = "GOOD" } else { status = "BAD" }
  msgText := fmt.Sprintf("Connection to ESI is %s", status)
  return createReply(msg, msgText)
}

func listIncursions(msg xmpp.Chat) xmpp.Chat {
  responseText := "\n"

  for _, incursion := range incursions {
    responseText += fmt.Sprintf("%s - Influence: %.2f%% - Status: %s - %d jumps \n",
    incursion.ToString(),
    incursion.Influence * 100, // Convert to % for easier reading
    incursion.State,
    incursion.Distance)
  }

  return createReply(msg, responseText)
}



func main() {
  commandsMap = NewCommandMap()
  commandsMap.AddCommand("incursions", listIncursions, "Lists the current incursions")
  commandsMap.AddCommand("uptime", getUptime, "Gets the current bot uptime")
  commandsMap.AddCommand("esi", printESIStatus, "Prints the bot's ESI connection status")

  userName := flag.String("username", "", "Username for Jabber")
  password := flag.String("password", "", "Password for Jabber")
  jabberChannel = flag.String("chat", "testbot", "MUC to join on start")
  flag.Parse()

  // Connect XMPP client
  log.Println("Creating client...")
  client, err := xmpp.NewClientNoTLS(jabberServer, *userName, *password, false)

  if err != nil {
    log.Fatalln("Failed to init client", err)
  }

  mucJID := fmt.Sprintf("%s@%s", *jabberChannel, jabberServer)
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
