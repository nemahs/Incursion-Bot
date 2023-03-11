package main

import (
	logging "IncursionBot/internal/Logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRespawnTime(t *testing.T) {
	assert := assert.New(t)
	testInc := Incursion{}
	testInc.StateChanged = time.Now()
	logger = logging.NewLogger(false)

	t.Run("Established Respawn", func(t *testing.T) {
		testInc.State = Established
		assert.Equal(testInc.StateChanged.Add(((8*24)+12)*time.Hour), respawnTime(testInc))
	})

	t.Run("Mobilizing Respawn", func(t *testing.T) {
		testInc.State = Mobilizing
		assert.Equal(testInc.StateChanged.Add(84*time.Hour), respawnTime(testInc))
	})

	t.Run("Withdrawing Respawn", func(t *testing.T) {
		testInc.State = Withdrawing
		assert.Equal(testInc.StateChanged.Add(36*time.Hour), respawnTime(testInc))
	})

	t.Run("Respawning - pre window", func(t *testing.T) {
		testInc.State = Respawning
		assert.Equal(testInc.StateChanged.Add(12*time.Hour), respawnTime(testInc))
	})

	t.Run("Respawning - in window", func(t *testing.T) {
		testInc.StateChanged = time.Now().Add(-13 * time.Hour)
		assert.Equal(testInc.StateChanged.Add(12*time.Hour), respawnTime(testInc))
	})

	t.Run("Bad state", func(t *testing.T) {
		testInc.State = Unknown
		assert.Zero(respawnTime(testInc))
	})

	t.Run("Zero time", func(t *testing.T) {
		testInc.State = Mobilizing
		testInc.StateChanged = time.Time{}

		assert.Zero(respawnTime(testInc))
	})
}

func TestNextRespawn(t *testing.T) {
	assert := assert.New(t)
	var testSubject SpawnTracker
	testTime := time.Now()
	timeOne := formatDuration(36*time.Hour - time.Second)
	logger = logging.NewLogger(false)
	time.Sleep(200)

	t.Run("1 incursion, no info", func(t *testing.T) {
		testSubject.currentIncursions = append(testSubject.currentIncursions, Incursion{})
		assert.Equal(unknownString, testSubject.nextRespawn())
	})

	t.Run("1 incursion, valid info", func(t *testing.T) {
		testSubject.currentIncursions[0].State = Withdrawing
		testSubject.currentIncursions[0].StateChanged = testTime

		assert.Equal(timeOne, testSubject.nextRespawn())
	})

	t.Run("3 incursions, some valid", func(t *testing.T) {
		testSubject.currentIncursions = append(testSubject.currentIncursions, Incursion{
			State:        Established,
			StateChanged: testTime,
		})

		testSubject.currentIncursions = append(testSubject.currentIncursions, Incursion{
			State: Withdrawing,
		})

		assert.Equal(timeOne, testSubject.nextRespawn())
	})

	t.Run("3 incursions, all valid", func(t *testing.T) {})

	t.Run("3 incursions, 2 respawning", func(t *testing.T) {})
}

func TestIncursionManagement(t *testing.T) {
	assert := assert.New(t)
	var testSubject SpawnTracker
	logger = logging.NewLogger(false)

	t.Run("Spawn incursion", func(t *testing.T) {
		newInc := Incursion{StagingSystem: NamedItem{ID: 1}}
		testSubject.Spawn(newInc)

		assert.Equal(1, len(testSubject.currentIncursions))
		assert.Empty(testSubject.respawningIncursions)
	})

	t.Run("Update incursion", func(t *testing.T) {
		updateInc := Incursion{StagingSystem: NamedItem{ID: 1}, State: Mobilizing, StateChanged: time.Now()}
		testSubject.Update(updateInc)

		assert.Equal(1, len(testSubject.currentIncursions))
		assert.Equal(Mobilizing, testSubject.currentIncursions[0].State)
		assert.NotZero(testSubject.currentIncursions[0].StateChanged)
	})

	t.Run("Update non-existing", func(t *testing.T) {
		newInc := Incursion{StagingSystem: NamedItem{ID: 2}, State: Withdrawing, StateChanged: time.Now()}
		testSubject.Update(newInc)

		assert.Equal(2, len(testSubject.currentIncursions))
		assert.Empty(testSubject.respawningIncursions)
		assert.Equal(2, testSubject.currentIncursions[1].StagingSystem.ID)
		assert.Equal(Withdrawing, testSubject.currentIncursions[1].State)
		assert.NotZero(testSubject.currentIncursions[1].StateChanged)
	})

	t.Run("Despawn", func(t *testing.T) {
		deadInc := Incursion{StagingSystem: NamedItem{ID: 1}}
		testSubject.Despawn(deadInc)

		assert.Equal(1, len(testSubject.currentIncursions))
		assert.Equal(1, len(testSubject.respawningIncursions))
		assert.Equal(1, testSubject.respawningIncursions[0].StagingSystem.ID)
		assert.Equal(2, testSubject.currentIncursions[0].StagingSystem.ID)
		assert.Equal(Respawning, testSubject.respawningIncursions[0].State)
		assert.NotZero(testSubject.respawningIncursions[0].StateChanged)
	})

	t.Run("Respawn", func(t *testing.T) {
		newInc := Incursion{StagingSystem: NamedItem{ID: 3}}
		testSubject.Spawn(newInc)

		assert.Equal(2, len(testSubject.currentIncursions))
		assert.Zero(len(testSubject.respawningIncursions))
	})
}
