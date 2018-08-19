// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains the set of display modes for the vault LEDs.

package vaultled

type Mode int

const (
	OffMode Mode = iota
	OneCubeMode
	TwoCubeMode
	ThreeCubeMode
	RedPlayedMode
	BluePlayedMode
)

var ModeNames = map[Mode]string{
	OffMode:        "Off",
	OneCubeMode:    "One Cube",
	TwoCubeMode:    "Two Cubes",
	ThreeCubeMode:  "Three Cubes",
	RedPlayedMode:  "Red Played",
	BluePlayedMode: "Blue Played",
}
