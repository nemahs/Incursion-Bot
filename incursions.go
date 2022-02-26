package main

import (
	"fmt"
)

// TODO: Figure out what to do with this enum
type IncursionState string

const (
	Established IncursionState = "Established"
	Mobilizing  IncursionState = "Mobilizing"
	Withdrawing	IncursionState = "Withdrawing"
)


type Incursion struct {
	StagingID     int							// ID of the staging system in this incursion. Used to uniquely identify incursions
	Constellation string					// Constellation the incursion is in
	HQSystem      string					// Name of the HQ system
	Influence     float64					// Influence of the incursion from 0 to 1 inclusive
	Region        string					// Region the incursion is in
	State         IncursionState	// Current state of the incursion
	Security      SecurityClass		// Security type of the staging system
	SecStatus     float64					// Security status of the staging system, -1 to 1 inclusive
	Distance      int							// Distance from home system
}

func (inc *Incursion) ToString() string {
	return fmt.Sprintf("%s {%.2f} (%s - %s)", inc.HQSystem, inc.SecStatus, inc.Constellation, inc.Region)
}

type IncursionList []Incursion
func (list *IncursionList) find(stagingId int) *Incursion {
  for _, incursion := range *list {
    if incursion.StagingID == stagingId { return &incursion }
  }
  return nil
}

// Updates the give incursion wih new data. Returns true if the state changed, False otherwise.
func updateIncursion(incursion *Incursion, newData IncursionResponse) bool {
  updated := false

  if incursion.State != newData.State {
    incursion.State = newData.State
    updated = true
  }

  incursion.Influence = newData.Influence
  return updated
}

// Creates a new Incursion object from ESI data
func createNewIncursion(incursion IncursionResponse) Incursion {
  stagingData := getSystemInfo(incursion.StagingID)
  constData := getConstInfo(incursion.ConstellationID)
  names := getNames([]int{constData.RegionID, incursion.StagingID})
  distance := GetRouteLength(homeSystem, incursion.StagingID)

  newIncursion := Incursion{
    StagingID: incursion.StagingID,
    Constellation: constData.Name,
    HQSystem: names[incursion.StagingID],
    Influence: incursion.Influence,
    Region: names[constData.RegionID],
    State: incursion.State,
    SecStatus: stagingData.SecStatus,
    Security: stagingData.SecurityClass,
    Distance: distance,
  }

  return newIncursion
}