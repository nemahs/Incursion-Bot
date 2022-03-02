package main

import (
  "net/http/httptest"
  "net/http"
  "testing"
  "encoding/json"
  "time"
)

var testReturn []int = []int{3, 5}
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

func TestCachedCall (t *testing.T) {
  esi := ESI{}
  var result []int
  
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
  if result[0] != 3 { t.Errorf("Expected result to be %d, got %d", testReturn, result) }
  if cache.Data.([]int)[0] != 3 { t.Errorf("Expected cache to be updated, was instead %d", cache.Data) }
  if cache.Etag != testETag { t.Errorf("Expected Etag to be updated") }
  
  server.Config.Handler = http.HandlerFunc(return304)
  cache.Data = 5
  
  // Expired cache
  err = esi.cachedCall(testReq, &cache, &result)
  if err != nil { t.Errorf("Expected call to return successfully, got %s", err) }
  if result[0] != 3 { t.Errorf("Expected result to be %d, got %d", 3, result) }
  if cache.Data.([]int)[0] != 3 { t.Errorf("Expected cache to be updated, was instead %d", cache.Data) }  
  
}
