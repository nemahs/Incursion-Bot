package main

import (
	Chat "IncursionBot/internal/ChatClient"
	"fmt"
	"time"
)

// Respond with the amount of time the bot's been up
func getUptime(msg Chat.ChatMsg) string {
	currentUptime := time.Since(startTime).Truncate(time.Second)
	msgText := fmt.Sprintf("Bot has been up for: %s", currentUptime)

	logger.Infof("Sending uptime in response to a message from %s", msg.Sender)
	return msgText
}

func printESIStatus(msg Chat.ChatMsg) string {
	var status string
	// if ESI.CheckESI() { status = "GOOD" } else { status = "BAD" }
	msgText := fmt.Sprintf("Connection to ESI is %s", status)
	logger.Infof("Sending ESI status in response to a message from %s", msg.Sender)
	return msgText
}

func listIncursions(msg Chat.ChatMsg) string {
	responseText := "\n"
	incursions := incManager.GetIncursions()

	for _, incursion := range incursions {
		responseText += fmt.Sprintf("%s - Influence: %.2f%% - Status: %s - %d jumps, Despawn: %s \n",
			incursion.ToString(),
			incursion.Influence*100, // Convert to % for easier reading
			incursion.State,
			incursion.Distance,
			incursion.TimeLeftString())
	}

	logger.Infof("Sending current incursions in response to a message from %s", msg.Sender)
	return responseText
}

func nextSpawn(msg Chat.ChatMsg) string {
	logger.Infof("Sending next spawn times in response to a message from %s", msg.Sender)
	return incManager.NextSpawns()
}
