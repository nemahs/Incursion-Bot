package main

import "github.com/mattn/go-xmpp"

type IDList []int

// Check if a slice contains the speicifed value
func (list IDList) contains(val int) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

// Recovers from any uncaught panics so that the main thread
// can restart the routine.
func cleanup(channel chan<- xmpp.Chat) {
	err := recover()
	if err != nil {
		Warning.Println("Recovered from unexpected error: ", err)
	}

	close(channel)
}