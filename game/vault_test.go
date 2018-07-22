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
	vault := Vault{Alliance: BlueAlliance}
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
		assert.Equal(t, BlueAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 1, vault.ForcePowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{Alliance: RedAlliance}
	ResetPowerUps()
	vault.UpdateCubes(2000, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, RedAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 2, vault.ForcePowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{Alliance: BlueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(3000, 0, 0)
	vault.UpdateButtons(true, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, BlueAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, force, vault.ForcePowerUp.effect)
		assert.Equal(t, 3, vault.ForcePowerUp.level)
	}

	vault.UpdateCubes(3000, 0, 0)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
}

func TestVaultBoost(t *testing.T) {
	vault := Vault{Alliance: BlueAlliance}
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
		assert.Equal(t, BlueAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 1, vault.BoostPowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{Alliance: RedAlliance}
	ResetPowerUps()
	vault.UpdateCubes(0, 0, 2000)
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, RedAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 2, vault.BoostPowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{Alliance: BlueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(0, 0, 3000)
	vault.UpdateButtons(false, false, true, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, BlueAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.effect)
		assert.Equal(t, 3, vault.BoostPowerUp.level)
	}

	vault.UpdateCubes(0, 0, 3000)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
}

func TestVaultMultipleActivations(t *testing.T) {
	redVault := Vault{Alliance: RedAlliance}
	redVault.UpdateCubes(1000, 3000, 1000)
	blueVault := Vault{Alliance: BlueAlliance}
	blueVault.UpdateCubes(1000, 3000, 1000)
	ResetPowerUps()

	redVault.UpdateButtons(true, false, false, timeAfterStart(0))
	redVault.UpdateButtons(false, false, false, timeAfterStart(1))
	if assert.NotNil(t, redVault.ForcePowerUp) {
		assert.Equal(t, Active, redVault.ForcePowerUp.GetState(timeAfterStart(0.5)))
	}
	assert.Equal(t, "force", redVault.CheckForNewlyPlayedPowerUp())
	assert.Equal(t, "", redVault.CheckForNewlyPlayedPowerUp())

	redVault.UpdateButtons(false, true, false, timeAfterStart(2))
	redVault.UpdateButtons(false, false, false, timeAfterStart(3))
	assert.True(t, redVault.LevitatePlayed)
	assert.Equal(t, "levitate", redVault.CheckForNewlyPlayedPowerUp())
	assert.Equal(t, "", redVault.CheckForNewlyPlayedPowerUp())

	blueVault.UpdateButtons(false, false, true, timeAfterStart(4))
	blueVault.UpdateButtons(false, false, false, timeAfterStart(5))
	if assert.NotNil(t, blueVault.BoostPowerUp) {
		assert.Equal(t, Queued, blueVault.BoostPowerUp.GetState(timeAfterStart(4.5)))
	}
	assert.Equal(t, "boost", blueVault.CheckForNewlyPlayedPowerUp())
	assert.Equal(t, "", blueVault.CheckForNewlyPlayedPowerUp())
	assert.Equal(t, Expired, redVault.ForcePowerUp.GetState(timeAfterStart(11)))
	assert.Equal(t, Active, blueVault.BoostPowerUp.GetState(timeAfterStart(11)))
	assert.Equal(t, Expired, blueVault.BoostPowerUp.GetState(timeAfterStart(21)))

	redVault.UpdateButtons(false, false, true, timeAfterStart(25))
	redVault.UpdateButtons(false, false, false, timeAfterStart(26))
	if assert.NotNil(t, redVault.BoostPowerUp) {
		assert.Equal(t, Active, redVault.BoostPowerUp.GetState(timeAfterStart(25.5)))
	}
	assert.Equal(t, "boost", redVault.CheckForNewlyPlayedPowerUp())
	assert.Equal(t, "", redVault.CheckForNewlyPlayedPowerUp())
}
