// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains the set of display modes for the vault LEDs.

package led

type VaultMode int

const (
	VaultOffMode VaultMode = iota
	OneCubeMode
	TwoCubeMode
	ThreeCubeMode
	RedPlayedMode
	BluePlayedMode
)

var VaultModeNames = map[VaultMode]string{
	VaultOffMode:   "Off",
	OneCubeMode:    "One Cube",
	TwoCubeMode:    "Two Cubes",
	ThreeCubeMode:  "Three Cubes",
	RedPlayedMode:  "Red Played",
	BluePlayedMode: "Blue Played",
}
