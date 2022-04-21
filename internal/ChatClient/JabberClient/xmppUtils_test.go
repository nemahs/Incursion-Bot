package jabber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMUC(t *testing.T) {
  testServer := "test.com"
  testMUC := "testRoom@" + testServer
  testJID := testMUC + "/someUser"

  muc := parseMuc(testJID,testServer)

  assert.Equal(t,testMUC, muc)
}