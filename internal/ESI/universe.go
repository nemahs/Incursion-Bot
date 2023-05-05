package ESI

import (
	logging "IncursionBot/internal/Logging"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type NameResponse struct {
	Category string `json:"category"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
}
type NameMap map[int]string // Map of item IDs to names

var cachedNames NameMap = make(NameMap)

func (c *ESIClient) GetNames(ids []int) (NameMap, error) {
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
		logging.Errorln("Failed to marshal IDs into json", err)
		return result, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/universe/names/", bytes.NewBuffer(data))
	if err != nil {
		logging.Errorln("Failed to create name request", req)
		return result, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logging.Errorln("Failed HTTP request for names", err)
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logging.Errorf("Name endpoint returned a status code of %d: %s", resp.StatusCode, string(body))
		return result, err
	}

	err = c.parseResults(resp, &responseData)
	if err != nil {
		logging.Errorln("Failed to parse name results", err)
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
	ID       int    `json:"constellation_id"`
	Name     string `json:"name"`
	RegionID int    `json:"region_id"`
}

var constDataCache CacheMap = make(CacheMap)

func (c *ESIClient) GetConstInfo(constID int) (ConstellationData, error) {
	var response ConstellationData
	url := fmt.Sprintf("%s/universe/constellations/%d/", c.baseURL, constID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logging.Errorf("Failed to create constellation info request for id: %d", constID)
		return response, err
	}

	cacheData := constDataCache[constID]
	err = c.cachedCall(req, &cacheData, &response)
	if err != nil {
		logging.Errorln("Error occurred in getting the constellation data", err)
		return response, err
	}

	return response, nil
}

// ----------- SYSTEM INFO -----------

type SystemData struct {
	ID            int           `json:"system_id"`
	Name          string        `json:"name"`
	SecStatus     float64       `json:"security_status"`
	Gates         []int         `json:"stargates"`
	SecurityClass SecurityClass // Not part of the response
}

var systemCache CacheMap = make(CacheMap)

func (c *ESIClient) GetSystemInfo(systemID int) (SystemData, error) {
	var results SystemData
	url := fmt.Sprintf("%s/universe/systems/%d/", c.baseURL, systemID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logging.Errorln("An error occurred creating the system info request", err)
		return results, err
	}

	cacheData := systemCache[systemID]
	err = c.cachedCall(req, &cacheData, &results)
	if err != nil {
		logging.Errorln("An error occurred getting system info", err)
		return results, err
	}
	results.SecurityClass = guessSecClass(results.SecStatus)

	return results, nil
}

// TODO: Cache this endpoint
type Route []int

func (c *ESIClient) GetRouteLength(startSystem int, endSystem int) (int, error) {
	var resultData Route
	url := fmt.Sprintf("%s/route/%d/%d/", c.baseURL, startSystem, endSystem)
	resp, err := http.Get(url)
	if err != nil {
		logging.Errorln("Failed HTTP request for route length", err)
		return -1, err
	}

	err = c.parseResults(resp, &resultData)
	if err != nil {
		logging.Errorln("Error occurred parsing results", err)
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
	} else if status > .1 {
		return LowSec
	}
	return NullSec
}

type StargateResponse struct {
	Destination struct {
		GateID   int `json:"stargate_id"`
		SystemID int `json:"system_id"`
	} `json:"destination"`
	Name   string `json:"name"`
	GateID int    `json:"stargate_id"`
}

var gateCache = make(CacheMap)

func (c *ESIClient) GetStargateData(gateID int) (StargateResponse, error) {
	var resultData StargateResponse
	url := fmt.Sprintf("%s/universe/stargates/%d/", c.baseURL, gateID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resultData, err
	}

	cacheData := gateCache[gateID]
	err = c.cachedCall(req, &cacheData, &resultData)
	return resultData, err
}

func (c *ESIClient) GetSystemConnections(systemID int) ([]StargateResponse, error) {
	systemData, err := c.GetSystemInfo(systemID)
	var result []StargateResponse

	if err != nil {
		return []StargateResponse{}, err
	}
	resultChan := make(chan StargateResponse)

	for _, gate := range systemData.Gates {
		go func(result chan<- StargateResponse, gateID int) {
			res, err := c.GetStargateData(gateID)
			if err != nil {
				resultChan <- StargateResponse{}
				fmt.Println(err)
			} else {
				resultChan <- res
			}
		}(resultChan, gate)
	}

	for range systemData.Gates {
		out := <-resultChan
		result = append(result, out)
	}

	return result, nil
}

type SovResponse struct {
	Alliance    int `json:"alliance_id"`
	Corporation int `json:"corporation_id"`
	Faction     int `json:"faction_id"`
	System      int `json:"system_ID"`
}

var sovCache CacheEntry

func (c *ESIClient) GetSovMap() ([]SovResponse, error) {
	var response []SovResponse
	url := fmt.Sprintf("%s/sovereignty/map", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}

	err = c.cachedCall(req, &sovCache, &response)
	return response, err
}
