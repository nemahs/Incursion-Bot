package incursions

import (
	"IncursionBot/internal/ESI"
	logging "IncursionBot/internal/Logging"
	"fmt"
	"sync"
	"time"
)

type NotifFunction func(Incursion)

type IncursionManager struct {
	incursionMut            sync.Mutex
	incursions              IncursionList
	nullTracker, lowTracker SpawnTracker

	OnNewIncursion     NotifFunction
	OnIncursionUpdate  NotifFunction
	OnIncursionDespawn NotifFunction
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

func (manager *IncursionManager) PopulateIncursions(initialList IncursionList, client *ESI.ESIClient) {
	var toSave IncursionList

	for _, incursion := range initialList {
		if incursion.Security == HighSec {
			continue
		}

		incursion.Layout = GenerateIncursionLayout(&incursion, client)
		logging.Infof("Found initial incursion in %s", incursion.ToString())
		toSave = append(toSave, incursion)
	}

	manager.incursionMut.Lock()
	manager.incursions = toSave
	manager.incursionMut.Unlock()
}

func (manager *IncursionManager) ProcessIncursions(newIncursions IncursionList, client *ESI.ESIClient) {
	var toSave IncursionList
	logging.Infoln("------Processing new set of incursions-----")

	for _, incursion := range newIncursions {
		if incursion.Security == HighSec {
			continue // We do not give a fuck about high sec
		}

		existingIncursion := manager.incursions.Find(incursion)

		if existingIncursion == nil {
			if !incursion.IsValid {
				logging.Errorf("Received an invalid incursion located in %d, discarding...", incursion.Layout.StagingSystem.ID)
				continue
			}
			incursion.StateChanged = time.Now()

			if incursion.Security == NullSec {
				manager.nullTracker.Spawn(incursion)
			} else {
				manager.lowTracker.Spawn(incursion)
			}

			incursion.Layout = GenerateIncursionLayout(&incursion, client)
			manager.OnNewIncursion(incursion)
			toSave = append(toSave, incursion)
		} else {
			logging.Infof("Found existing incursion in %s to update", existingIncursion.ToString())

			// Attempt to regenerate the spawn layout if generation was interrupted by ESI/networking previously
			if !existingIncursion.Layout.IsComplete() {
				incursion.Layout = GenerateIncursionLayout(existingIncursion, client)
			}

			if existingIncursion.Update(incursion.Influence, incursion.State) {
				existingIncursion.StateChanged = time.Now()

				if incursion.Security == NullSec {
					manager.nullTracker.Update(*existingIncursion)
				} else {
					manager.lowTracker.Update(*existingIncursion)
				}

				manager.OnIncursionUpdate(*existingIncursion)
			}

			toSave = append(toSave, *existingIncursion)
		}
	}

	// Check for despawns
	for _, existingIncursion := range manager.incursions {
		if newIncursions.Find(existingIncursion) == nil {
			logging.Infof("Incursion in %s despawned", existingIncursion.ToString())
			if existingIncursion.Security == NullSec {
				manager.nullTracker.Despawn(existingIncursion)
			} else {
				manager.lowTracker.Despawn(existingIncursion)
			}

			manager.OnIncursionDespawn(existingIncursion)
		}
	}

	manager.incursionMut.Lock()
	manager.incursions = toSave
	manager.incursionMut.Unlock()
}
