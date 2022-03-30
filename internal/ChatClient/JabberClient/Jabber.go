package jabber

import (
	Chat "IncursionBot/internal/ChatClient"
	logging "IncursionBot/internal/Logging"
	"errors"
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
	nickname string
	client *xmpp.Client
}

var logger = logging.NewLogger()

const retryDuration = time.Minute // Time to wait between reconnect attempts

// Create a new jabber connection
func CreateNewJabberConnection(server string, channel string, username string, password string, nickname string) (JabberConnection, error) {
	newServer := JabberConnection {
		server: server,
		channel: channel,
		username: username,
		password: password,
		nickname: nickname,
	}

	err := newServer.ConnectToChannel()
	return newServer, err
}

// Connect to the configured server and the configured channel
func (conn *JabberConnection) ConnectToChannel() error {
	var err error
	logger.Infof("Connecting to %s...", conn.server)
	// The connection breaks if you try to initiate connected, better to let the server promote the connection to TLS
	conn.client, err = xmpp.NewClientNoTLS(conn.server, conn.username, conn.password, false)

	if err != nil { return err }
	if !conn.client.IsEncrypted() { return errors.New("Server did not promote connection to TLS") }

  mucJID := fmt.Sprintf("%s@%s", conn.channel, conn.server)
	logger.Infof("Joining %s as %s", mucJID, conn.nickname)
	_, err = conn.client.JoinMUCNoHistory(mucJID, conn.nickname)

	return err
}

// Tries to reconnect to the configured server in case of a disconnect
// TODO: Add exponential backoff?
func (comm *JabberConnection) reconnectLoop() {
	for ok := true; ok; {
		comm.client.Close()
		time.Sleep(retryDuration)
		err := comm.ConnectToChannel()

		ok = (err == nil)
	}
}

// Gets the next chat message, skipping over non-chat related messages (presence notifications, etc.)
func (comm *JabberConnection) GetNextChatMessage() (Chat.ChatMsg, error) {
	for {
		msg, err := comm.client.Recv()

		if err != nil {
			if err == io.EOF {
				logger.Warningln("Connection to server is broken, attempting to reconnect")
				comm.reconnectLoop()
				continue
			}

			return Chat.ChatMsg{}, err // Something weird happened, pass up to someone else to handle
		}

		chatMsg, ok := msg.(xmpp.Chat)
		if !ok || len(chatMsg.Text) == 0 { continue } // Not a valid chat message

		return Chat.ChatMsg{
			Sender: chatMsg.Remote,
			Type: parseMsgType(chatMsg),
			Text: chatMsg.Text,
		}, nil
	}
}

func (conn *JabberConnection) ReplyToMsg(message string, origMsg Chat.ChatMsg) error {
	msg := conn.createReply(origMsg, message)

	_, err := conn.client.Send(msg)
	return err
}

// TODO: Add more than one channel it can be in
func (conn *JabberConnection) BroadcastToChannel(message string, channel string) error {
	msg := conn.newGroupMessage(message, channel)

	_, err := conn.client.Send(msg)
	return err
}

func (conn *JabberConnection) BroadcastToDefaultChannel(message string) error {
	msg := conn.newGroupMessage(conn.channel, message)

	_, err := conn.client.Send(msg)
	return err
}

func (conn *JabberConnection) SendToUser(message string, user string) error {
	msg := xmpp.Chat{
		Type: privateMessage,
		Remote: user,
		Text: message,
	}

	_, err := conn.client.Send(msg)
	return err
}

func parseMsgType(msg xmpp.Chat) Chat.MessageType {
	switch msg.Type {
	case privateMessage:
		return Chat.PrivateMessage
	case conferenceChat:
		return Chat.ChannelMessage
	}

	return Chat.Unknown
}

// Create a message to go to a conference room
func (conn *JabberConnection) newGroupMessage(muc string, text string) xmpp.Chat {
  return xmpp.Chat{
    Remote: fmt.Sprintf("%s@%s", muc, conn.server),
    Type: conferenceChat,
    Text:   text,
  }
}

// Create a reply back to the chat that requested it
func (conn *JabberConnection) createReply(origMsg Chat.ChatMsg, response string) xmpp.Chat {
  result := xmpp.Chat{
    Text: response,
  }

  if origMsg.Type == Chat.ChannelMessage {
    // Strip username off the end for remote
    result.Remote = parseMuc(origMsg.Sender, conn.server)
		result.Type = conferenceChat
  } else {
    result.Remote = origMsg.Sender
		result.Type = privateMessage
  }

  return result
}
