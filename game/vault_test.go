// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	zeroCubeDistance  = 125
	oneCubeDistance   = 98
	twoCubeDistance   = 58
	threeCubeDistance = 43
)

func TestVaultNumCubes(t *testing.T) {
	// TODO(patrick): Update with real values once there is a physical setup to test with.
	vault := Vault{}
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 0, vault.BoostCubes)

	vault.UpdateCubes(oneCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 1, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 0, vault.BoostCubes)

	vault.UpdateCubes(zeroCubeDistance, oneCubeDistance, oneCubeDistance)
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 1, vault.LevitateCubes)
	assert.Equal(t, 1, vault.BoostCubes)

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, twoCubeDistance)
	assert.Equal(t, 0, vault.ForceCubes)
	assert.Equal(t, 0, vault.LevitateCubes)
	assert.Equal(t, 2, vault.BoostCubes)

	vault.UpdateCubes(twoCubeDistance, twoCubeDistance, threeCubeDistance)
	assert.Equal(t, 2, vault.ForceCubes)
	assert.Equal(t, 2, vault.LevitateCubes)
	assert.Equal(t, 3, vault.BoostCubes)

	vault.UpdateCubes(threeCubeDistance, threeCubeDistance, threeCubeDistance)
	assert.Equal(t, 3, vault.ForceCubes)
	assert.Equal(t, 3, vault.LevitateCubes)
	assert.Equal(t, 3, vault.BoostCubes)

	assert.Equal(t, 0, vault.ForceCubesPlayed)
	assert.Equal(t, 0, vault.BoostCubesPlayed)
}

func TestVaultLevitate(t *testing.T) {
	vault := Vault{}

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(zeroCubeDistance, oneCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(zeroCubeDistance, twoCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(zeroCubeDistance, threeCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(true, false, true, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes(zeroCubeDistance, threeCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, true, false, time.Now())
	assert.True(t, vault.LevitatePlayed)

	vault.UpdateCubes(zeroCubeDistance, threeCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.True(t, vault.LevitatePlayed)
}

func TestVaultForce(t *testing.T) {
	vault := Vault{Alliance: BlueAlliance}
	ResetPowerUps()

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(true, false, false, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	vault.UpdateCubes(threeCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, true, true, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	// Activation with one cube.
	vault.UpdateCubes(oneCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, BlueAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, Force, vault.ForcePowerUp.Effect)
		assert.Equal(t, 1, vault.ForcePowerUp.Level)
	}
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 1, vault.ForceCubesPlayed)

	// Activation with two cubes.
	vault = Vault{Alliance: RedAlliance}
	ResetPowerUps()
	vault.UpdateCubes(twoCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, RedAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, Force, vault.ForcePowerUp.Effect)
		assert.Equal(t, 2, vault.ForcePowerUp.Level)
		assert.Equal(t, 2, vault.ForceCubesPlayed)
	}
	vault.UpdateCubes(threeCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 2, vault.ForceCubesPlayed)

	// Activation with three cubes.
	vault = Vault{Alliance: BlueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(threeCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(true, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, BlueAlliance, vault.ForcePowerUp.Alliance)
		assert.Equal(t, Force, vault.ForcePowerUp.Effect)
		assert.Equal(t, 3, vault.ForcePowerUp.Level)
	}
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 3, vault.ForceCubesPlayed)

	vault.UpdateCubes(threeCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
}

func TestVaultBoost(t *testing.T) {
	vault := Vault{Alliance: BlueAlliance}
	ResetPowerUps()

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	vault.UpdateButtons(false, false, true, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, threeCubeDistance)
	vault.UpdateButtons(true, true, false, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	// Activation with one cube.
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, oneCubeDistance)
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, BlueAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, Boost, vault.BoostPowerUp.Effect)
		assert.Equal(t, 1, vault.BoostPowerUp.Level)
	}
	vault.UpdateCubes(zeroCubeDistance, twoCubeDistance, zeroCubeDistance)
	assert.Equal(t, 1, vault.BoostCubesPlayed)

	// Activation with two cubes.
	vault = Vault{Alliance: RedAlliance}
	ResetPowerUps()
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, twoCubeDistance)
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, RedAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, Boost, vault.BoostPowerUp.Effect)
		assert.Equal(t, 2, vault.BoostPowerUp.Level)
	}
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 2, vault.BoostCubesPlayed)

	// Activation with three cubes.
	vault = Vault{Alliance: BlueAlliance}
	ResetPowerUps()
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, threeCubeDistance)
	vault.UpdateButtons(false, false, true, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, BlueAlliance, vault.BoostPowerUp.Alliance)
		assert.Equal(t, Boost, vault.BoostPowerUp.Effect)
		assert.Equal(t, 3, vault.BoostPowerUp.Level)
	}
	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, zeroCubeDistance)
	assert.Equal(t, 3, vault.BoostCubesPlayed)

	vault.UpdateCubes(zeroCubeDistance, zeroCubeDistance, threeCubeDistance)
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
}

func TestVaultMultipleActivations(t *testing.T) {
	redVault := Vault{Alliance: RedAlliance}
	redVault.UpdateCubes(oneCubeDistance, threeCubeDistance, oneCubeDistance)
	blueVault := Vault{Alliance: BlueAlliance}
	blueVault.UpdateCubes(oneCubeDistance, threeCubeDistance, oneCubeDistance)
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
