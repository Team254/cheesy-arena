// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains the set of display modes for an LED strip.

package led

type StripMode int

const (
	OffMode StripMode = iota
	RedMode
	GreenMode
	BlueMode
	WhiteMode
	PurpleMode
	ChaseMode
	RandomMode
	FadeRedBlueMode
	FadeSingleMode
	GradientMode
	BlinkMode
)

var StripModeNames = map[StripMode]string{
	OffMode:         "Off",
	RedMode:         "Red",
	GreenMode:       "Green",
	BlueMode:        "Blue",
	WhiteMode:       "White",
	PurpleMode:      "Purple",
	ChaseMode:       "Chase",
	RandomMode:      "Random",
	FadeRedBlueMode: "Fade Red/Blue",
	FadeSingleMode:  "Fade Single",
	GradientMode:    "Gradient",
	BlinkMode:       "Blink",
}
