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

type IncursionResponse struct {
  Constellation_Id int
  Infested_Solar_Systems []int
  Influence float64
  Staging_Solar_System_Id int
  State string
}

var incursionData []IncursionResponse
var lastTag string
var currentExpirationTime time.Time


func getIncursions() ([]IncursionResponse, time.Time) {
  var result []IncursionResponse
  req, _ := http.NewRequest("GET", esiURL + "/incursions/", nil)
  req.Header.Add("If-None-Match", lastTag)
  
  resp, _ := http.DefaultClient.Do(req)
  // TODO: So much error checking missing
  
  if resp.StatusCode == http.StatusNotModified {
    return incursionData, currentExpirationTime
  }
  
  expirationTime, _ := time.Parse(time.RFC1123 , resp.Header.Get("Expires"))
  lastTag = resp.Header.Get("ETag")
	parseResults(resp, &result)
  
  return result, expirationTime
}


type NameResponse struct {
  Category string
  Id int
  Name string
}

func parseResults(resp *http.Response, resultStruct interface{}) error {
	parsedBody, err := ioutil.ReadAll(resp.Body)

	if err != nil { return err }

	err = json.Unmarshal(parsedBody, resultStruct)
	return err
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
		result[nameData.Id] = nameData.Name
	}

  return result
}

type ConstellationData struct {
    Name string
    Region_Id int
  }

func getConstInfo(constID int) ConstellationData {
  var response ConstellationData
  url := fmt.Sprintf("%s/universe/constellations/%d/", esiURL, constID)
  resp, _ := http.Get(url)

	if resp.StatusCode != http.StatusOK {
		log.Panicf("%+v", resp)
	}

  parseResults(resp, &response)

  return response
}

type SystemData struct {
	Id int
	Name string
	Security_Status float64
	Security_Class SecurityClass
}

func getSystemInfo(systemID int) SystemData {
	var results SystemData
	url := fmt.Sprintf("%s/universe/systems/%d/", esiURL, systemID)

	resp, _ := http.Get(url)
	parseResults(resp, &results)
	results.Security_Class = guessSecClass(results.Security_Status)

	return results
}

func GetRouteLength(startSystem int, endSystem int) int {
	var resultData []int
	url := fmt.Sprintf("%s/route/%d/%d/", esiURL, startSystem, endSystem)
	resp, _ := http.Get(url)

	parseResults(resp, &resultData)

	return len(resultData)
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