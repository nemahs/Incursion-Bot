package incursions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateIncursion(t *testing.T) {
	var testIncursion Incursion

	testIncursion.State = Established
	testIncursion.Influence = 0

	res := testIncursion.Update(.5, Established)
	assert.False(t, res)
	assert.Equal(t, .5, testIncursion.Influence)

	res = testIncursion.Update(.5, Mobilizing)
	assert.True(t, res)
	assert.Equal(t, Mobilizing, testIncursion.State)
}

func TestFind(t *testing.T) {
	var testList IncursionList
	testIncursion := Incursion{
		Layout: IncursionLayout{StagingSystem: NamedItem{ID: 1234}},
	}

	testList = append(testList, testIncursion)

	assert.NotEmpty(t, testList.Find(testIncursion))

	newIncursion := Incursion{
		Layout: IncursionLayout{StagingSystem: NamedItem{ID: 2345}},
	}

	assert.Empty(t, testList.Find(newIncursion))
}
