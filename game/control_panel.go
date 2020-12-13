// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents the state of an alliance Control Panel in the 2020 game.

package game

import (
	"math"
	"math/rand"
	"time"
)

type ControlPanel struct {
	CurrentColor ControlPanelColor
	ControlPanelStatus
	ControlPanelLightState
	rotationStarted           bool
	rotationStartSegmentCount int
	lastSegmentCountDiff      int
	rotationStopTime          time.Time
	positionTargetColor       ControlPanelColor
	lastPositionCorrect       bool
	positionStopTime          time.Time
}

type ControlPanelColor int

// This ordering matches the values in the official FRC PLC code: 0:UnknownError, 1:Red, 2:Blue, 3:Green, 4:Yellow
const (
	ColorUnknown ControlPanelColor = iota
	ColorRed
	ColorBlue
	ColorGreen
	ColorYellow
)

type ControlPanelStatus int

const (
	ControlPanelNone ControlPanelStatus = iota
	ControlPanelRotation
	ControlPanelPosition
)

type ControlPanelLightState int

const (
	ControlPanelLightOff ControlPanelLightState = iota
	ControlPanelLightOn
	ControlPanelLightFlashing
)

const (
	rotationControlMinSegments        = 24
	rotationControlMaxSegments        = 40
	rotationControlStopDurationSec    = 2
	positionControlStopMinDurationSec = 3
	positionControlStopMaxDurationSec = 5
)

// Updates the internal state of the control panel given the current state of the hardware counts and the rest of the
// score.
func (controlPanel *ControlPanel) UpdateState(segmentCount int, stage2AtCapacity, stage3AtCapacity bool,
	currentTime time.Time) {
	if !stage2AtCapacity {
		controlPanel.ControlPanelStatus = ControlPanelNone
		controlPanel.ControlPanelLightState = ControlPanelLightOff
	} else if controlPanel.ControlPanelStatus == ControlPanelNone {
		controlPanel.assessRotationControl(segmentCount, currentTime)
	} else if controlPanel.ControlPanelStatus == ControlPanelRotation && stage3AtCapacity {
		controlPanel.assessPositionControl(currentTime)
	} else {
		controlPanel.ControlPanelLightState = ControlPanelLightOff
	}
}

// Returns the target color for position control, assigning it randomly if it is not yet designated.
func (controlPanel *ControlPanel) GetPositionControlTargetColor() ControlPanelColor {
	if controlPanel.positionTargetColor == ColorUnknown {
		if controlPanel.CurrentColor == ColorUnknown {
			// If the sensor or manual scorekeeping did not detect/set the current color, pick one of the four at
			// random.
			controlPanel.positionTargetColor = ControlPanelColor(rand.Intn(4) + 1)
		} else {
			// Randomly pick one of the non-current colors.
			newColor := int(controlPanel.CurrentColor) + rand.Intn(3) + 1
			if newColor > 4 {
				newColor -= 4
			}
			controlPanel.positionTargetColor = ControlPanelColor(newColor)
		}
	}
	return controlPanel.positionTargetColor
}

// Returns the string that is to be sent to the driver station for the given color.
func GetGameDataForColor(color ControlPanelColor) string {
	switch color {
	case ColorRed:
		return "R"
	case ColorBlue:
		return "B"
	case ColorGreen:
		return "G"
	case ColorYellow:
		return "Y"
	}
	return ""
}

// Updates the state of the control panel while rotation control is in the process of being performed.
func (controlPanel *ControlPanel) assessRotationControl(segmentCount int, currentTime time.Time) {
	if !controlPanel.rotationStarted {
		controlPanel.rotationStarted = true
		controlPanel.rotationStartSegmentCount = segmentCount
	}

	segmentCountDiff := int(math.Abs(float64(segmentCount - controlPanel.rotationStartSegmentCount)))
	if segmentCountDiff < rotationControlMinSegments {
		// The control panel still needs to be rotated more.
		controlPanel.ControlPanelLightState = ControlPanelLightOn
	} else if segmentCountDiff < rotationControlMaxSegments {
		// The control panel has been rotated the correct amount and needs to stop on a single color.
		if segmentCountDiff != controlPanel.lastSegmentCountDiff {
			// The control panel is still moving; reset the timer.
			controlPanel.rotationStopTime = currentTime
			controlPanel.ControlPanelLightState = ControlPanelLightFlashing
		} else if currentTime.Sub(controlPanel.rotationStopTime) < rotationControlStopDurationSec*time.Second {
			controlPanel.ControlPanelLightState = ControlPanelLightFlashing
		} else {
			// The control panel has been stopped long enough; rotation control is complete.
			controlPanel.ControlPanelStatus = ControlPanelRotation
			controlPanel.ControlPanelLightState = ControlPanelLightOff
		}
	} else {
		// The control panel has been rotated too much; reset the count.
		controlPanel.rotationStartSegmentCount = segmentCount
		controlPanel.ControlPanelLightState = ControlPanelLightOn
	}
	controlPanel.lastSegmentCountDiff = segmentCountDiff
}

// Updates the state of the control panel while position control is in the process of being performed.
func (controlPanel *ControlPanel) assessPositionControl(currentTime time.Time) {
	positionCorrect := controlPanel.CurrentColor == controlPanel.GetPositionControlTargetColor() &&
		controlPanel.CurrentColor != ColorUnknown
	if positionCorrect && !controlPanel.lastPositionCorrect {
		controlPanel.positionStopTime = currentTime
	}
	controlPanel.lastPositionCorrect = positionCorrect

	if !positionCorrect {
		controlPanel.ControlPanelLightState = ControlPanelLightOn
	} else if currentTime.Sub(controlPanel.positionStopTime) < positionControlStopMinDurationSec*time.Second {
		// The control panel is on the target color but may still be moving.
		controlPanel.ControlPanelLightState = ControlPanelLightOn
	} else if currentTime.Sub(controlPanel.positionStopTime) < positionControlStopMaxDurationSec*time.Second {
		// The control panel is stopped on the target color, but not long enough to count.
		controlPanel.ControlPanelLightState = ControlPanelLightFlashing
	} else {
		// The target color has been present for long enough; position control is complete.
		controlPanel.ControlPanelStatus = ControlPanelPosition
		controlPanel.ControlPanelLightState = ControlPanelLightOff
	}
}
