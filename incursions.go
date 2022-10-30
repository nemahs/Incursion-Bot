package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type SecurityClass string

const (
  HighSec SecurityClass = "High"
  LowSec  SecurityClass = "Low"
  NullSec SecurityClass = "Null"
)

func guessSecClass(status float64) SecurityClass {
  roundedSecStatus := ccp_round(status)

  if roundedSecStatus >= .5 {
    return HighSec
  } else if (roundedSecStatus >= .1) {
    return LowSec
  }
  return NullSec
}

func ccp_round(status float64) float64 {
  if status > 0.0 && status < 0.05 {
    return math.Ceil(status * 10) / 10
  }

  return math.Round(status * 10) / 10
}

const mobilizingLifetime time.Duration = time.Hour * 72
const withdrawingLifetime time.Duration = time.Hour * 24
const establishedMaxLife time.Duration = time.Hour * 24 * 8

type IncursionState string

const (
  Established IncursionState = "established"
  Mobilizing  IncursionState = "mobilizing"
  Withdrawing IncursionState = "withdrawing"
  Respawning IncursionState = "respawning"
  Unknown IncursionState = ""
)

func parseState(val string) IncursionState {
  toTest := strings.ToLower(val)

  switch toTest {
  case string(Established):
    return Established
  case string(Mobilizing):
    return Mobilizing
  case string(Withdrawing):
    return Withdrawing
  default:
    return Unknown
  }
}

type NamedItem struct {
  Name string
  ID int
}

type Incursion struct {
  Constellation NamedItem         // Constellation the incursion is in
  StagingSystem NamedItem         // Name of the HQ system
  Influence     float64           // Influence of the incursion from 0 to 1 inclusive
  Region        NamedItem         // Region the incursion is in
  State         IncursionState    // Current state of the incursion
  Security      SecurityClass     // Security type of the staging system
  SecStatus     float64           // Security status of the staging system, -1 to 1 inclusive
  Distance      int               // Distance from home system
  StateChanged  time.Time        // Time the state changed to this current state
  IsValid       bool
}

func (inc *Incursion) Equal(other Incursion) bool {
  return inc.StagingSystem.ID == other.StagingSystem.ID
}

func (inc *Incursion) ToString() string {
  return fmt.Sprintf("%s {%.2f} (%s - %s)", inc.StagingSystem.Name, inc.SecStatus, inc.Constellation.Name, inc.Region.Name)
}

func (inc *Incursion) TimeLeftInSpawn() (time.Time, error) {
  logger.Infof("Stage changed: %s", inc.StateChanged)

  switch inc.State {
  case Established:
    return inc.StateChanged.Add(establishedMaxLife), nil
  case Mobilizing:
    return inc.StateChanged.Add(mobilizingLifetime), nil
  case Withdrawing:
    return inc.StateChanged.Add(withdrawingLifetime), nil
  }

  return time.Time{}, fmt.Errorf("Not a state we can deal with")
}

func (inc *Incursion) TimeLeftString() string {
  if inc.StateChanged.IsZero() {
    return "Unknown"
  }

  despawn, err := inc.TimeLeftInSpawn()
  if err != nil {
    logger.Errorf("Error occurred getting time left in spawn %s: %s", inc.StagingSystem.Name, err)
    return "Unknown"
  }

  logger.Debugf("Despawn result: %s", despawn)
  if (inc.State == Established) {
    return fmt.Sprintf("NLT %s", despawn.UTC().Format(timeFormat))
  }
  
  return fmt.Sprintf(despawn.UTC().Format(timeFormat))
}


// Updates the give incursion wih new data. Returns true if the state changed, False otherwise.
func (incursion *Incursion) Update(influence float64, state IncursionState) bool {
  incursion.Influence = influence

  if incursion.State != state {
    incursion.State = state
    incursion.StateChanged = time.Now()
    return true
  }

  return false
}