// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
	assert.NotContains(t, recorder.Body.String(), "tbaPublishingEnabled\" checked")

	// Change the settings and check the response.
	recorder = postHttpResponse("/setup/settings", "name=Chezy Champs&code=CC&displayBackgroundColor=#ff00ff&"+
		"numElimAlliances=16&tbaPublishingEnabled=on&tbaEventCode=2014cc&tbaSecretId=secretId&tbaSecret=tbasec&"+
		"initialTowerStrength=9001")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/setup/settings")
	assert.Contains(t, recorder.Body.String(), "Chezy Champs")
	assert.Contains(t, recorder.Body.String(), "CC")
	assert.Contains(t, recorder.Body.String(), "#ff00ff")
	assert.Contains(t, recorder.Body.String(), "16")
	assert.Contains(t, recorder.Body.String(), "tbaPublishingEnabled\" checked")
	assert.Contains(t, recorder.Body.String(), "2014cc")
	assert.Contains(t, recorder.Body.String(), "secretId")
	assert.Contains(t, recorder.Body.String(), "tbasec")
	assert.Contains(t, recorder.Body.String(), "9001")
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

func TestSetupSettingsClearDb(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	db.CreateTeam(new(Team))
	db.CreateMatch(&Match{Type: "qualification"})
	db.CreateMatchResult(new(MatchResult))
	db.CreateRanking(new(Ranking))
	db.CreateAllianceTeam(new(AllianceTeam))
	recorder := postHttpResponse("/setup/db/clear", "")
	assert.Equal(t, 302, recorder.Code)

	teams, _ := db.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ := db.GetMatchesByType("qualification")
	assert.Empty(t, matches)
	rankings, _ := db.GetAllRankings()
	assert.Empty(t, rankings)
	db.CalculateRankings()
	assert.Empty(t, rankings)
	alliances, _ := db.GetAllAlliances()
	assert.Empty(t, alliances)
}

func TestSetupSettingsBackupRestoreDb(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Modify a parameter so that we know when the database has been restored.
	eventSettings.Name = "Chezy Champs"
	db.SaveEventSettings(eventSettings)

	// Back up the database.
	recorder := getHttpResponse("/setup/db/save")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/octet-stream", recorder.HeaderMap["Content-Type"][0])
	backupBody := recorder.Body

	// Wipe the database to reset the defaults.
	clearDb()
	defer clearDb()
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	assert.NotEqual(t, "Chezy Champs", eventSettings.Name)

	// Check restoring with a missing file.
	recorder = postHttpResponse("/setup/db/restore", "")
	assert.Contains(t, recorder.Body.String(), "No database backup file was specified")
	assert.NotEqual(t, "Chezy Champs", eventSettings.Name)

	// Check restoring with a corrupt file.
	recorder = postFileHttpResponse("/setup/db/restore", "databaseFile", bytes.NewBufferString("invalid"))
	assert.Contains(t, recorder.Body.String(), "Could not read uploaded database backup file")
	assert.NotEqual(t, "Chezy Champs", eventSettings.Name)

	// Check restoring with the backup retrieved before.
	recorder = postFileHttpResponse("/setup/db/restore", "databaseFile", backupBody)
	fmt.Println(recorder.Body.String())
	assert.Equal(t, "Chezy Champs", eventSettings.Name)
}

func postFileHttpResponse(path string, paramName string, file *bytes.Buffer) *httptest.ResponseRecorder {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(paramName, "file.ext")
	io.Copy(part, file)
	writer.Close()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	newHandler().ServeHTTP(recorder, req)
	return recorder
}
