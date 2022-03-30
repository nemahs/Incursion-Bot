package main

import (
	"fmt"
	"time"
)

type SecurityClass string

const (
  HighSec SecurityClass = "High"
  LowSec  SecurityClass = "Low"
  NullSec SecurityClass = "Null"
)

func guessSecClass(status float64) SecurityClass {
  if status > .5 {
    return HighSec
  } else if (status > .1) {
    return LowSec
  }
  return NullSec
}


// TODO: Figure out what to do with this enum
type IncursionState string

const (
  Established IncursionState = "established"
  Mobilizing  IncursionState = "mobilizing"
  Withdrawing IncursionState = "withdrawing"
)

type NamedItem struct {
  Name string
  ID int
}

type Incursion struct {
  Constellation NamedItem         // Constellation the incursion is in
  StagingSystem NamedItem         // Name of the HQ system
  Influence     float64           // Influence of the incursion from 0 to 1 inclusive
  Region        NamedItem         // Region the incursion is in
  State         string            // Current state of the incursion
  Security      SecurityClass     // Security type of the staging system
  SecStatus     float64           // Security status of the staging system, -1 to 1 inclusive
  Distance      int               // Distance from home system
  StateChanged  time.Time        // Time the state changed to this current state
}

func (inc *Incursion) ToString() string {
  return fmt.Sprintf("%s {%.2f} (%s - %s)", inc.StagingSystem.Name, inc.SecStatus, inc.Constellation.Name, inc.Region.Name)
}

func (inc *Incursion) TimeLeftInSpawn() (time.Time, error) {
  logger.Infof("Stage changed: %s", inc.StateChanged)

  switch inc.State {
  case string(Established):
    return inc.StateChanged.Add(7 * 24 * time.Hour), nil
  case string(Mobilizing):
    return inc.StateChanged.Add(72 * time.Hour), nil
  case string(Withdrawing):
    return inc.StateChanged.Add(24 * time.Hour), nil
  }

  return time.Now(), fmt.Errorf("Not a state we can deal with")
}

func (inc *Incursion) TimeLeftString() string {
  if inc.StateChanged.IsZero() {
    return "Unknown"
  }

  despawn, _ := inc.TimeLeftInSpawn()
  logger.Infof("Despawn result: %s", despawn)
  if (inc.State == string(Established)) {
    return fmt.Sprintf("NLT %s", despawn.UTC().Format(timeFormat))
  }
  
  return fmt.Sprintf("%s", despawn.UTC().Format(timeFormat))
}

type IncursionList []Incursion
func (list *IncursionList) find(inc Incursion) *Incursion {
  for _, incursion := range *list {
    if incursion.StagingSystem.ID == inc.StagingSystem.ID { return &incursion }
  }
  return nil
}

// Updates the give incursion wih new data. Returns true if the state changed, False otherwise.
func (incursion *Incursion) Update(influence float64, state string) bool {
  incursion.Influence = influence

  if incursion.State != state {
    incursion.State = state
    return true
  }

  return false
}