package main

import (
	"fmt"
	"time"
)

const respawnWindowStart time.Duration = time.Hour * 12
const respawnWindowEnd time.Duration = time.Hour * 36
const day time.Duration = time.Hour * 24
const unknownString string = "Unknown"

type IncursionTimeTracker struct {
	currentIncursions    IncursionList
	respawningIncursions IncursionList
}

func respawnTime(incursion Incursion) time.Time {
	if incursion.StateChanged.IsZero() {
		return time.Time{} // Cannot make meaningful guess about the respawn time
	}

	switch incursion.State {
	case Respawning:
		return incursion.StateChanged.Add(respawnWindowStart)
	case Established, Mobilizing, Withdrawing:
		lifeTime, err := incursion.TimeLeftInSpawn()
		if err != nil {
			logger.Errorln("Error getting time left in spawn", err)
			return time.Time{}
		}
		return lifeTime.Add(respawnWindowStart)
	default:
		logger.Warningf("Unknown incursion state %s", incursion.State)
	}

	return time.Time{}
}

// The default toString only shows up to hours, we'd like to show days
func formatDuration(duration time.Duration) string {
	var result string

	if duration > day {
		result += fmt.Sprintf("%dd", int(duration.Hours() / 24))
		duration = duration % day
	}

	result += fmt.Sprintf("%dh", int(duration.Hours()))
	duration = duration % time.Hour

	result += fmt.Sprintf("%dm", int(duration.Minutes()))
	return result
}

func (tracker *IncursionTimeTracker) Despawn(incursion Incursion) {
	tracker.currentIncursions.RemoveFunc(incursion.Equal)

	incursion.State = Respawning
	incursion.StateChanged = time.Now()
	tracker.respawningIncursions = append(tracker.respawningIncursions, incursion)
	logger.Debugf("Added respawning incursion from %s", incursion.ToString())
}

func (tracker *IncursionTimeTracker) Spawn(incursion Incursion) {
	tracker.currentIncursions = append(tracker.currentIncursions, incursion)

	if !tracker.respawningIncursions.Empty() {
		var toRemove int = 0
		for i, incursion := range tracker.respawningIncursions {
			if incursion.StateChanged.Before(tracker.respawningIncursions[toRemove].StateChanged) {
				toRemove = i
			}
		}

		tracker.respawningIncursions.Remove(toRemove)
		logger.Debugln("Removed respawning incursion")
	}

	logger.Debugf("Tracking new incursion in %s", incursion.ToString())
}

func (tracker *IncursionTimeTracker) Update(incursion Incursion) {
	found := tracker.currentIncursions.find(incursion)

	if found != nil {
		*found = incursion
		logger.Debugln("Updated incursion")
	} else {
		tracker.currentIncursions = append(tracker.currentIncursions, incursion)
		logger.Debugf("Found an update for an incursion we weren't tracking in %s, adding to list", incursion.ToString())
	}
}

func (tracker *IncursionTimeTracker) nextRespawn() string {
	var nextToRespawn Incursion
	var nextRespawnTime time.Time

	toCheck := append(tracker.currentIncursions, tracker.respawningIncursions...)
	for _, incursion := range toCheck {
		logger.Debugf("Considering %s", incursion.StagingSystem.Name)
		respawnTime := respawnTime(incursion)
		if !respawnTime.IsZero() && (respawnTime.Before(nextRespawnTime) || nextRespawnTime.IsZero()) {
			logger.Debugf("%s now the next to respawn", incursion.StagingSystem.Name)
			nextToRespawn = incursion
			nextRespawnTime = respawnTime
		}
	}

	if nextRespawnTime.IsZero() {
		return unknownString
	}

	logger.Infof("Picked %s as next to respawn, respawn time %s", nextToRespawn.StagingSystem.Name, nextRespawnTime)
	switch nextToRespawn.State {
	case Established:
		return fmt.Sprintf("No more than %s", formatDuration(time.Until(nextRespawnTime)))
	case Respawning:
		if time.Now().After(nextRespawnTime) {
			endOfSpawn := nextToRespawn.StateChanged.Add(respawnWindowEnd)
			return fmt.Sprintf("Currently in a spawn window for another %s", formatDuration(time.Until(endOfSpawn)))
		}
		fallthrough // Not yet in a spawn window, return the normal string format
	default:
		return formatDuration(time.Until(nextRespawnTime))
	}
}