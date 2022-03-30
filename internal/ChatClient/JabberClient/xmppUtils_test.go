package jabber

import (
	"log"
	"testing"

	"github.com/mattn/go-xmpp"
	"github.com/stretchr/testify/assert"
)

func TestNewGroupMessage(t *testing.T) {
  testMuc := "testChan"
  testText := "Test broadcast"

  result := newGroupMessage(testMuc, testText)

  assert.Equal(t, conferenceChat, result.Type)
  assert.Equal(t, testMuc + "@" + jabberServer, result.Remote)
  assert.Equal(t, testText, result.Text)
}


func TestCreateReply(t *testing.T) {
  testRoom := "testRoom@" + jabberServer
  testRequest := xmpp.Chat {
    Remote:  testRoom,
    Type:   "testType",
    Text:   "Hi",
  }

  replyText := "Reply"

  result := createReply(testRequest, replyText)

  assert.Equal(t, testRoom, result.Remote)
  assert.Equal(t, testRequest.Type, result.Type)
  assert.Equal(t, replyText, result.Text)

  // Change type to groupchat to make sure it removes the user from the remote
  testRequest.Type = conferenceChat
  testRequest.Remote = testRequest.Remote + "/subUser"
  log.Println(testRequest.Remote)

  result = createReply(testRequest, replyText)
  assert.Equal(t, testRoom, result.Remote)
}

func TestParseMUC(t *testing.T) {
  testServer := "test.com"
  testMUC := "testRoom@" + testServer
  testJID := testMUC + "/someUser"

  muc := parseMuc(testJID,testServer)

  assert.Equal(t,testMUC, muc)
}