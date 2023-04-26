package ESI

import (
	"fmt"
	"net/http"
)

type AllianceDetailResponse struct {
	Name   string
	Ticker string
}

var allianceCache = make(CacheMap)

func (c *ESIClient) GetAllianceData(allianceID int) (AllianceDetailResponse, error) {
	var resultData AllianceDetailResponse
	url := fmt.Sprintf("%s/alliances/%d", c.baseURL, allianceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resultData, err
	}

	cacheData := allianceCache[allianceID]
	err = c.cachedCall(req, &cacheData, &resultData)
	return resultData, err
}
