package ESI

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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


func TestCachedCall (t *testing.T) {
  assert := assert.New(t)
  esi := ESIClient{}
  var result int
  var cache CacheEntry
  var cacheValue int = 5
  
  server := httptest.NewServer(http.HandlerFunc(successfulReturn))
  
  testReq, _ := http.NewRequest("GET", server.URL, nil)
  
  err := esi.cachedCall(testReq, nil, &result)
  assert.Error(err)
  
  err = esi.cachedCall(nil, &cache, &result)
  assert.Error(err)

  // Normal path
  err = esi.cachedCall(testReq, &cache, &result)
  assert.NoError(err)
  assert.Equal(testReturn, result)
  assert.Equal(testReturn, int(cache.Data.Int()))
  assert.Equal(cache.Etag, testETag)
  
  cache.Data.Set(reflect.ValueOf(cacheValue))
  cache.ExpirationTime = time.Time{}
  assert.True(cache.Expired())

  t.Run("Caching", func(t *testing.T) {
    // Expired cache
    err := esi.cachedCall(testReq, &cache, &result)
    assert.NoError(err)
    assert.Equal(testReturn, result)
    assert.Equal(testReturn, int(cache.Data.Int()))

    // Non-expired cache
    cache.Data.Set(reflect.ValueOf(cacheValue))
    cache.ExpirationTime = time.Now().Add(time.Minute)
    assert.False(cache.Expired())

    err = esi.cachedCall(testReq, &cache, &result)
    assert.NoError(err)
    assert.Equal(cacheValue, result)
  })
  
  
  t.Run("Error returns", func(t *testing.T) {
    // NOT MODIFIED
    server.Config.Handler = http.HandlerFunc(return304)
    cache.ExpirationTime = time.Time{}
    cache.Data.Set(reflect.ValueOf(cacheValue))
    err = esi.cachedCall(testReq, &cache, &result)
    assert.NoError(err)
    assert.Equal(cacheValue, result)
    assert.False(cache.ExpirationTime.IsZero())

    // Zero Cache with Server error
    server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.WriteHeader(http.StatusInternalServerError)  
    })
    cache.Data = reflect.Value{}
    err = esi.cachedCall(testReq, &cache, &result)
    assert.Error(err)
  })
}
