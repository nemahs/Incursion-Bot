package main

import (
	"fmt"
	"sync"
	"time"
)

type NotifFunction func(Incursion)

type IncursionManager struct {
	incursionMut            sync.Mutex
	incursions              IncursionList
	nullTracker, lowTracker IncursionTimeTracker

	onNewIncursion     NotifFunction
	onIncursionUpdate  NotifFunction
	onIncursionDespawn NotifFunction
}

func (manager *IncursionManager) GetIncursions() IncursionList {
	manager.incursionMut.Lock()
	defer manager.incursionMut.Unlock()
	return manager.incursions
}

func (manager *IncursionManager) NextSpawns() string {
	return fmt.Sprintf("\nNext nullsec spawn window: %s\nNext lowsec spawn window: %s",
		manager.nullTracker.nextRespawn(),
		manager.lowTracker.nextRespawn())
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
	logger.Infoln("------Processing new set of incursions-----")

	for _, incursion := range newIncursions {
		if incursion.Security == HighSec {
			continue // We do not give a fuck about high sec
		}

		existingIncursion := manager.incursions.find(incursion)

		if existingIncursion == nil {
			if !incursion.IsValid {
				logger.Errorf("Received an invalid incursion located in %d, discarding...", incursion.StagingSystem.ID)
				continue
			}
			incursion.StateChanged = time.Now()

			if incursion.Security == NullSec {
				manager.nullTracker.Spawn(incursion)
			} else {
				manager.lowTracker.Spawn(incursion)
			}

			manager.onNewIncursion(incursion)
			toSave = append(toSave, incursion)
		} else {
			logger.Infof("Found existing incursion in %s to update", existingIncursion.ToString())
			if existingIncursion.Update(incursion.Influence, incursion.State) {
				existingIncursion.StateChanged = time.Now()

				if incursion.Security == NullSec {
					manager.nullTracker.Update(*existingIncursion)
				} else {
					manager.lowTracker.Update(*existingIncursion)
				}

				manager.onIncursionUpdate(*existingIncursion)
			}

			toSave = append(toSave, *existingIncursion)
		}
	}

	// Check for despawns
	for _, existingIncursion := range manager.incursions {
		if newIncursions.find(existingIncursion) == nil {
			logger.Infof("Incursion in %s despawned", existingIncursion.ToString())
			if existingIncursion.Security == NullSec {
				manager.nullTracker.Despawn(existingIncursion)
			} else {
				manager.lowTracker.Despawn(existingIncursion)
			}

			manager.onIncursionDespawn(existingIncursion)
		}
	}

	manager.incursionMut.Lock()
	manager.incursions = toSave
	manager.incursionMut.Unlock()
}
