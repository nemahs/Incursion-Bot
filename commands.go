package main

import (
	Chat "IncursionBot/internal/ChatClient"
	logging "IncursionBot/internal/Logging"
	"fmt"
	"strings"
	"time"
)

// Respond with the amount of time the bot's been up
func getUptime(msg Chat.ChatMsg) string {
	currentUptime := time.Since(startTime).Truncate(time.Second)
	msgText := fmt.Sprintf("Bot has been up for: %s", currentUptime)

	logging.Infof("Sending uptime in response to a message from %s", msg.Sender)
	return msgText
}

func printESIStatus(msg Chat.ChatMsg) string {
	var status string
	// if ESI.CheckESI() { status = "GOOD" } else { status = "BAD" }
	msgText := fmt.Sprintf("Connection to ESI is %s", status)
	logging.Infof("Sending ESI status in response to a message from %s", msg.Sender)
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
			incursion.TimeLeftString(timeFormat))
	}

	logging.Infof("Sending current incursions in response to a message from %s", msg.Sender)
	return responseText
}

func nextSpawn(msg Chat.ChatMsg) string {
	logging.Infof("Sending next spawn times in response to a message from %s", msg.Sender)
	return incManager.NextSpawns()
}

func waitlistInstructions(msg Chat.ChatMsg) string {
	logging.Infof("Sending waitlist instructions in response to a message from %s", msg.Sender)
	return `To join the waitlist, check that a fleet is actively running, then x up in the imperium.incursions channel in-game with the ships that you have.
Do not join the waitlist if you are not deployed to the HQ system. Do not move yourself.`
}

func printLayout(msg Chat.ChatMsg) string {
	incursions := incManager.GetIncursions()

	name := strings.Fields(msg.Text)[1]
	logging.Debugln(name)

	for _, incursion := range incursions {
		if incursion.Layout.StagingSystem.Name == name || incursion.Constellation.Name == name {
			resultText := "\n"

			resultText += "Staging: " + incursion.Layout.StagingSystem.Name + "\n"
			for _, vanguard := range incursion.Layout.VanguardSystems {
				resultText += "Vanguard: " + vanguard.Name + "\n"
			}

			for _, assault := range incursion.Layout.AssaultSystems {
				resultText += "Assault: " + assault.Name + "\n"
			}

			resultText += "HQ: " + incursion.Layout.HQSystem.Name + "\n"
			return resultText
		}
	}

	return "No spawn found"
}
