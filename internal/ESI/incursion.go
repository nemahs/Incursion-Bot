package ESI

import (
	"net/http"
	"time"
)

type IncursionResponse struct {
	ConstellationID  int     `json:"constellation_id"`
	IncursionSystems []int   `json:"infested_solar_systems"`
	Influence        float64 `json:"influence"`
	StagingID        int     `json:"staging_solar_system_id"`
	State            string  `json:"state"`
}

var incursionsCache CacheEntry

func (c *ESIClient) GetIncursions() ([]IncursionResponse, time.Time, error) {
	var result []IncursionResponse
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/incursions/", nil)
	if err != nil {
		errorLog.Println("Failed to create request for incursions", err)
		return result, incursionsCache.ExpirationTime, err
	}
	err = c.cachedCall(req, &incursionsCache, &result)

	if err != nil {
		errorLog.Println("Error occured while getting incursions", err)
		return result, incursionsCache.ExpirationTime, err
	}

	return result, incursionsCache.ExpirationTime, nil
}
