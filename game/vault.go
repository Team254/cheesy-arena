// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2018 vault element.

package game

import (
	"time"
)

type Vault struct {
	Alliance
	ForceCubes         int
	ForceCubesPlayed   int
	LevitateCubes      int
	LevitatePlayed     bool
	BoostCubes         int
	BoostCubesPlayed   int
	ForcePowerUp       *PowerUp
	BoostPowerUp       *PowerUp
	newlyPlayedPowerUp string
}

// Updates the state of the vault given the state of the individual power cube sensors.
func (vault *Vault) UpdateCubes(forceDistance, levitateDistance, boostDistance uint16) {
	vault.ForceCubes = countCubes(forceDistance)
	vault.LevitateCubes = countCubes(levitateDistance)
	vault.BoostCubes = countCubes(boostDistance)
}

// Updates the state of the vault given the state of the power up buttons.
func (vault *Vault) UpdateButtons(forceButton, levitateButton, boostButton bool, currentTime time.Time) {
	if levitateButton && vault.LevitateCubes == 3 && !vault.LevitatePlayed {
		vault.LevitatePlayed = true
		vault.newlyPlayedPowerUp = "levitate"
	}

	if forceButton && vault.ForceCubes > 0 && vault.ForcePowerUp == nil {
		vault.ForcePowerUp = maybeActivatePowerUp(&PowerUp{Effect: Force, Alliance: vault.Alliance,
			Level: vault.ForceCubes}, currentTime)
		if vault.ForcePowerUp != nil {
			vault.ForceCubesPlayed = vault.ForceCubes
			vault.newlyPlayedPowerUp = "force"
		}
	}

	if boostButton && vault.BoostCubes > 0 && vault.BoostPowerUp == nil {
		vault.BoostPowerUp = maybeActivatePowerUp(&PowerUp{Effect: Boost, Alliance: vault.Alliance,
			Level: vault.BoostCubes}, currentTime)
		if vault.BoostPowerUp != nil {
			vault.BoostCubesPlayed = vault.BoostCubes
			vault.newlyPlayedPowerUp = "boost"
		}
	}
}

// CheckForNewlyPlayedPowerUp returns the name of the newly-played power up if there is one, or an empty string otherwise, and resets the state.
func (vault *Vault) CheckForNewlyPlayedPowerUp() string {
	powerUp := vault.newlyPlayedPowerUp
	vault.newlyPlayedPowerUp = ""
	return powerUp
}

func countCubes(distance uint16) int {
	// Ed Jordan's measurements:
	//   Empty    125
	//   1 Short  98
	//   1 Tall   92
	//   2 Short  68
	//   2 Tall   58
	//   3 Short  43
	//   3 Tall   26
	if distance <= 15 {
		// The sensor is probably disconnected or obstructed; this is too tall to be a cube stack.
		return 0
	}
	if distance <= 50 {
		return 3
	}
	if distance <= 75 {
		return 2
	}
	if distance <= 105 {
		return 1
	}
	return 0
}
