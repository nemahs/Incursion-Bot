package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)


const esiURL string = "https://esi.evetech.net/latest"

// Util functions

// Parse JSON results from HTTP response into a given struct
func parseResults(resp *http.Response, resultStruct interface{}) error {
  parsedBody, err := ioutil.ReadAll(resp.Body)

  if err != nil { return err }

  err = json.Unmarshal(parsedBody, resultStruct)
  return err
}

func cachedCall(req *http.Request, cache *CacheEntry, resultStruct interface{}) error {
  if !cache.Expired() {
    resultStruct = cache.Data //lint:ignore SA4006 resultStruct is an output interface
    return nil
  }

  req.Header.Add("If-None-Match", cache.Etag)

  resp, err := http.DefaultClient.Do(req)
  if err != nil { return err }
  cache.ExpirationTime, err = time.Parse(time.RFC1123 , resp.Header.Get("Expires"))
  if err != nil { return err }

  switch resp.StatusCode {
  case http.StatusNotModified: {
    resultStruct = cache.Data //lint:ignore SA4006 resultStruct is an output interface

    return nil
  }
  case http.StatusOK: break // Expected case
  case http.StatusServiceUnavailable:
    fallthrough
  case http.StatusInternalServerError:
    log.Println("ESI is having problems, returning cached data instead")
    resultStruct = cache.Data //lint:ignore SA4006 resultStruct is an output interface

    return nil
  default: 
    data, _ := ioutil.ReadAll(resp.Body)
    return fmt.Errorf("status code %d received from server: %s", resp.StatusCode, string(data))
  }

  err = parseResults(resp, resultStruct)
  if err != nil { return err }
  cache.Data = resultStruct

  return nil
}

// Incursion functions

type IncursionResponse struct {
  ConstellationID   int `json:"constellation_id"`
  IncursionSystems  []int `json:"infested_solar_systems"`
  Influence         float64
  StagingID         int `json:"staging_solar_system_id"`
  State             IncursionState
}

var incursionsCache CacheEntry
func getIncursions() ([]IncursionResponse, time.Time, error) {
  var result []IncursionResponse
  req, err := http.NewRequest("GET", esiURL + "/incursions/", nil)
  if err != nil {
    Error.Println("Failed to create request for incursions", err)
    return result, incursionsCache.ExpirationTime, err
  }
  err = cachedCall(req, &incursionsCache, &result)
  
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
func getNames(ids []int) NameMap {
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
    return result
  }

  // Find the remaining names
  data, err := json.Marshal(unknownIDs)
  if err != nil {
    Error.Println("Failed to marshal IDs into json", err)
    return result
  }

  req, err := http.NewRequest("POST", esiURL + "/universe/names/", bytes.NewBuffer(data))
  if err != nil {
    Error.Println("Failed to create name request", req)
    return result
  }

  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    Error.Println("Failed HTTP request for names", err)
    return result
  }

  if resp.StatusCode != http.StatusOK {
    body, _ := ioutil.ReadAll(resp.Body)
    Error.Printf("Name endpoint returned a status code of %d: %s", resp.StatusCode, string(body))
    return result
  }

  err = parseResults(resp, &responseData)
  if err != nil {
    Error.Println("Failed to parse name results", err)
    return result
  }


  // Return result
  for _, nameData := range responseData {
    cachedNames[nameData.ID] = nameData.Name
    result[nameData.ID] = nameData.Name
  }

  return result
}

// ------- CONSTELLATION INFO --------

type ConstellationData struct {
  ID        int `json:"constellation_id"`
  Name      string
  RegionID  int `json:"region_id"`
}

var constDataCache CacheMap = make(CacheMap)
func getConstInfo(constID int) ConstellationData {
  var response ConstellationData
  url := fmt.Sprintf("%s/universe/constellations/%d/", esiURL, constID)
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    Error.Printf("Failed to create constellation info request for id: %d", constID)
    return response
  }

  cacheData := constDataCache[constID]
  err = cachedCall(req, &cacheData, &response)
  if err != nil {
    Error.Println("Error occurred in getting the constellation data", err)
  }

  return response
}

// ----------- SYSTEM INFO -----------

type SystemData struct {
  ID            int `json:"system_id"`
  Name          string
  SecStatus     float64 `json:"security_status"`
  SecurityClass SecurityClass
}

var systemCache CacheMap = make(CacheMap)
func getSystemInfo(systemID int) SystemData {
  var results SystemData
  url := fmt.Sprintf("%s/universe/systems/%d/", esiURL, systemID)
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    Error.Println("An error occurred creating the system info request", err)
    return results
  }

  cacheData := systemCache[systemID]
  err = cachedCall(req, &cacheData, &results)
  if err != nil {
    Error.Println("An error occurred getting system info", err)
    return results
  }
  results.SecurityClass = guessSecClass(results.SecStatus)

  return results
}


// ----- ROUTE -----

// TODO: Cache this endpoint
type Route []int

func GetRouteLength(startSystem int, endSystem int) int {
  var resultData Route
  url := fmt.Sprintf("%s/route/%d/%d/", esiURL, startSystem, endSystem)
  resp, err := http.Get(url)
  if err != nil {
    Error.Println("Failed HTTP request for route length", err)
    return -1
  }

  err = parseResults(resp, &resultData)
  if err != nil {
    Error.Println("Error occurred parsing results", err)
    return -1
  }

  return len(resultData) - 2 // Subtract off the start and end systems
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

func checkESI() bool {
  // TODO: Mess with this so it uses swagger to verify the integrety of each endpoint
  url := "https://esi.evetech.net/latest/swagger.json"
  resp, err := http.Get(url)

  if err != nil {
    Error.Println("Error occurred querying ESI:", err)
    return false
  }

  return resp.StatusCode == http.StatusOK
}
