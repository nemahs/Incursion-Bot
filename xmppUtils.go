package main

import (
  "fmt"
  "regexp"

  "github.com/mattn/go-xmpp"
)

const (
  conferenceChat string = "groupchat"
  privateMessage string = "chat"
)

// Create a message to go to a conference room
func newGroupMessage(muc string, text string) xmpp.Chat {
  return xmpp.Chat{
    Remote: fmt.Sprintf("%s@%s", muc, jabberServer),
    Type: conferenceChat,
    Text:   text,
  }
}

// Create a reply back to the chat that requested it
func createReply(request xmpp.Chat, response string) xmpp.Chat {
  result := xmpp.Chat{
    Type: request.Type,
    Text: response,
  }

  if request.Type == conferenceChat {
    // Strip username off the end for remote
    result.Remote = parseMuc(request.Remote, jabberServer)
  } else {
    result.Remote = request.Remote
  }

  return result
}

// Parse the JID to extract the MUC it came from
func parseMuc(jid string, server string) string {
  mucReg, _ := regexp.Compile(".*@" + server)
  muc := string(mucReg.Find([]byte(jid)))

  return muc
}