package main

import (
	Chat "IncursionBot/internal/ChatClient"
	jabber "IncursionBot/internal/ChatClient/JabberClient"
	logging "IncursionBot/internal/Logging"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const homeSystem int = 30004759 // 1DQ1-A
const commandPrefix byte = '!'  // All commands must start with this prefix
const timeFormat string = "Mon _2 Jan 15:04"

var commandsMap CommandMap      // Map of all supported commands, their functions, and their help messages
var logger logging.Logger       // Logger for the main application
var startTime time.Time         // Time the bot was started
var incManager IncursionManager // Manages known incursions and informs on state changes

// Returns goon home regions (currently Delve, Querious, and Period Basis)
func getHomeRegions() IDList {
	return IDList{10000060, 10000050, 10000063}
}

type IDList []int

func (list IDList) contains(val int) bool {
	for _, entry := range list {
		if entry == val {
			return true
		}
	}

	return false
}

func mainLoop() {
	incursionUpdateChan := make(chan IncursionList)
	firstRun := true
	go pollESI(incursionUpdateChan)

	for {
		newUpdates := <-incursionUpdateChan

		if firstRun {
			incManager.PopulateIncursions(newUpdates)
		} else {
			incManager.ProcessIncursions(newUpdates)
		}

		firstRun = false
	}
}

// Creates a notification message for a new incursion, creating a special message if the incursion is in a home region
func getNewIncursionMsg(newIncursion Incursion) string {
	if getHomeRegions().contains(newIncursion.Region.ID) {
		return fmt.Sprintf(":siren: New incursion detected in a home region! %s - %d jumps :siren:", newIncursion.ToString(), newIncursion.Distance)
	}

	return fmt.Sprintf("New incursion detected in %s - %d jumps", newIncursion.ToString(), newIncursion.Distance)
}

// Polls jabber and processes any commands received
func pollChat(jabber Chat.ChatServer) {
	for {
		msg, err := jabber.GetNextChatMessage()

		if err != nil {
			logger.Errorln("Error encountered receiving message: ", err)
			continue
		}

		if msg.Text[0] != commandPrefix {
			//Not a command, ignore
			continue
		}

		// Slice off the command prefix
		function, present := commandsMap.GetFunction(msg.Text[1:])
		if !present {
			logger.Warningf("Unknown or unsupported command: %s", msg.Text)
			continue
		}

		jabber.ReplyToMsg(function(msg), msg)
	}
}

// Parse a given file for a username and password. Expects the first line to be the username, and the second to be the password
func parseFile(fileName string) (*string, *string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file %s, error: %s", fileName, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	ok := scanner.Scan()
	if !ok {
		log.Fatalf("Failed to get username from file %s", fileName)
	}
	userName := scanner.Text()

	ok = scanner.Scan()
	if !ok {
		log.Fatalf("Failed to get password from file %s", fileName)
	}
	password := scanner.Text()

	return &userName, &password
}

func init() {
	startTime = time.Now()

	// Add commands to the command map
	commandsMap = NewCommandMap()
	commandsMap.AddCommand("incursions", listIncursions, "Lists the current incursions")
	commandsMap.AddCommand("uptime", getUptime, "Gets the current bot uptime")
	commandsMap.AddCommand("esi", printESIStatus, "Prints the bot's ESI connection status")
	commandsMap.AddCommand("nextspawn", nextSpawn, "Lists the start of the next spawn window for null and low incursions")
}

func main() {
	// Parse command line flags
	userName := flag.String("username", "", "Username for Jabber")
	password := flag.String("password", "", "Password for Jabber")
	userFile := flag.String("file", "", "File containing jabber username and password, line separated")
	debug := flag.Bool("debug", false, "Enables additional logging")

	jabberServer := flag.String("server", "conference.goonfleet.com", "Jabber server to connect to")
	jabberChannel := flag.String("chat", "testbot", "MUC to join on start")
	botNick := flag.String("nickname", "IncursionBot", "Name bot will connect to MUC with")
	flag.Parse()

	logger = logging.NewLogger(*debug)

	if *userFile != "" {
		userName, password = parseFile(*userFile)
	}

	if *userName == "" || *password == "" {
		log.Fatalln("One or more required parameters was missing")
	}

	client, err := jabber.CreateNewJabberConnection(*jabberServer, *jabberChannel, *userName, *password, *botNick)
	if err != nil {
		log.Fatalln("Failed initial connection to the server: ", err)
	}

	incManager = IncursionManager{
		onNewIncursion: func(i Incursion) {
			msgText := getNewIncursionMsg(i)
			logger.Infoln("Sending new incursion notification to chat")
			client.BroadcastToDefaultChannel(msgText)
		},
		onIncursionUpdate: func(i Incursion) {
			msgText := fmt.Sprintf("Incursion in %s changed state to %s", i.ToString(), i.State)
			logger.Infoln("Sending state change notification to chat")
			client.BroadcastToDefaultChannel(msgText)
		},
		onIncursionDespawn: func(i Incursion) {
			msgText := fmt.Sprintf("Incursion in %s despawned", i.ToString())
			logger.Infof("Sending despawn notification for %s", i.ToString())
			client.BroadcastToDefaultChannel(msgText)
		},
	}

	go pollChat(&client)
	mainLoop()
}
