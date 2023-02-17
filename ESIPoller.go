package main

import (
	"IncursionBot/internal/ESI"
	"time"
)

func pollESI(incursionChan chan<- IncursionList) {
	esi := ESI.NewClient()

	for {
		incursionResponses, nextPollTime, err := esi.GetIncursions()
		if err != nil {
			logger.Warningln("Error getting basic incursion data, sleeping 1 min then reattempting", err)
			time.Sleep(time.Minute)
			continue
		}

		var incursions IncursionList
		for _, response := range incursionResponses {
			newIncursion := createIncursion(response, esi)
			incursions = append(incursions, newIncursion)
		}

		incursionChan <- incursions
		time.Sleep(time.Until(nextPollTime))
	}
}

func getDistance(stagingID int, client ESI.ESIClient, resultChan chan<- int) {
	distance, err := client.GetRouteLength(homeSystem, stagingID)
	if err != nil {
		logger.Errorf("Ran into error when getting system data for incursion in %d", stagingID)
		resultChan <- -1
		return
	}

	resultChan <- distance
}

func createIncursion(incursion ESI.IncursionResponse, client ESI.ESIClient) Incursion {
	newIncursion := Incursion{
		Constellation: NamedItem{ID: incursion.ConstellationID},
		StagingSystem: NamedItem{ID: incursion.StagingID},
		Influence:     incursion.Influence,
		State:         parseState(incursion.State),
		StateChanged:  time.Time{},
		IsValid:       false,
	}

	distanceChan := make(chan int)
	go getDistance(incursion.StagingID, client, distanceChan)

	stagingData, err := client.GetSystemInfo(incursion.StagingID)
	if err != nil {
		logger.Errorf("Ran into error when getting system data for incursion in %d", newIncursion.StagingSystem.ID)
		return newIncursion
	}

	newIncursion.StagingSystem.Name = stagingData.Name
	newIncursion.SecStatus = stagingData.SecStatus
	newIncursion.Security = guessSecClass(newIncursion.SecStatus)

	constData, err := client.GetConstInfo(incursion.ConstellationID)
	if err != nil {
		logger.Errorf("Ran into error when getting system data for incursion in %d", newIncursion.StagingSystem.ID)
		return newIncursion
	}

	newIncursion.Constellation.Name = constData.Name
	newIncursion.Region.ID = constData.RegionID

	names, err := client.GetNames([]int{constData.RegionID})
	if err != nil {
		logger.Errorf("Ran into error when getting system data for incursion in %d", newIncursion.StagingSystem.ID)
		return newIncursion
	}

	newIncursion.Distance = <-distanceChan

	newIncursion.Region.Name = names[constData.RegionID]

	newIncursion.IsValid = true
	return newIncursion
}
