// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2018 vault element.

package game

import (
	"time"
)

type Vault struct {
	alliance         int
	numForceCubes    int
	numLevitateCubes int
	numBoostCubes    int
	LevitatePlayed   bool
	ForcePowerUp     *PowerUp
	BoostPowerUp     *PowerUp
}

// Updates the state of the vault given the state of the individual power cube sensors.
func (vault *Vault) UpdateCubes(forceDistance, levitateDistance, boostDistance uint16) {
	vault.numForceCubes = countCubes(forceDistance)
	vault.numLevitateCubes = countCubes(levitateDistance)
	vault.numBoostCubes = countCubes(boostDistance)
}

// Updates the state of the vault given the state of the power up buttons.
func (vault *Vault) UpdateButtons(forceButton, levitateButton, boostButton bool, currentTime time.Time) {
	if levitateButton && vault.numLevitateCubes == 3 && !vault.LevitatePlayed {
		vault.LevitatePlayed = true
	}

	if forceButton && vault.numForceCubes > 0 && vault.ForcePowerUp == nil {
		vault.ForcePowerUp = maybeActivatePowerUp(&PowerUp{kind: force, alliance: vault.alliance,
			level: vault.numForceCubes}, currentTime)
	}

	if boostButton && vault.numBoostCubes > 0 && vault.BoostPowerUp == nil {
		vault.BoostPowerUp = maybeActivatePowerUp(&PowerUp{kind: boost, alliance: vault.alliance,
			level: vault.numBoostCubes}, currentTime)
	}
}

// Returns the total count of power cubes that have been placed in the vault.
func (vault *Vault) GetNumCubes() int {
	return vault.numForceCubes + vault.numLevitateCubes + vault.numBoostCubes
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
