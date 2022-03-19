package main

import (
	"IncursionBot/internal/ESI"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FailIncursionData struct {
	FailSystem, FailConst, FailNames, FailRoute bool
}

func (c FailIncursionData) GetSystemInfo(int) (ESI.SystemData, error) {
	if c.FailSystem {
		return ESI.SystemData{}, fmt.Errorf("")
	}

	return ESI.SystemData{
		ID: 1234,
		Name: "TestSystem",
		SecStatus: -1.0,
		SecurityClass: ESI.LowSec,
	}, nil
}

func (c FailIncursionData) GetConstInfo(int) (ESI.ConstellationData, error) {
	if c.FailConst {
		return ESI.ConstellationData{}, fmt.Errorf("")
	}

	return ESI.ConstellationData{
		ID: 1234,
		Name: "TestConst",
		RegionID: 2345,
	}, nil
}

func (c FailIncursionData) GetNames([]int) (ESI.NameMap, error) {
	if c.FailNames {
		return ESI.NameMap{}, fmt.Errorf("")
	}

	return ESI.NameMap{
		2345: "TestRegion",
	}, nil
}

func (c FailIncursionData) GetRouteLength(int, int) (int, error) {
	if c.FailRoute {
		return -1, fmt.Errorf("")
	}

	return 5, nil
}

func TestNewIncursion(t *testing.T) {
	assert := assert.New(t)

	var mockFetcher FailIncursionData = FailIncursionData{
		FailSystem: true,
		FailConst: true,
		FailNames: true,
		FailRoute: true,
	}
	testResponse := ESI.IncursionResponse {
		Influence: .5,
		StagingID: 1234,
		State: string(Established),
	}

	_, err := CreateNewIncursion(testResponse, mockFetcher)
	assert.Error(err)

	mockFetcher.FailSystem = false
	_, err = CreateNewIncursion(testResponse, mockFetcher)
	assert.Error(err)

	mockFetcher.FailConst = false
	_, err = CreateNewIncursion(testResponse, mockFetcher)
	assert.Error(err)

	mockFetcher.FailNames = false
	_, err = CreateNewIncursion(testResponse, mockFetcher)
	assert.Error(err)

	mockFetcher.FailRoute = false
	res, err := CreateNewIncursion(testResponse, mockFetcher)
	assert.NoError(err)
	assert.Equal("TestConst", res.Constellation.Name)
	assert.Equal(5, res.Distance)
	assert.Equal(.5, res.Influence)
	assert.Equal("TestRegion", res.Region.Name)
	assert.Equal(-1.0, res.SecStatus)
	assert.Equal(ESI.LowSec, res.Security)
	assert.Equal(1234, res.StagingSystem.ID)
	assert.Equal(string(Established), res.State)
}


func TestUpdateIncursion(t *testing.T) {
	var testIncursion Incursion
	var testResponse ESI.IncursionResponse

	var testPtr *Incursion = nil
	res := testPtr.Update(testResponse)
	assert.False(t, res)

	testIncursion.State = string(Established)
	testIncursion.Influence = 0
	testResponse.State = string(Established)
	testResponse.Influence = .5

	res = testIncursion.Update(testResponse)
	assert.False(t, res)
	assert.Equal(t, testResponse.Influence, testIncursion.Influence)

	testResponse.State = string(Mobilizing)
	res = testIncursion.Update(testResponse)
	assert.True(t, res)
	assert.Equal(t, testResponse.State, testIncursion.State)
}