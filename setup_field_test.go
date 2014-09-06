// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupField(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	mainArena.allianceStationDisplays["12345"] = ""
	recorder := getHttpResponse("/setup/field")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "12345")
	assert.NotContains(t, recorder.Body.String(), "selected")

	recorder = postHttpResponse("/setup/field", "displayId=12345&allianceStation=B1")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/setup/field")
	assert.Contains(t, recorder.Body.String(), "12345")
	assert.Contains(t, recorder.Body.String(), "selected")

	recorder = postHttpResponse("/setup/field/lights", "mode=strobe")
	assert.Equal(t, 302, recorder.Code)
	assert.Equal(t, "strobe", mainArena.lights.currentMode)
}
