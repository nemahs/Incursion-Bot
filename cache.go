package main

import "time"

type CacheEntry struct {
	Data           interface{}
	ExpirationTime time.Time
	Etag string
}

func (entry *CacheEntry) Expired() bool {
	return time.Now().After(entry.ExpirationTime)
}


type CacheMap map[int] CacheEntry