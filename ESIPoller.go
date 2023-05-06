package main

import (
	"IncursionBot/internal/ESI"
	incursions "IncursionBot/internal/Incursions"
	logging "IncursionBot/internal/Logging"
	"time"
)

func pollESI(incursionChan chan<- incursions.IncursionList) {
	for {
		incursionResponses, nextPollTime, err := esi.GetIncursions()
		if err != nil {
			logging.Warningln("Error getting basic incursion data, sleeping 1 min then reattempting", err)
			time.Sleep(time.Minute)
			continue
		}

		var incursions incursions.IncursionList
		for _, response := range incursionResponses {
			newIncursion := createIncursion(response, esi)
			incursions = append(incursions, newIncursion)
		}

		incursionChan <- incursions
		logging.Debugf("Sleeping until %s", nextPollTime.String())
		time.Sleep(time.Until(nextPollTime))
	}
}

func getDistance(stagingID int, client ESI.ESIClient, resultChan chan<- int) {
	distance, err := client.GetRouteLength(homeSystem, stagingID)
	if err != nil {
		logging.Errorf("Ran into error when getting system data for incursion in %d", stagingID)
		resultChan <- -1
		return
	}

	resultChan <- distance
}

func getSov(systemID int, client ESI.ESIClient, resultChan chan<- string) {
	resultChan <- GetSovOwner(systemID, &client)
}

func createIncursion(incursion ESI.IncursionResponse, client ESI.ESIClient) incursions.Incursion {
	newIncursion := incursions.Incursion{
		Constellation: incursions.NamedItem{ID: incursion.ConstellationID},
		Influence:     incursion.Influence,
		State:         incursions.ParseState(incursion.State),
		StateChanged:  time.Time{},
		IsValid:       false,
	}

	distanceChan := make(chan int)
	sovChan := make(chan string)
	go getDistance(incursion.StagingID, client, distanceChan)
	go getSov(incursion.StagingID, client, sovChan)

	newIncursion.Layout.StagingSystem = incursions.CreateNamedItem(incursion.StagingID, &client)
	stagingData, err := client.GetSystemInfo(incursion.StagingID)
	if err != nil {
		logging.Errorf("Ran into error when getting system data for incursion in %d", incursion.StagingID)
		return newIncursion
	}

	newIncursion.SecStatus = stagingData.SecStatus
	newIncursion.Security = incursions.ParseSecurityClass(newIncursion.SecStatus)

	constData, err := client.GetConstInfo(incursion.ConstellationID)
	if err != nil {
		logging.Errorf("Ran into error when getting system data for incursion in %d", incursion.StagingID)
		return newIncursion
	}

	newIncursion.Constellation.Name = constData.Name
	newIncursion.Region.ID = constData.RegionID

	names, err := client.GetNames([]int{constData.RegionID})
	if err != nil {
		logging.Errorf("Ran into error when getting system data for incursion in %d", incursion.StagingID)
		return newIncursion
	}

	newIncursion.Distance = <-distanceChan

	newIncursion.SovOwner = <-sovChan

	newIncursion.Region.Name = names[constData.RegionID]

	newIncursion.IsValid = true

	newIncursion.Systems = incursion.IncursionSystems

	return newIncursion
}
