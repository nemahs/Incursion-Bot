package jabber

import (
	"regexp"
)

const (
	conferenceChat string = "groupchat"
	privateMessage string = "chat"
)

// Parse the JID to extract the MUC it came from
func parseMuc(jid string, server string) string {
	mucReg, _ := regexp.Compile(".*@" + server)
	muc := string(mucReg.Find([]byte(jid)))

	return muc
}
