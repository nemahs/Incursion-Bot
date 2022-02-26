package main

import (
	"fmt"
	"log"
)

type Incursion struct {
	StagingID     int
	Constellation string
	HQSystem      string
	Influence     float64
	Region        string
	State         string
	Security      string
	SecStatus     float64
	Distance      int
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

func updateIncursion(incursion *Incursion, newData IncursionResponse) bool {
  updated := false

  if incursion.State != newData.State {
    incursion.State = newData.State
    updated = true
  }

  incursion.Influence = newData.Influence
  return updated
}

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
    SecStatus: stagingData.Security_Status,
    Security: string(stagingData.Security_Class),
    Distance: distance,
  }

  log.Printf("Incursion: %+v", newIncursion)
  return newIncursion
}