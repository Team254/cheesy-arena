// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestControlPanelGetPositionControlTargetColor(t *testing.T) {
	rand.Seed(0)
	var controlPanel ControlPanel

	controlPanel.CurrentColor = ColorUnknown
	results := getPositionTargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 2543, 2527, 2510, 2420}, results)

	controlPanel.CurrentColor = ColorRed
	results = getPositionTargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 0, 3351, 3311, 3338}, results)

	controlPanel.CurrentColor = ColorBlue
	results = getPositionTargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3335, 0, 3320, 3345}, results)

	controlPanel.CurrentColor = ColorGreen
	results = getPositionTargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3328, 3296, 0, 3376}, results)

	controlPanel.CurrentColor = ColorYellow
	results = getPositionTargetColorNTimes(&controlPanel, 10000)
	assert.Equal(t, [5]int{0, 3303, 3388, 3309, 0}, results)
}

func TestGetGameDataForColor(t *testing.T) {
	assert.Equal(t, "", GetGameDataForColor(ColorUnknown))
	assert.Equal(t, "R", GetGameDataForColor(ColorRed))
	assert.Equal(t, "B", GetGameDataForColor(ColorBlue))
	assert.Equal(t, "G", GetGameDataForColor(ColorGreen))
	assert.Equal(t, "Y", GetGameDataForColor(ColorYellow))
	assert.Equal(t, "", GetGameDataForColor(-100))
}

func TestControlPanelUpdateState(t *testing.T) {
	rand.Seed(0)
	var controlPanel ControlPanel
	controlPanel.ControlPanelStatus = ControlPanelRotation
	currentTime := time.Now()

	// Check before Stage 2 capacity is reached.
	controlPanel.UpdateState(0, false, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(30, false, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(50, false, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)

	// Check rotation control.
	controlPanel.UpdateState(60, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(80, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(37, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(36, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(40, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(35, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(21, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(20, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(44, true, false, currentTime)
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(55, true, false, currentTime.Add(1*time.Millisecond))
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(55, true, false, currentTime.Add(2000*time.Millisecond))
	assert.Equal(t, ControlPanelNone, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(55, true, false, currentTime.Add(2001*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(-1000, true, false, currentTime.Add(3000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)

	// Check position control.
	assert.Equal(t, ColorUnknown, controlPanel.positionTargetColor)
	controlPanel.UpdateState(1000, true, true, currentTime.Add(5000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	assert.Equal(t, ColorGreen, controlPanel.GetPositionControlTargetColor())
	controlPanel.CurrentColor = ColorBlue
	controlPanel.UpdateState(1001, true, true, currentTime.Add(6000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.CurrentColor = ColorGreen
	controlPanel.UpdateState(1002, true, true, currentTime.Add(7000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(1002, true, true, currentTime.Add(9999*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(1002, true, true, currentTime.Add(10000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.CurrentColor = ColorYellow
	controlPanel.UpdateState(1003, true, true, currentTime.Add(11000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(1003, true, true, currentTime.Add(20000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.CurrentColor = ColorGreen
	controlPanel.UpdateState(1002, true, true, currentTime.Add(21000*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOn, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(1002, true, true, currentTime.Add(25999*time.Millisecond))
	assert.Equal(t, ControlPanelRotation, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightFlashing, controlPanel.ControlPanelLightState)
	controlPanel.UpdateState(1002, true, true, currentTime.Add(26000*time.Millisecond))
	assert.Equal(t, ControlPanelPosition, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)
	controlPanel.CurrentColor = ColorRed
	controlPanel.UpdateState(0, true, true, currentTime.Add(26001*time.Millisecond))
	assert.Equal(t, ControlPanelPosition, controlPanel.ControlPanelStatus)
	assert.Equal(t, ControlPanelLightOff, controlPanel.ControlPanelLightState)
}

// Invokes the method N times and returns a map of the counts for each result, for statistical testing.
func getPositionTargetColorNTimes(controlPanel *ControlPanel, n int) [5]int {
	var results [5]int
	for i := 0; i < n; i++ {
		controlPanel.positionTargetColor = ColorUnknown
		results[controlPanel.GetPositionControlTargetColor()]++
	}
	return results
}
