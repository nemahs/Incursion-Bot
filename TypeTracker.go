package main

import (
	"fmt"
	"time"
)

const respawnWindowStart time.Duration = time.Hour * 12
const respawnWindowEnd time.Duration = time.Hour * 36
const day time.Duration = time.Hour * 24

type IncursionTimeTracker struct {
	currentIncursions    IncursionList
	respawningIncursions IncursionList
}

func respawnTime(incursion Incursion) time.Time {
	switch incursion.State {
	case Respawning:
		return incursion.StateChanged.Add(respawnWindowStart)
	case Established, Mobilizing, Withdrawing:
		lifeTime, _ := incursion.TimeLeftInSpawn()
		return lifeTime.Add(respawnWindowStart)
	}

	return time.Time{}
}

// The default toString only shows up to hours, we'd like to show days
func formatDuration(duration time.Duration) string {
	var result string

	if duration > day {
		result += fmt.Sprint(duration.Truncate(day), "d")
		duration = duration % day
	}

	result += duration.String()

	return result
}

func (tracker *IncursionTimeTracker) Despawn(incursion Incursion) {
	toRemove := -1
	for i, curIncursion := range tracker.currentIncursions {
		if curIncursion.StagingSystem == incursion.StagingSystem {
			toRemove = i
			break
		}
	}

	if toRemove > -1 {
		tracker.currentIncursions = append(tracker.currentIncursions[:toRemove], tracker.currentIncursions[toRemove+1:]...)
	}

	incursion.State = Respawning
	incursion.StateChanged = time.Now()
	tracker.respawningIncursions = append(tracker.respawningIncursions, incursion)
	logger.Debugln("Added respawning incursion")
}

func (tracker *IncursionTimeTracker) Spawn(incursion Incursion) {
	tracker.currentIncursions = append(tracker.currentIncursions, incursion)

	if len(tracker.respawningIncursions) > 0 {
		var toRemove int = -1
		for i, incursion := range tracker.respawningIncursions {
			if incursion.StateChanged.Before(tracker.respawningIncursions[i].StateChanged) {
				toRemove = i
			}
		}

		tracker.respawningIncursions = append(tracker.respawningIncursions[:toRemove], tracker.respawningIncursions[toRemove+1:]...)
		logger.Debugln("Removed respawning incursion")
	}

	logger.Debugln("Tracking new incursion")
}

func (tracker *IncursionTimeTracker) Update(incursion Incursion) {
	found := tracker.currentIncursions.find(incursion)

	if found != nil {
		*found = incursion
		logger.Debugln("Updated incursion")
	} else {
		tracker.currentIncursions = append(tracker.currentIncursions, incursion)
		logger.Debugln("Found an update for an incursion we weren't tracking, adding to list")
	}
}

func (tracker *IncursionTimeTracker) nextRespawn() string {
	var nextToRespawn Incursion
	var nextRespawnTime time.Time

	toCheck := append(tracker.currentIncursions, tracker.respawningIncursions...)
	for _, incursion := range toCheck {
		logger.Debugf("Considering %s", incursion.StagingSystem.Name)
		respawnTime := respawnTime(incursion)
		if respawnTime.Before(nextRespawnTime) || nextRespawnTime.IsZero() {
			logger.Debugf("%s now the next to respawn", incursion.StagingSystem.Name)
			nextToRespawn = incursion
			nextRespawnTime = respawnTime
		}
	}

	logger.Infof("Picked %s as next to respawn, respawn time %s", nextToRespawn.StagingSystem.Name, nextRespawnTime)
	switch nextToRespawn.State {
	case Established:
		return fmt.Sprintf("No later than %s", formatDuration(time.Until(nextRespawnTime)))
	case Respawning:
		if time.Now().After(nextRespawnTime) {
			endOfSpawn := nextToRespawn.StateChanged.Add(respawnWindowEnd)
			return fmt.Sprintf("Currently in a spawn window for another %s", formatDuration(time.Until(endOfSpawn)))
		}
		fallthrough // Not yet in a spawn window, return the normal string format
	default:
		if nextRespawnTime.IsZero() {
			return "Unknown"
		}
		return formatDuration(time.Until(nextRespawnTime))
	}
}