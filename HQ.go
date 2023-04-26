package main

import (
	"IncursionBot/internal/ESI"
	"IncursionBot/internal/Utils"
)

func GuessHQSystem(incursion ESI.IncursionResponse, esi ESI.ESIClient) (hqSystem NamedItem) {
	vanguards := len(incursion.IncursionSystems) - 3 // Staging and HQ
	assaults := 1

	if vanguards > 3 {
		vanguards--
		assaults++
	}

	if vanguards > 4 {
		vanguards--
		assaults++
	}
	logger.Debugf("Staging %d, should be %d vanguards and %d assaults", incursion.StagingID, vanguards, assaults)

	connections, _ := esi.GetSystemConnections(incursion.StagingID)
	data := HQGuessData{
		remainingAssaults:  assaults,
		remainingVanguards: vanguards,
	}
	data.visited = append(data.visited, incursion.StagingID)

	for _, gate := range connections {
		system := gate.Destination.SystemID
		if Utils.IDList(incursion.IncursionSystems).Contains(system) {
			data.queue.Add(Utils.QueueDataType{SystemID: system, Distance: 1})
		}
	}

	// Error handling in case ESI dorks up
	defer func() {
		if r := recover(); r != nil {
			logger.Errorln("Recovered from panic while guessing HQ system", r)
			hqSystem = NamedItem{Name: "Unknown", ID: 0}
		}
	}()

	hqSystem = traverseSystems(&data, incursion.IncursionSystems, esi)

	logger.Infof("Guessed that HQ system was %s\n", hqSystem.Name)
	return hqSystem
}

type HQGuessData struct {
	remainingAssaults  int
	remainingVanguards int
	queue              Utils.Queue
	visited            Utils.IDList
}

type TestList []ESI.StargateResponse

// Sorts the list from highest ID to lowest
func (list *TestList) ReverseSort() {
	var result TestList
	for _, entry := range *list {
		result = append(TestList{entry}, result...)
	}

	*list = result
}

func traverseSystems(data *HQGuessData, validSystems Utils.IDList, esi ESI.ESIClient) NamedItem {
	if data.queue.IsEmpty() {
		panic("Queue was empty, ESI issue may have occurred")
	}

	currentSystem := data.queue.Pop()

	systemInfo, _ := esi.GetSystemInfo(currentSystem.SystemID)
	if data.remainingAssaults == 0 && data.remainingVanguards == 0 {
		logger.Debugf("Guessing %s is the HQ", systemInfo.Name)
		return NamedItem{ID: currentSystem.SystemID, Name: systemInfo.Name}
	} // Found our boy

	var connectingSystems TestList
	connectingSystems, _ = esi.GetSystemConnections(currentSystem.SystemID)

	// Guess the type
	if data.remainingVanguards > 0 {
		data.remainingVanguards--
		logger.Debugf("Guessing %s is a vanguard", systemInfo.Name)
	} else {
		data.remainingAssaults--
		logger.Debugf("Guessing %s is an assault", systemInfo.Name)
	}

	data.visited = append(data.visited, currentSystem.SystemID)

	for _, gate := range connectingSystems {
		system := gate.Destination.SystemID

		if !data.visited.Contains(system) && validSystems.Contains(system) {
			data.queue.Add(Utils.QueueDataType{SystemID: system, Distance: currentSystem.Distance + 1})
		}
	}

	return traverseSystems(data, validSystems, esi)
}
