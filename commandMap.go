package main

import (
	"fmt"
	"regexp"

	"github.com/mattn/go-xmpp"
)

type commandFunc func(xmpp.Chat) xmpp.Chat

type CommandMap struct {
	funcMap map[string]commandFunc
	helpMap map[string]string
}

func NewCommandMap() CommandMap {
	newMap := CommandMap{
		funcMap: make(map[string]commandFunc),
		helpMap: make(map[string]string),
	}

	newMap.AddCommand("!help", newMap.HelpText, "This help message")

	return newMap
}

// Parse the JID to extract the MUC it came from
func parseMuc(jid string, server string) string {
	mucReg, _ := regexp.Compile(".*@" + server)
	muc := string(mucReg.Find([]byte(jid)))

	return muc
}

func (m *CommandMap) HelpText(msg xmpp.Chat) xmpp.Chat {
	response := xmpp.Chat {
		Remote: parseMuc(msg.Remote, jabberServer),
		Type: msg.Type,
	}

	responseText := "Commands: \n"

	for command, help := range m.helpMap {
		responseText += fmt.Sprintf("%s  -  %s\n", command, help)
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