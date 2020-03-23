// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents the state of an alliance Control Panel in the 2020 game.

package game

import "math/rand"

type ControlPanel struct {
	CurrentColor ControlPanelColor
}

type ControlPanelColor int

const (
	ColorUnknown ControlPanelColor = iota
	ColorRed
	ColorGreen
	ColorBlue
	ColorYellow
)

type ControlPanelStatus int

const (
	ControlPanelNone ControlPanelStatus = iota
	ControlPanelRotation
	ControlPanelPosition
)

// Returns a random color that does not match the current color.
func (controlPanel *ControlPanel) GetStage3TargetColor() ControlPanelColor {
	if controlPanel.CurrentColor == ColorUnknown {
		// If the sensor or manual scorekeeping did not detect/set the current color, pick one of the four at random.
		return ControlPanelColor(rand.Intn(4) + 1)
	}
	newColor := int(controlPanel.CurrentColor) + rand.Intn(3) + 1
	if newColor > 4 {
		newColor -= 4
	}
	return ControlPanelColor(newColor)
}

// Returns the string that is to be sent to the driver station for the given color.
func GetGameDataForColor(color ControlPanelColor) string {
	switch color {
	case ColorRed:
		return "R"
	case ColorGreen:
		return "G"
	case ColorBlue:
		return "B"
	case ColorYellow:
		return "Y"
	}
	return ""
}
