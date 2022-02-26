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

func parseResults(resp *http.Response, resultStruct interface{}) error {
	parsedBody, err := ioutil.ReadAll(resp.Body)

	if err != nil { return err }

	err = json.Unmarshal(parsedBody, resultStruct)
	return err
}

func cachedCall(req *http.Request, cache *CacheEntry, resultStruct interface{}) error {
	if !cache.Expired() {
		resultStruct = cache.Data
		return nil
	}

	req.Header.Add("If-None-Match", cache.Etag)

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
  cache.ExpirationTime, err = time.Parse(time.RFC1123 , resp.Header.Get("Expires"))
	if err != nil { return err }

	if resp.StatusCode == http.StatusNotModified {
		resultStruct = cache.Data
		return nil
	}

	err = parseResults(resp, resultStruct)
	if err != nil { return err }
	cache.Data = resultStruct

	return nil
}

// Incursion functions

type IncursionResponse struct {
  ConstellationID 	int `json:"constellation_id"`
  IncursionSystems	[]int `json:"infested_solar_systems"`
  Influence					float64
  StagingID					int `json:"staging_solar_system_id"`
  State							IncursionState
}

var incursionsCache CacheEntry
func getIncursions() ([]IncursionResponse, time.Time) {
	var result []IncursionResponse
  req, _ := http.NewRequest("GET", esiURL + "/incursions/", nil)
	err := cachedCall(req, &incursionsCache, &result)
  
	if err != nil {
		log.Println("Error occured while getting incursions", err)
	}

  return result, incursionsCache.ExpirationTime
}

type NameResponse struct {
  Category	string
  ID				int
  Name			string
}

// TODO: CACHE ALL THIS SHIT
func getNames(ids []int) map[int]string {
  var responseData []NameResponse
	result := make(map[int]string)

	data, _ := json.Marshal(ids)
  req, _ := http.NewRequest("POST", esiURL + "/universe/names/", bytes.NewBuffer(data))
  resp, _ := http.DefaultClient.Do(req)
	parseResults(resp, &responseData)

	for _, nameData := range responseData {
		result[nameData.ID] = nameData.Name
	}

  return result
}

type ConstellationData struct {
	Name			string
	RegionID	int `json:"region_id"`
}

var constDataCache CacheMap = make(CacheMap)
func getConstInfo(constID int) ConstellationData {
  var response ConstellationData
  url := fmt.Sprintf("%s/universe/constellations/%d/", esiURL, constID)
	req, _ := http.NewRequest("GET", url, nil)

	cacheData := constDataCache[constID]
	err := cachedCall(req, &cacheData, &response)

	if err != nil {
		log.Println("Error occurred in getting the constellation data", err)
	}
  return response
}

type SystemData struct {
	ID 						int
	Name 					string
	SecStatus 		float64 `json:"security_status"`
	SecurityClass SecurityClass
}

var systemCache CacheMap = make(CacheMap)
func getSystemInfo(systemID int) SystemData {
	var results SystemData
	url := fmt.Sprintf("%s/universe/systems/%d/", esiURL, systemID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("An error occurred creating the system info request", err)
		return results
	}

	cacheData := systemCache[systemID]
	err = cachedCall(req, &cacheData, results)
	if err != nil {
		log.Println("An error occurred getting system info", err)
		return results
	}
	results.SecurityClass = guessSecClass(results.SecStatus)

	return results
}

func GetRouteLength(startSystem int, endSystem int) int {
	var resultData []int
	url := fmt.Sprintf("%s/route/%d/%d/", esiURL, startSystem, endSystem)
	resp, _ := http.Get(url)

	parseResults(resp, &resultData)

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

func checkESI() {
	// TODO: Mess with this so it uses swagger to verify the integrety of each endpoint
	url := "https://esi.evetech.net/latest/swagger.json"
	resp, _ := http.Get(url)

	parsedData, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}

	json.Unmarshal(parsedData, &result)

	for key, _ := range result["paths"].(map[string]interface{}) {
		log.Println(key)
	}
}