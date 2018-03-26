// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVaultNumCubes(t *testing.T) {
	vault := Vault{}
	assert.Equal(t, 0, vault.GetNumCubes())

	vault.UpdateCubes([3]bool{true, false, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	assert.Equal(t, 1, vault.GetNumCubes())

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, false, true}, [3]bool{true, false, false})
	assert.Equal(t, 2, vault.GetNumCubes())

	vault.UpdateCubes([3]bool{false, true, true}, [3]bool{false, false, true}, [3]bool{true, true, false})
	assert.Equal(t, 2, vault.GetNumCubes())

	vault.UpdateCubes([3]bool{true, true, false}, [3]bool{true, true, false}, [3]bool{true, true, true})
	assert.Equal(t, 7, vault.GetNumCubes())

	vault.UpdateCubes([3]bool{true, true, true}, [3]bool{true, true, true}, [3]bool{true, true, true})
	assert.Equal(t, 9, vault.GetNumCubes())
}

func TestVaultLevitate(t *testing.T) {
	vault := Vault{}

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, true, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, true, false, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, true, true}, [3]bool{false, false, false})
	vault.UpdateButtons(true, false, true, time.Now())
	assert.False(t, vault.LevitatePlayed)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, true, true}, [3]bool{false, false, false})
	vault.UpdateButtons(false, true, false, time.Now())
	assert.True(t, vault.LevitatePlayed)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{true, true, true}, [3]bool{false, false, false})
	vault.UpdateButtons(false, false, false, time.Now())
	assert.True(t, vault.LevitatePlayed)
}

func TestVaultForce(t *testing.T) {
	vault := Vault{alliance: blueAlliance}
	ResetPowerUps()

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(true, false, false, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	vault.UpdateCubes([3]bool{true, true, true}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, true, true, time.Now())
	assert.Nil(t, vault.ForcePowerUp)

	// Activation with one cube.
	vault.UpdateCubes([3]bool{true, false, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, blueAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.kind)
		assert.Equal(t, 1, vault.ForcePowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{alliance: redAlliance}
	ResetPowerUps()
	vault.UpdateCubes([3]bool{true, true, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(true, false, false, time.Now())
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, redAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.kind)
		assert.Equal(t, 2, vault.ForcePowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{alliance: blueAlliance}
	ResetPowerUps()
	vault.UpdateCubes([3]bool{true, true, true}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(true, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
	if assert.NotNil(t, vault.ForcePowerUp) {
		assert.Equal(t, blueAlliance, vault.ForcePowerUp.alliance)
		assert.Equal(t, force, vault.ForcePowerUp.kind)
		assert.Equal(t, 3, vault.ForcePowerUp.level)
	}

	vault.UpdateCubes([3]bool{true, true, true}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.ForcePowerUp)
}

func TestVaultBoost(t *testing.T) {
	vault := Vault{alliance: blueAlliance}
	ResetPowerUps()

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{false, false, false})
	vault.UpdateButtons(false, false, true, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{true, true, true})
	vault.UpdateButtons(true, true, false, time.Now())
	assert.Nil(t, vault.BoostPowerUp)

	// Activation with one cube.
	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{true, false, false})
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, blueAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.kind)
		assert.Equal(t, 1, vault.BoostPowerUp.level)
	}

	// Activation with two cubes.
	vault = Vault{alliance: redAlliance}
	ResetPowerUps()
	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{true, true, false})
	vault.UpdateButtons(false, false, true, time.Now())
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, redAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.kind)
		assert.Equal(t, 2, vault.BoostPowerUp.level)
	}

	// Activation with three cubes.
	vault = Vault{alliance: blueAlliance}
	ResetPowerUps()
	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{true, true, true})
	vault.UpdateButtons(false, false, true, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
	if assert.NotNil(t, vault.BoostPowerUp) {
		assert.Equal(t, blueAlliance, vault.BoostPowerUp.alliance)
		assert.Equal(t, boost, vault.BoostPowerUp.kind)
		assert.Equal(t, 3, vault.BoostPowerUp.level)
	}

	vault.UpdateCubes([3]bool{false, false, false}, [3]bool{false, false, false}, [3]bool{true, true, true})
	vault.UpdateButtons(false, false, false, time.Now())
	assert.NotNil(t, vault.BoostPowerUp)
}
