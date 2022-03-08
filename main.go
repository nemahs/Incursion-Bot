package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-xmpp"
)

// TODO: Estimate time left in spawn

var (
  Info *log.Logger
  Warning *log.Logger
  Error *log.Logger
)

const maxRetries int = 10
const jabberServer string = "conference.goonfleet.com"  // Jabber server to connect to
const homeSystem int = 30004759                         // 1DQ1-A
const botNick string = "IncursionBot"                   // Bot will connect to jabber using this nickname
const commandPrefix byte = '!'                          // All commands must start with this prefix

var commandsMap CommandMap     // Map if all supported commands, their functions, and their help messages
var incursions IncursionList   // List of currently tracked incursions
var incursionsMutex sync.Mutex // Synchronize access to the incursionsList
var jabberChannel *string      // Jabber channel to broadcast to
var esi ESI                    // ESI client
var startTime time.Time        // Time the bot was started


// Returns goon home regions (currently Delve, Querious, and Period Basis)
func getHomeRegions() []int {
  return []int{10000060, 10000050, 10000063}
}

// Periodically polls ESI to get incursion data, and notifies chat of any changes
func pollIncursionsData(msgChan chan<- xmpp.Chat) {
  defer cleanup(msgChan)
  firstRun := true
  
  for {
    var newIncursionList IncursionList // List of incursions we've seen in this loop

    incursionResponses, nextPollTime, err := esi.getIncursions()

    if err != nil {
      Warning.Println("Error occurred getting incursions, sleeping 1 min then reattempting", err)
      time.Sleep(time.Minute)
      continue
    }
    
    for _, incursionData := range incursionResponses {
      existingIncursion := incursions.find(incursionData.StagingID)
      stagingInfo, err := esi.getSystemInfo(incursionData.StagingID)
      if err != nil {
        if existingIncursion != nil { 
          // Keep the previous incursion to not trigger a despawn
          newIncursionList = append(newIncursionList, *existingIncursion)
        }
        
        Error.Printf("Got error while parsing incursion system data for %d: %s", incursionData.StagingID, err)
        continue
      }

      if stagingInfo.SecurityClass == HighSec {
        continue // We do not give a fuck about highsec
      }

      if existingIncursion == nil {
        // No existing incursion found, make a new one
        newIncursion, err := CreateNewIncursion(incursionData)
        if err != nil {
          Error.Printf("Got error while creating an incursion: %s", err)
          continue // Skip this incursion, it's in a weird state
        }
        
        newIncursionList = append(newIncursionList, newIncursion)
        Info.Printf("Found new incursion in %s", newIncursion.ToString())

        // Don't want to spam chats with "NEW INCURSION" whenever the bot starts, so notifications are inhibited on the first run
        if !firstRun {
          var msgText string
          if contains(getHomeRegions(), newIncursion.Region.ID) {
            msgText = fmt.Sprintf(":siren: New incursion detected in a home region! %s - %d jumps :siren:", newIncursion.ToString(), newIncursion.Distance)
          } else {
            msgText = fmt.Sprintf("New incursion detected in %s - %d jumps", newIncursion.ToString(), newIncursion.Distance)
          }
          Info.Printf("Sending new incursion notification to %s", *jabberChannel)
          msgChan <- newGroupMessage(*jabberChannel, msgText)
        }
      } else {
        // Update data and check if anything changed
        Info.Printf("Found existing incursion in %s to update", existingIncursion.ToString())
        if UpdateIncursion(existingIncursion, incursionData) {
          msgText := fmt.Sprintf("Incursion in %s changed state to %s", existingIncursion.ToString(), existingIncursion.State)
          Info.Printf("Sending state change notification to %s", *jabberChannel)
          msgChan <- newGroupMessage(*jabberChannel, msgText)
        }

        newIncursionList = append(newIncursionList, *existingIncursion)
      }
    }

    // Check if any incursions have despawned and report
    for _, existing := range incursions {
      if newIncursionList.find(existing.StagingSystem.ID) == nil {
        msgText := fmt.Sprintf("Incursion in %s despawned", existing.ToString())
        Info.Printf("Sending despawn notification to %s for %s", *jabberChannel, existing.ToString())
        msgChan <- newGroupMessage(*jabberChannel, msgText)
      }
    }

    incursionsMutex.Lock()
    incursions = newIncursionList
    incursionsMutex.Unlock()
    
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
      Error.Println("Error encountered receiving message: ", err)
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
      Warning.Printf("Unknown or unsupported command: %s", chatMsg.Text)
      continue
    }

    msgChan <- function(chatMsg)
  }
}

// ------------- COMMANDS --------------------

// Respond with the amount of time the bot's been up
func getUptime(msg xmpp.Chat) xmpp.Chat {
  currentUptime := time.Since(startTime).Truncate(time.Second)
  msgText := fmt.Sprintf("Bot has been up for: %s", currentUptime)

  Info.Printf("Sending uptime in response to a message from %s", msg.Remote)
  return createReply(msg, msgText)
}

func printESIStatus(msg xmpp.Chat) xmpp.Chat {
  var status string
  if esi.CheckESI() { status = "GOOD" } else { status = "BAD" }
  msgText := fmt.Sprintf("Connection to ESI is %s", status)
  Info.Printf("Sending ESI status in response to a message from %s", msg.Remote)
  return createReply(msg, msgText)
}

func listIncursions(msg xmpp.Chat) xmpp.Chat {
  responseText := "\n"

  incursionsMutex.Lock()
  for _, incursion := range incursions {
    responseText += fmt.Sprintf("%s - Influence: %.2f%% - Status: %s - %d jumps \n",
    incursion.ToString(),
    incursion.Influence * 100, // Convert to % for easier reading
    incursion.State,
    incursion.Distance)
  }

  incursionsMutex.Unlock()
  Info.Printf("Sending current incursions in response to a message from %s", msg.Remote)
  return createReply(msg, responseText)
}

func parseFile(fileName string) (*string, *string) {
  f, err := os.Open(fileName)
  if err != nil {
    panic(err)
  }
  defer f.Close()

  scanner := bufio.NewScanner(f)

  scanner.Scan()
  userName := scanner.Text()

  scanner.Scan()
  password := scanner.Text()
  
  return &userName, &password

}


func init() {
  // Create loggers
  Info = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile|log.LUTC)
  Warning = log.New(os.Stdout, "WARN: ", log.LstdFlags|log.Lshortfile|log.LUTC)
  Error = log.New(os.Stdout, "ERROR: ", log.LstdFlags|log.Lshortfile|log.LUTC)
  
	startTime = time.Now()
	
  // Add commands to the command map
  commandsMap = NewCommandMap()
  commandsMap.AddCommand("incursions", listIncursions, "Lists the current incursions")
  commandsMap.AddCommand("uptime", getUptime, "Gets the current bot uptime")
  commandsMap.AddCommand("esi", printESIStatus, "Prints the bot's ESI connection status")
  
  esi = NewClient(esiURL)
}

func main() {
  // Parse command line flags
  userName := flag.String("username", "", "Username for Jabber")
  password := flag.String("password", "", "Password for Jabber")
  jabberChannel = flag.String("chat", "testbot", "MUC to join on start")
  userFile := flag.String("file", "", "File containing jabber username and password, line separated")
  flag.Parse()

  if *userFile != "" {
    userName, password = parseFile(*userFile)
  }

  // Connect XMPP client
  Info.Println("Creating client...")
  // goonfleet dot com promotes a connection to TLS later, the connection needs to start unencrypted
  // If the client attempts to initiate TLS, things break
  client, err := xmpp.NewClientNoTLS(jabberServer, *userName, *password, false)

  if err != nil {
    Error.Fatalln("Failed to init client", err)
  }

  mucJID := fmt.Sprintf("%s@%s", *jabberChannel, jabberServer)
  Info.Printf("Joining %s", mucJID)
  _, err = client.JoinMUCNoHistory(mucJID, botNick)

  if err != nil { Error.Println("Failed to join MUC", err) }

  // Spawn ESI and receive routines
  Info.Println("Client created, starting routines...")
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
          currentRetries--
          Warning.Printf("Restarting incursions routine after crash, %d tries remaining", currentRetries)
          go pollIncursionsData(esiChan)
        } else {
          _, err := client.Send(msg)

          if err != nil { log.Println(err) }
        }
      }
      
      case msg, ok := <-jabberChan: {
        if !ok {
          jabberChan = make(chan xmpp.Chat)
          currentRetries--
          Warning.Printf("Restarting jabber routine after crash, %d tries remaining", currentRetries)
          go pollChat(jabberChan, client)
        } else {
          _, err := client.Send(msg)

          if err != nil { log.Println(err) }
        }
      }
    }
  }

  Error.Fatalln("Max retries reached, shutting down...")
}
