package main

import (
	"fmt"
	"time"

	"github.com/mattn/go-xmpp"
)

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
  incursions := testIncursions.Get()

  for _, incursion := range incursions {
    responseText += fmt.Sprintf("%s - Influence: %.2f%% - Status: %s - %d jumps, Despawn: %s \n",
    incursion.ToString(),
    incursion.Influence * 100, // Convert to % for easier reading
    incursion.State,
    incursion.Distance,
    incursion.TimeLeftString())
  }

  Info.Printf("Sending current incursions in response to a message from %s", msg.Remote)
  return createReply(msg, responseText)
}

func nextSpawn(msg xmpp.Chat) xmpp.Chat {
  reponseText := ""

  // TODO: Need to track incursions that have despawned and when they despawned

  return createReply(msg, reponseText)
}