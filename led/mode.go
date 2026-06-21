// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains the set of display modes for the hub LEDs.

package led

type Mode int

const (
	OffMode Mode = iota
	RedMode
	BlueMode
	GreenMode
	PurpleMode
	WhiteMode
	RedPulseMode
	BluePulseMode
	RedStartupMode
	BlueStartupMode
	RedAdvantageMode
	BlueAdvantageMode
	RainbowMode
)

var ModeNames = map[Mode]string{
	OffMode:           "Off",
	RedMode:           "Red",
	BlueMode:          "Blue",
	GreenMode:         "Field Safe",
	PurpleMode:        "Field Cleanup",
	WhiteMode:         "Scoring Assessment",
	RedPulseMode:      "Red Pulse",
	BluePulseMode:     "Blue Pulse",
	RedStartupMode:    "Red Startup",
	BlueStartupMode:   "Blue Startup",
	RedAdvantageMode:  "Red Advantage",
	BlueAdvantageMode: "Blue Advantage",
	RainbowMode:       "Rainbow",
}

// Returns the solid color associated with the given mode.
func colorForMode(mode Mode) Color {
	switch mode {
	case RedMode:
		return Red
	case BlueMode:
		return Blue
	case GreenMode:
		return Green
	case PurpleMode:
		return Purple
	case WhiteMode:
		return White
	default:
		return Black
	}
}
