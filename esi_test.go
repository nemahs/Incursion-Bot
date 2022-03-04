package main

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "reflect"
  "testing"
  "time"
)

var testReturn int = 3
const testETag = "abcde"

func return304 (rw http.ResponseWriter, req *http.Request) {
  rw.Header().Set("Expires", time.Now().Format(time.RFC1123))
  rw.WriteHeader(http.StatusNotModified)
}

func successfulReturn (rw http.ResponseWriter, req *http.Request) {
  result, _ := json.Marshal(testReturn)
  rw.Header().Set("Expires", time.Now().Format(time.RFC1123))
  rw.Header().Set("ETag", testETag)
  
  rw.Write(result)
}

func intCompare(t *testing.T, expected int, actual int) {
  if expected != actual { t.Errorf("Expected %d, got %d", expected, actual) }
}

func TestCachedCall (t *testing.T) {
  esi := ESI{}
  var result int
  cacheValue := 5
  
  server := httptest.NewServer(http.HandlerFunc(successfulReturn))
  
  testReq, _ := http.NewRequest("GET", server.URL, nil)
  
  err := esi.cachedCall(testReq, nil, &result)
  if err == nil { t.Errorf("Expected call to return an error for a nil cache") }
  var cache CacheEntry
  
  err = esi.cachedCall(nil, &cache, &result)
  if err == nil { t.Errorf("Expected call to return an error for a nil request") }
  
  // Normal path
  err = esi.cachedCall(testReq, &cache, &result)
  if err != nil { t.Errorf("Expected call to return successfully, got %s", err) }
  intCompare(t, testReturn, result)
  intCompare(t, testReturn, int(cache.Data.Int()))
  if cache.Etag != testETag { t.Errorf("Expected Etag to be updated") }
  
  cache.Data.Set(reflect.ValueOf(cacheValue))
  cache.ExpirationTime = time.Time{}
  if !cache.Expired() { t.Errorf("Cache should be expired, is not") }
  
  // Expired cache
  err = esi.cachedCall(testReq, &cache, &result)
  if err != nil { t.Errorf("Expected call to return successfully, got %s", err) }
  intCompare(t, testReturn, result)
  intCompare(t, testReturn, int(cache.Data.Int()))

  // Non-expired cache
  cache.Data.Set(reflect.ValueOf(cacheValue))
  cache.ExpirationTime = time.Now().Add(time.Minute)
  if cache.Expired() { t.Errorf("Cache should not be expired, but is") }

  err = esi.cachedCall(testReq, &cache, &result)
  if err != nil { t.Errorf("Expected call to return successfully, got %s", err) }
  intCompare(t, cacheValue, result)
  
  // NOT MODIFIED
  server.Config.Handler = http.HandlerFunc(return304)
  cache.ExpirationTime = time.Time{}
  cache.Data.Set(reflect.ValueOf(cacheValue))
  err = esi.cachedCall(testReq, &cache, &result)
  if err != nil { t.Errorf("Expected call to return successfully, got %s", err) }
  intCompare(t, cacheValue, result)
  if cache.ExpirationTime.IsZero() { t.Errorf("Expected cache expiration to be updated but it was not")}


}
