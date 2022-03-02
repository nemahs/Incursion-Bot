package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "reflect"
  "time"
)


const esiURL string = "https://esi.evetech.net/latest"

type ESI struct {
  baseURL string
}

func NewClient(baseURL string) ESI {
  return ESI{baseURL: baseURL}
}

// Util functions

// Parse JSON results from HTTP response into a given struct
func (c *ESI) parseResults(resp *http.Response, resultStruct interface{}) error {
  if resp == nil { return fmt.Errorf("resp was nil") }

  parsedBody, err := ioutil.ReadAll(resp.Body)
  if err != nil { return err }

  err = json.Unmarshal(parsedBody, resultStruct)
  return err
}

func parseExpirationTime(resp *http.Response) (time.Time, error) {
    return time.Parse(time.RFC1123 , resp.Header.Get("Expires"))
}

func (c *ESI) cachedCall(req *http.Request, cache *CacheEntry, resultStruct interface{}) error {
  if req == nil || cache == nil { 
    return fmt.Errorf("one of the inputs was null")
  }
  
  if !cache.Expired() {
    reflect.ValueOf(resultStruct).Elem().Set(cache.Data)
    return nil
  }

  req.Header.Add("If-None-Match", cache.Etag)

  resp, err := http.DefaultClient.Do(req)
  if err != nil { return err }

  switch resp.StatusCode {
  case http.StatusOK: // Expected case
    err = c.parseResults(resp, resultStruct)
    if err != nil { return err }
    cache.Data = reflect.ValueOf(resultStruct).Elem()
    cache.ExpirationTime, err = parseExpirationTime(resp)
    cache.Etag = resp.Header.Get("ETag")
    return err
  case http.StatusNotModified:
    reflect.ValueOf(resultStruct).Elem().Set(cache.Data)
    cache.ExpirationTime, err = parseExpirationTime(resp)
    return err
  case http.StatusServiceUnavailable, http.StatusInternalServerError:
    log.Println("ESI is having problems, returning cached data instead")
    return nil
  default: 
    data, _ := ioutil.ReadAll(resp.Body)
    return fmt.Errorf("status code %d received from server: %s", resp.StatusCode, string(data))
  }
}

// Incursion functions

type IncursionResponse struct {
  ConstellationID   int            `json:"constellation_id"`
  IncursionSystems  []int          `json:"infested_solar_systems"`
  Influence         float64        `json:"influence"`
  StagingID         int            `json:"staging_solar_system_id"`
  State             IncursionState `json:"state"`
}

var incursionsCache CacheEntry
func (c *ESI) getIncursions() ([]IncursionResponse, time.Time, error) {
  var result []IncursionResponse
  req, err := http.NewRequest("GET", c.baseURL + "/incursions/", nil)
  if err != nil {
    Error.Println("Failed to create request for incursions", err)
    return result, incursionsCache.ExpirationTime, err
  }
  err = c.cachedCall(req, &incursionsCache, &result)
  
  if err != nil {
    Error.Println("Error occured while getting incursions", err)
    return result, incursionsCache.ExpirationTime, err
  }

  return result, incursionsCache.ExpirationTime, nil
}

// --------- NAME RESOLUTION ---------

type NameResponse struct {
  Category  string
  ID        int
  Name      string
}
type NameMap map[int]string // Map of item IDs to names

var cachedNames NameMap = make(NameMap)
func (c *ESI) getNames(ids []int) (NameMap, error) {
  var responseData []NameResponse
  result := make(NameMap)

  // Filter out names that we already know
  var unknownIDs []int
  for _, id := range ids {
    cacheEntry, pres := cachedNames[id]

    if !pres {
      unknownIDs = append(unknownIDs, id)
    } else {
      result[id] = cacheEntry
    }
  }

  if len(unknownIDs) == 0 {
    // We already know all the IDs, no need to bother ESI
    return result, nil
  }

  // Find the remaining names
  data, err := json.Marshal(unknownIDs)
  if err != nil {
    Error.Println("Failed to marshal IDs into json", err)
    return result, err
  }

  req, err := http.NewRequest("POST", c.baseURL + "/universe/names/", bytes.NewBuffer(data))
  if err != nil {
    Error.Println("Failed to create name request", req)
    return result, err
  }

  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    Error.Println("Failed HTTP request for names", err)
    return result, err
  }

  if resp.StatusCode != http.StatusOK {
    body, _ := ioutil.ReadAll(resp.Body)
    Error.Printf("Name endpoint returned a status code of %d: %s", resp.StatusCode, string(body))
    return result, err
  }

  err = c.parseResults(resp, &responseData)
  if err != nil {
    Error.Println("Failed to parse name results", err)
    return result, err
  }


  // Return result
  for _, nameData := range responseData {
    cachedNames[nameData.ID] = nameData.Name
    result[nameData.ID] = nameData.Name
  }

  return result, nil
}

// ------- CONSTELLATION INFO --------

type ConstellationData struct {
  ID        int `json:"constellation_id"`
  Name      string
  RegionID  int `json:"region_id"`
}

var constDataCache CacheMap = make(CacheMap)
func (c *ESI) getConstInfo(constID int) (ConstellationData, error) {
  var response ConstellationData
  url := fmt.Sprintf("%s/universe/constellations/%d/", c.baseURL, constID)
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    Error.Printf("Failed to create constellation info request for id: %d", constID)
    return response, err
  }

  cacheData := constDataCache[constID]
  err = c.cachedCall(req, &cacheData, &response)
  if err != nil {
    Error.Println("Error occurred in getting the constellation data", err)
    return response, err
  }

  return response, nil
}

// ----------- SYSTEM INFO -----------

type SystemData struct {
  ID            int `json:"system_id"`
  Name          string
  SecStatus     float64 `json:"security_status"`
  SecurityClass SecurityClass
}

var systemCache CacheMap = make(CacheMap)
func (c *ESI) getSystemInfo(systemID int) (SystemData, error) {
  var results SystemData
  url := fmt.Sprintf("%s/universe/systems/%d/", c.baseURL, systemID)
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    Error.Println("An error occurred creating the system info request", err)
    return results, err
  }

  cacheData := systemCache[systemID]
  err = c.cachedCall(req, &cacheData, &results)
  if err != nil {
    Error.Println("An error occurred getting system info", err)
    return results, err
  }
  results.SecurityClass = guessSecClass(results.SecStatus)

  return results, nil
}


// ----- ROUTE -----

// TODO: Cache this endpoint
type Route []int

func (c *ESI) GetRouteLength(startSystem int, endSystem int) (int, error) {
  var resultData Route
  url := fmt.Sprintf("%s/route/%d/%d/", c.baseURL, startSystem, endSystem)
  resp, err := http.Get(url)
  if err != nil {
    Error.Println("Failed HTTP request for route length", err)
    return -1, err
  }

  err = c.parseResults(resp, &resultData)
  if err != nil {
    Error.Println("Error occurred parsing results", err)
    return -1, err
  }

  return len(resultData) - 2, nil // Subtract off the start and end systems
}

type SecurityClass string

const (
  HighSec SecurityClass = "High"
  LowSec  SecurityClass = "Low"
  NullSec SecurityClass = "Null"
)

func guessSecClass(status float64) SecurityClass {
  if status > .5 {
    return HighSec
  } else if (status > .1) {
    return LowSec
  }
  return NullSec
}

func (c *ESI) CheckESI() bool {
  // TODO: Mess with this so it uses swagger to verify the integrety of each endpoint
  url := fmt.Sprintf("%s/swagger.json", c.baseURL)
  resp, err := http.Get(url)

  if err != nil {
    Error.Println("Error occurred querying ESI:", err)
    return false
  }

  return resp.StatusCode == http.StatusOK
}
