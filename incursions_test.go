package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestUpdateIncursion(t *testing.T) {
	var testIncursion Incursion


	testIncursion.State = string(Established)
	testIncursion.Influence = 0

	res := testIncursion.Update(.5, string(Established))
	assert.False(t, res)
	assert.Equal(t, .5, testIncursion.Influence)

	res = testIncursion.Update(.5, string(Mobilizing))
	assert.True(t, res)
	assert.Equal(t, string(Mobilizing), testIncursion.State)
}

func TestFind(t *testing.T) {
	var testList IncursionList
	testIncursion := Incursion {
		StagingSystem: NamedItem{ID: 1234},
	}

	testList = append(testList, testIncursion)

	assert.NotEmpty(t, testList.find(testIncursion))

	newIncursion := Incursion {
		StagingSystem: NamedItem{ID: 2345},
	}

	assert.Empty(t, testList.find(newIncursion))
}