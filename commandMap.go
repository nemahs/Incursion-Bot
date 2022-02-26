package main

import (
	"fmt"

	"github.com/mattn/go-xmpp"
)

// Takes in a message from chat and returns the appropriate response message
type commandFunc func(xmpp.Chat) xmpp.Chat

// Map of supported commands and their functions
type CommandMap struct {
	funcMap map[string]commandFunc
	helpMap map[string]string
}

func NewCommandMap() CommandMap {
	newMap := CommandMap{
		funcMap: make(map[string]commandFunc),
		helpMap: make(map[string]string),
	}

	newMap.AddCommand("help", newMap.HelpText, "This help message")

	return newMap
}

// Default command to send all the supported commands in the map
func (m *CommandMap) HelpText(msg xmpp.Chat) xmpp.Chat {
	response := xmpp.Chat {
		Remote: parseMuc(msg.Remote, jabberServer),
		Type: msg.Type,
	}

	responseText := "Commands: \n"

	for command, help := range m.helpMap {
		responseText += fmt.Sprintf("%c%s  -  %s\n", commandPrefix, command, help)
	}

	response.Text = responseText

	return response
}

func (m *CommandMap) AddCommand(commandName string, function commandFunc, helpText string) {
	m.funcMap[commandName] = function
	m.helpMap[commandName] = helpText
}

func (m *CommandMap) GetFunction(commandName string) (commandFunc, bool) {
	function, pres := m.funcMap[commandName]
	return function, pres
}