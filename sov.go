package main

import "IncursionBot/internal/ESI"

func GetSovOwner(systemID int, esi *ESI.ESIClient) string {
	result := ""

	sovList, err := esi.GetSovMap()

	if err != nil {
		logger.Errorln("Error occurred getting sov map", err)
		return ""
	}

	for _, sov := range sovList {
		if sov.System == systemID {

			if sov.Alliance == 0 {
				return "" // Owned by no one
			}

			allianceData, err := esi.GetAllianceData(sov.Alliance)

			if err != nil {
				logger.Errorln("Error occurred getting alliance data", err)
				return ""
			}

			result = allianceData.Ticker
			logger.Infof("Guessed sov owner was %s", result)
			break
		}
	}

	return result
}
