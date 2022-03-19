package main

import (
	"fmt"
	"io"
	"time"

	"github.com/mattn/go-xmpp"
)

type JabberConnection struct {
	server   string
	channel  string
	username string
	password string
	client *xmpp.Client
}

// Create a new jabber connection
func CreateNewJabberConnection(server string, channel string, username string, password string) (JabberConnection, error) {
	newServer := JabberConnection {
		server: server,
		channel: channel,
		username: username,
		password: password,
	}

	err := newServer.ConnectToChannel()
	return newServer, err
}

// Connect to the configured server and the configured channel
func (conn *JabberConnection) ConnectToChannel() error {
	var err error
	Info.Printf("Connecting to %s...", conn.server)
	conn.client, err = xmpp.NewClientNoTLS(conn.server, conn.username, conn.password, false)

	if err != nil { return err }

  mucJID := fmt.Sprintf("%s@%s", *jabberChannel, jabberServer)
	Info.Printf("Joining %s as %s", mucJID, botNick)
	_, err = conn.client.JoinMUCNoHistory(mucJID, botNick)

	return err
}

// Tries to reconnect to the configured server in case of a disconnect
// TODO: Add exponential backoff?
func (comm *JabberConnection) reconnectLoop() {
	for ok := true; ok; {
		comm.client.Close()
		time.Sleep(time.Minute)
		err := comm.ConnectToChannel()

		ok = (err == nil)
	}
}

// Gets the next chat message, skipping over non-chat related messages (presence notifications, etc.)
func (comm *JabberConnection) GetNextChatMessage() (*xmpp.Chat, error) {
	for {
		msg, err := comm.client.Recv()

		if err != nil {
			if err == io.EOF {
				Warning.Println("Connection to server is broken, attempting to reconnect")
				comm.reconnectLoop()
				continue
			}

			return nil, err // Something weird happened, pass up to someone else to handle
		}

		chatMsg, ok := msg.(xmpp.Chat)
		if !ok || len(chatMsg.Text) == 0 { continue } // Not a valid chat message

		return &chatMsg, nil
	}
}

// Wraps sending jabber messages on the current client
func (conn *JabberConnection) Send(msg xmpp.Chat) error {
	_, err := conn.client.Send(msg)
	return err
}