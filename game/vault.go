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
	LevitateCubes      int
	BoostCubes         int
	LevitatePlayed     bool
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
		vault.ForcePowerUp = maybeActivatePowerUp(&PowerUp{effect: force, Alliance: vault.Alliance,
			level: vault.ForceCubes}, currentTime)
		if vault.ForcePowerUp != nil {
			vault.newlyPlayedPowerUp = "force"
		}
	}

	if boostButton && vault.BoostCubes > 0 && vault.BoostPowerUp == nil {
		vault.BoostPowerUp = maybeActivatePowerUp(&PowerUp{effect: boost, Alliance: vault.Alliance,
			level: vault.BoostCubes}, currentTime)
		if vault.BoostPowerUp != nil {
			vault.newlyPlayedPowerUp = "boost"
		}
	}
}

// Returns the name of the newly-played power up if there is one, or an empty string otherwise, and resets the state.
func (vault *Vault) CheckForNewlyPlayedPowerUp() string {
	powerUp := vault.newlyPlayedPowerUp
	vault.newlyPlayedPowerUp = ""
	return powerUp
}

func countCubes(distance uint16) int {
	// TODO(patrick): Update with real values once there is a physical setup to test with.
	if distance >= 3000 {
		return 3
	}
	if distance >= 2000 {
		return 2
	}
	if distance >= 1000 {
		return 1
	}
	return 0
}
