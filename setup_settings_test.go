// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupSettings(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Check the default setting values.
	recorder := getHttpResponse("/setup/settings")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Untitled Event")
	assert.Contains(t, recorder.Body.String(), "UE")
	assert.Contains(t, recorder.Body.String(), "#00ff00")
	assert.Contains(t, recorder.Body.String(), "8")

	// Change the settings and check the response.
	recorder = postHttpResponse("/setup/settings", "name=Chezy Champs&code=CC&displayBackgroundColor=#ff00ff&"+
		"numElimAlliances=16")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Chezy Champs")
	assert.Contains(t, recorder.Body.String(), "CC")
	assert.Contains(t, recorder.Body.String(), "#ff00ff")
	assert.Contains(t, recorder.Body.String(), "16")
}

func TestSetupSettingsInvalidValues(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Invalid color value.
	recorder := postHttpResponse("/setup/settings", "numAlliances=8&displayBackgroundColor=blorpy")
	assert.Contains(t, recorder.Body.String(), "must be a valid hex color value")

	// Invalid number of alliances.
	recorder = postHttpResponse("/setup/settings", "numAlliances=1&displayBackgroundColor=#000")
	assert.Contains(t, recorder.Body.String(), "must be between 2 and 16")
}
