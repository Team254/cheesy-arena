// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestControlPanelGetStage3TargetColor(t *testing.T) {
	rand.Seed(0)
	var controlPanel ControlPanel

	controlPanel.CurrentColor = ColorUnknown
	results := getStage3TargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 2543, 2527, 2510, 2420}, results)

	controlPanel.CurrentColor = ColorRed
	results = getStage3TargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 0, 3351, 3311, 3338}, results)

	controlPanel.CurrentColor = ColorGreen
	results = getStage3TargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3335, 0, 3320, 3345}, results)

	controlPanel.CurrentColor = ColorBlue
	results = getStage3TargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3328, 3296, 0, 3376}, results)

	controlPanel.CurrentColor = ColorYellow
	results = getStage3TargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3303, 3388, 3309, 0}, results)
}

// Invokes the method N times and returns a map of the counts for each result, for statistical testing.
func getStage3TargetColorNTimes(controlPanel *ControlPanel, n int) [5]int {
	var results [5]int
	for i := 0; i < n; i++ {
		results[controlPanel.GetStage3TargetColor()]++
	}
	return results
}
