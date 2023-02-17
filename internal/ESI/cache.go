package ESI

import (
	"reflect"
	"time"
)

type CacheEntry struct {
	Data           reflect.Value
	ExpirationTime time.Time
	Etag           string
}

func (entry *CacheEntry) Expired() bool {
	return time.Now().After(entry.ExpirationTime)
}

type CacheMap map[int]CacheEntry
