// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVaultNumCubes(t *testing.T) {
	// TODO(patrick): Update with real values once there is a physical setup to test with.
	vault := Vault{}
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 0, vault.BoostCubes)

	vault.UpdateCubes(1000, 0, 0)
	assert.Equal(t, 1, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 0, vault.BoostCubes)

	vault.UpdateCubes(0, 1000, 1000)
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 1, vault.LevitateCubes)
	assert.Equal(t, 1, vault.BoostCubes)

	vault.UpdateCubes(0, 0, 2000)
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 2, vault.BoostCubes)

	vault.UpdateCubes(2000, 2000, 3000)
	assert.Equal(t, 2, vault.ForceCubes)
	assert.Equal(t, 2, vault.LevitateCubes)
	assert.Equal(t, 3, vault.BoostCubes)

	vault.UpdateCubes(3000, 3000, 3000)
	assert.Equal(t, 3, vault.ForceCubes)
	assert.Equal(t, 3, vault.LevitateCubes)
	assert.Equal(t, 3, vault.BoostCubes)
}

func TestVaultLevitate(t *testing.T) {
	vault := Vault{}

	vault.UpdateCubes(0, 0, 0)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(0, 1000, 0)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(0, 2000, 0)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(0, 3000, 0)
	vault.UpdateButtons(true, false, true, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(0, 3000, 0)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.True(t, vault.LevitatePlayed)

	vault.UpdateCubes(0, 3000, 0)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.True(t, vault.LevitatePlayed)
}

func TestVaultForce(t *testing.T) {
	vault := Vault{alliance: blueAlliance}
	ResetPowerUps()

	vault.UpdateCubes(0, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	vault.UpdateCubes(3000, 0, 0)
	vault.UpdateButtons(false, true, true, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	// Activation with one cube.
	vault.UpdateCubes(1000, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, blueAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 1, vault.ForcePowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{alliance: redAlliance}
	ResetPowerUps()
	vault.UpdateCubes(2000, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, redAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 2, vault.ForcePowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{alliance: blueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(3000, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, blueAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 3, vault.ForcePowerUp.level)
	}

	vault.UpdateCubes(3000, 0, 0)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
}

func TestVaultBoost(t *testing.T) {
	vault := Vault{alliance: blueAlliance}
	ResetPowerUps()

	vault.UpdateCubes(0, 0, 0)
	vault.UpdateButtons(false, false, true, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	vault.UpdateCubes(0, 0, 3000)
	vault.UpdateButtons(true, true, false, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	// Activation with one cube.
	vault.UpdateCubes(0, 0, 1000)
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, blueAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 1, vault.BoostPowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{alliance: redAlliance}
	ResetPowerUps()
	vault.UpdateCubes(0, 0, 2000)
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, redAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 2, vault.BoostPowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{alliance: blueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(0, 0, 3000)
	vault.UpdateButtons(false, false, true, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, blueAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 3, vault.BoostPowerUp.level)
	}

	vault.UpdateCubes(0, 0, 3000)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
}
