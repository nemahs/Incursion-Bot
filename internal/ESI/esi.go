package ESI

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)


const esiURL string = "https://esi.evetech.net/"
var errorLog log.Logger

type ESIClient struct {
  baseURL string
}

func NewClient() ESIClient {
  return ESIClient{baseURL: esiURL + "latest" }
}

func NewClientWithVersion(version string) ESIClient {
  return ESIClient{baseURL: esiURL + version }
}

func init() {
  errorLog = *log.New(os.Stderr, "ERR:", log.LUTC|log.LstdFlags)
}


// Parse JSON results from HTTP response into a given struct
func (c *ESIClient) parseResults(resp *http.Response, resultStruct interface{}) error {
  if resp == nil { return fmt.Errorf("resp was nil") }

  parsedBody, err := ioutil.ReadAll(resp.Body)
  if err != nil { return err }

  err = json.Unmarshal(parsedBody, resultStruct)
  return err
}

func parseExpirationTime(resp *http.Response) (time.Time, error) {
    return time.Parse(time.RFC1123 , resp.Header.Get("Expires"))
}

func (c *ESIClient) cachedCall(req *http.Request, cache *CacheEntry, resultStruct interface{}) error {
  if req == nil || cache == nil { 
    return fmt.Errorf("one of the inputs was null")
  }

  result := reflect.ValueOf(resultStruct)

  
  if !cache.Expired() {
    result.Elem().Set(cache.Data)
    return nil
  }

  req.Header.Add("If-None-Match", cache.Etag)

  resp, err := http.DefaultClient.Do(req)
  if err != nil { return err }

  switch resp.StatusCode {
  case http.StatusOK: // Expected case
    err = c.parseResults(resp, resultStruct)
    if err != nil { return err }
    cache.Data = result.Elem()
    cache.ExpirationTime, err = parseExpirationTime(resp)
    cache.Etag = resp.Header.Get("ETag")
    return err
  case http.StatusNotModified:
    if !cache.Data.IsValid() {
      return fmt.Errorf("cache was empty")
    }

    result.Elem().Set(cache.Data)
    cache.ExpirationTime, err = parseExpirationTime(resp)
    return err
  case http.StatusServiceUnavailable, http.StatusInternalServerError, http.StatusGatewayTimeout, http.StatusBadGateway:
    log.Println("ESI is having problems, returning cached data instead")
    log.Printf("Got status code: %d", resp.StatusCode)
    if !cache.Data.IsValid() {
      return fmt.Errorf("cache was empty")
    }

    result.Elem().Set(cache.Data)
    return nil
  default: 
    data, _ := ioutil.ReadAll(resp.Body)
    return fmt.Errorf("status code %d received from server: %s", resp.StatusCode, string(data))
  }
}

func (c *ESIClient) CheckESI() bool {
  // TODO: Mess with this so it uses swagger to verify the integrety of each endpoint
  url := fmt.Sprintf("%s/swagger.json", c.baseURL)
  resp, err := http.Get(url)

  if err != nil {
    errorLog.Println("Error occurred querying ESI:", err)
    return false
  }

  return resp.StatusCode == http.StatusOK
}
