// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package model

import (
	"os"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
)

func TestEventSettingsDefaults(t *testing.T) {
	// Use temporary database
	os.Remove("test_settings.db")
	db, err := OpenDatabase("test_settings.db")
	assert.Nil(t, err)
	defer os.Remove("test_settings.db")

	// 1. Test: Get settings (should automatically create default values if empty)
	settings, err := db.GetEventSettings()
	assert.Nil(t, err)

	// Verify that the 2026 default values are loaded (from the game package)
	assert.Equal(t, game.EnergizedFuelThreshold, settings.EnergizedFuelThreshold)
	assert.Equal(t, game.SuperchargedFuelThreshold, settings.SuperchargedFuelThreshold)
	assert.Equal(t, game.TraversalPointThreshold, settings.TraversalPointThreshold)
}

func TestUpdateEventSettings(t *testing.T) {
	os.Remove("test_settings_update.db")
	db, err := OpenDatabase("test_settings_update.db")
	assert.Nil(t, err)
	defer os.Remove("test_settings_update.db")

	settings, _ := db.GetEventSettings()

	// 2. Test: Modify and save settings
	settings.Name = "2026 Championship"
	settings.EnergizedFuelThreshold = 999 // Modify RP threshold
	settings.SuperchargedFuelThreshold = 1000

	err = db.UpdateEventSettings(settings)
	assert.Nil(t, err)

	// Reload and verify
	newSettings, _ := db.GetEventSettings()
	assert.Equal(t, "2026 Championship", newSettings.Name)
	assert.Equal(t, 999, newSettings.EnergizedFuelThreshold)
	assert.Equal(t, 1000, newSettings.SuperchargedFuelThreshold)
}
