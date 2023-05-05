package main

import (
	"IncursionBot/internal/ESI"
	logging "IncursionBot/internal/Logging"
)

func GetSovOwner(systemID int, esi *ESI.ESIClient) string {
	result := ""

	sovList, err := esi.GetSovMap()

	if err != nil {
		logging.Errorln("Error occurred getting sov map", err)
		return ""
	}

	for _, sov := range sovList {
		if sov.System == systemID {

			if sov.Alliance == 0 {
				return "" // Owned by no one
			}

			allianceData, err := esi.GetAllianceData(sov.Alliance)

			if err != nil {
				logging.Errorln("Error occurred getting alliance data", err)
				return ""
			}

			result = allianceData.Ticker
			logging.Infof("Guessed sov owner was %s", result)
			break
		}
	}

	return result
}
