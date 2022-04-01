package main

import (
	"fmt"
	"math"
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


// TODO: Figure out what to do with this enum
type IncursionState string

const (
  Established IncursionState = "Established"
  Mobilizing  IncursionState = "Mobilizing"
  Withdrawing IncursionState = "Withdrawing"
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
}

func (inc *Incursion) ToString() string {
  return fmt.Sprintf("%s {%.2f} (%s - %s)", inc.StagingSystem.Name, inc.SecStatus, inc.Constellation.Name, inc.Region.Name)
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