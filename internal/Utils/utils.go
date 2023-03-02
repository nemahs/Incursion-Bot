package Utils

import (
	"log"

	"github.com/mattn/go-xmpp"
)

type IDList []int

// Check if a slice contains the speicifed value
func (list IDList) Contains(val int) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

// Recovers from any uncaught panics so that the main thread
// can restart the routine.
func Cleanup(channel chan<- xmpp.Chat, logger *log.Logger) {
	err := recover()
	if err != nil {
		logger.Println("Recovered from unexpected error: ", err)
	}

	close(channel)
}
