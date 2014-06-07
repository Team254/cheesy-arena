// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventSettingsReadWrite(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	eventSettings, err := db.GetEventSettings()
	assert.Nil(t, err)
	assert.Equal(t, EventSettings{0, "Untitled Event", "UE", "#00ff00", 8, "F", "L", ""}, *eventSettings)

	eventSettings.Name = "Chezy Champs"
	eventSettings.Code = "cc"
	eventSettings.DisplayBackgroundColor = "#ff00ff"
	eventSettings.NumElimAlliances = 6
	eventSettings.SelectionRound1Order = "F"
	eventSettings.SelectionRound2Order = "F"
	eventSettings.SelectionRound3Order = "L"
	err = db.SaveEventSettings(eventSettings)
	assert.Nil(t, err)
	eventSettings2, err := db.GetEventSettings()
	assert.Nil(t, err)
	assert.Equal(t, *eventSettings, *eventSettings2)
}
