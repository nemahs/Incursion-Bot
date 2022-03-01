package main

import (
  "log"
  "runtime/debug"
  "testing"

  "github.com/mattn/go-xmpp"
)

func assertStringEquals(t *testing.T, expected string, actual string) {
  if expected != actual {
    t.Errorf("Expected string to be %s, got %s", expected, actual)
    debug.PrintStack()
  }
}

func TestNewGroupMessage(t *testing.T) {
  testMuc := "testChan"
  testText := "Test broadcast"

  result := newGroupMessage(testMuc, testText)

  assertStringEquals(t, conferenceChat, result.Type)
  assertStringEquals(t, testMuc + "@" + jabberServer, result.Remote)
  assertStringEquals(t, testText, result.Text)
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

  assertStringEquals(t, testRoom, result.Remote)
  assertStringEquals(t, testRequest.Type, result.Type)
  assertStringEquals(t, replyText, result.Text)

  // Change type to groupchat to make sure it removes the user from the remote
  testRequest.Type = conferenceChat
  testRequest.Remote = testRequest.Remote + "/subUser"
  log.Println(testRequest.Remote)

  result = createReply(testRequest, replyText)
  assertStringEquals(t, testRoom, result.Remote)
}

func TestParseMUC(t *testing.T) {
  testServer := "test.com"
  testMUC := "testRoom@" + testServer
  testJID := testMUC + "/someUser"

  muc := parseMuc(testJID,testServer)

  assertStringEquals(t,testMUC, muc)
}