package incursions

import (
	"IncursionBot/internal/ESI"
	logging "IncursionBot/internal/Logging"
	"IncursionBot/internal/Utils"
)

func calculateLayoutAmounts(numberOfSystems int) (assaults int, vanguards int) {
	vanguards = numberOfSystems - 3 // Staging, HQ, and the 1 mandatory assault
	assaults = 1

	if vanguards > 3 {
		vanguards--
		assaults++
	}

	if vanguards > 4 {
		vanguards--
		assaults++
	}

	return
}

func GenerateIncursionLayout(incursion *Incursion, esi *ESI.ESIClient) (layout IncursionLayout) {
	assaults, vanguards := calculateLayoutAmounts(len(incursion.Systems))

	stagingID := incursion.Layout.StagingSystem.ID

	logging.Debugf("Staging %d, should be %d vanguards and %d assaults", stagingID, vanguards, assaults)
	layout.StagingSystem = incursion.Layout.StagingSystem // Preserve previously known staging
	layout.HQSystem.Name = "Unknown"                      // Default HQ name

	connections, _ := esi.GetSystemConnections(stagingID)
	data := HQGuessData{
		remainingAssaults:  assaults,
		remainingVanguards: vanguards,
		layout:             &layout,
	}
	data.visited = append(data.visited, stagingID)

	for _, gate := range connections {
		system := gate.Destination.SystemID
		if Utils.IDList(incursion.Systems).Contains(system) {
			data.queue.Add(Utils.QueueDataType{SystemID: system, Distance: 1})
		}
	}

	traverseSystems(&data, incursion.Systems, esi)
	return
}

type HQGuessData struct {
	remainingAssaults  int
	remainingVanguards int
	queue              Utils.Queue
	visited            Utils.IDList
	layout             *IncursionLayout
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

func traverseSystems(data *HQGuessData, validSystems Utils.IDList, esi *ESI.ESIClient) {
	if data.queue.IsEmpty() {
		logging.Errorln("Queue was empty, ESI issue may have occurred")
		return
	}

	currentSystem := data.queue.Pop()
	systemInfo, err := esi.GetSystemInfo(currentSystem.SystemID)

	if err != nil {
		logging.Errorln("Couldn't get system info, stopping guess", err)
		return
	}

	item := NamedItem{ID: currentSystem.SystemID, Name: systemInfo.Name}
	if data.remainingAssaults == 0 && data.remainingVanguards == 0 {
		logging.Debugf("Guessing %s is the HQ", systemInfo.Name)
		data.layout.HQSystem = item
		return
	} // Found our boy

	var connectingSystems TestList
	connectingSystems, err = esi.GetSystemConnections(currentSystem.SystemID)

	if err != nil {
		logging.Errorf("Couldn't get the system connections to %d, stopping guessing: %v",
			currentSystem.SystemID,
			err)
	}

	// Guess the type
	if data.remainingVanguards > 0 {
		data.remainingVanguards--
		logging.Debugf("Guessing %s is a vanguard", systemInfo.Name)
		data.layout.VanguardSystems = append(data.layout.VanguardSystems, item)
	} else {
		data.remainingAssaults--
		logging.Debugf("Guessing %s is an assault", systemInfo.Name)
		data.layout.AssaultSystems = append(data.layout.AssaultSystems, item)
	}

	data.visited = append(data.visited, currentSystem.SystemID)

	for _, gate := range connectingSystems {
		system := gate.Destination.SystemID

		if !data.visited.Contains(system) && validSystems.Contains(system) {
			data.queue.Add(Utils.QueueDataType{SystemID: system, Distance: currentSystem.Distance + 1})
		}
	}

	traverseSystems(data, validSystems, esi)
}
