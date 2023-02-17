package main

import (
	Chat "IncursionBot/internal/ChatClient"
	"fmt"
)

// Takes in a message from chat and returns the appropriate response message
type commandFunc func(Chat.ChatMsg) string

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
func (m *CommandMap) HelpText(msg Chat.ChatMsg) string {
	responseText := "Commands: \n"

	for command, help := range m.helpMap {
		responseText += fmt.Sprintf("%c%s  -  %s\n", commandPrefix, command, help)
	}

	return responseText
}

func (m *CommandMap) AddCommand(commandName string, function commandFunc, helpText string) {
	m.funcMap[commandName] = function
	m.helpMap[commandName] = helpText
}

func (m *CommandMap) GetFunction(commandName string) (commandFunc, bool) {
	function, pres := m.funcMap[commandName]
	return function, pres
}
