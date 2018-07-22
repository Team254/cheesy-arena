// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains the set of display modes for an LED strip.

package led

type Mode int

const (
	OffMode Mode = iota
	RedMode
	GreenMode
	BlueMode
	WhiteMode
	ChaseMode
	WarmupMode
	Warmup2Mode
	Warmup3Mode
	Warmup4Mode
	OwnedMode
	NotOwnedMode
	ForceMode
	BoostMode
	RandomMode
	FadeRedBlueMode
	FadeSingleMode
	GradientMode
	BlinkMode
)

var ModeNames = map[Mode]string{
	OffMode:         "Off",
	RedMode:         "Red",
	GreenMode:       "Green",
	BlueMode:        "Blue",
	WhiteMode:       "White",
	ChaseMode:       "Chase",
	WarmupMode:      "Warmup",
	Warmup2Mode:     "Warmup Purple",
	Warmup3Mode:     "Warmup Sneaky",
	Warmup4Mode:     "Warmup Gradient",
	OwnedMode:       "Owned",
	NotOwnedMode:    "Not Owned",
	ForceMode:       "Force",
	BoostMode:       "Boost",
	RandomMode:      "Random",
	FadeRedBlueMode: "Fade Red/Blue",
	FadeSingleMode:  "Fade Single",
	GradientMode:    "Gradient",
	BlinkMode:       "Blink",
}
