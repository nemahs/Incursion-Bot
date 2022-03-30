package main

import "sync"

type NotifFunction func(Incursion)

// TODO: Manage respawn intervals
type IncursionManager struct {
	incursionMut sync.Mutex
	incursions   IncursionList

	onNewIncursion     NotifFunction
	onIncursionUpdate  NotifFunction
	onIncursionDespawn NotifFunction
}

const nullSecSpawns int = 3
const lowSecSpawns int = 1

func (manager *IncursionManager) GetIncursions() IncursionList {
	manager.incursionMut.Lock()
	defer manager.incursionMut.Unlock()
	return manager.incursions
}

func (manager *IncursionManager) PopulateIncursions(initialList IncursionList) {
	var toSave IncursionList

	for _, incursion := range initialList {
		if incursion.Security == HighSec {
			continue
		}

		logger.Infof("Found initial incursion in %s", incursion.ToString())
		toSave = append(toSave, incursion)
	}

	manager.incursionMut.Lock()
	manager.incursions = toSave
	manager.incursionMut.Unlock()
}

func (manager *IncursionManager) ProcessIncursions(newIncursions IncursionList) {
	var toSave IncursionList

	for _, incursion := range newIncursions {
		if (incursion.Security == HighSec) {
			continue // We do not give a fuck about high sec
		}

		existingIncursion := manager.incursions.find(incursion)

		if existingIncursion == nil {
			logger.Infof("Found new incursion in %s", incursion.ToString())

			manager.onNewIncursion(incursion)
		} else {
			logger.Infof("Found existing incursion in %s to update", existingIncursion.ToString())
			if existingIncursion.Update(incursion.Influence, incursion.State) {
				manager.onIncursionUpdate(incursion)
			}
		}

		toSave = append(toSave, incursion)
	}

	// Check for despawns
	for _, existingIncursion := range manager.incursions {
		if newIncursions.find(existingIncursion) == nil {
			logger.Infof("Incursion in %s despawned", existingIncursion.ToString())
			manager.onIncursionDespawn(existingIncursion)
		}
	}

	manager.incursionMut.Lock()
	manager.incursions = toSave
	manager.incursionMut.Unlock()
}