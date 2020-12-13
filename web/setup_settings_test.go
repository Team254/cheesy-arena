// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"bytes"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupSettings(t *testing.T) {
	web := setupTestWeb(t)

	// Check the default setting values.
	recorder := web.getHttpResponse("/setup/settings")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Untitled Event")
	assert.Contains(t, recorder.Body.String(), "8")
	assert.NotContains(t, recorder.Body.String(), "tbaPublishingEnabled\" checked")

	// Change the settings and check the response.
	recorder = web.postHttpResponse("/setup/settings", "name=Chezy Champs&code=CC&numElimAlliances=16&"+
		"tbaPublishingEnabled=on&tbaEventCode=2014cc&tbaSecretId=secretId&tbaSecret=tbasec")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/settings")
	assert.Contains(t, recorder.Body.String(), "Chezy Champs")
	assert.Contains(t, recorder.Body.String(), "16")
	assert.Contains(t, recorder.Body.String(), "tbaPublishingEnabled\" checked")
	assert.Contains(t, recorder.Body.String(), "2014cc")
	assert.Contains(t, recorder.Body.String(), "secretId")
	assert.Contains(t, recorder.Body.String(), "tbasec")
}

func TestSetupSettingsInvalidValues(t *testing.T) {
	web := setupTestWeb(t)

	// Invalid number of alliances.
	recorder := web.postHttpResponse("/setup/settings", "numAlliances=1")
	assert.Contains(t, recorder.Body.String(), "must be between 2 and 16")
}

func TestSetupSettingsClearDb(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateTeam(new(model.Team))
	web.arena.Database.CreateMatch(&model.Match{Type: "qualification"})
	web.arena.Database.CreateMatchResult(new(model.MatchResult))
	web.arena.Database.CreateRanking(new(game.Ranking))
	web.arena.Database.CreateAllianceTeam(new(model.AllianceTeam))
	recorder := web.postHttpResponse("/setup/db/clear", "")
	assert.Equal(t, 303, recorder.Code)

	teams, _ := web.arena.Database.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ := web.arena.Database.GetMatchesByType("qualification")
	assert.Empty(t, matches)
	rankings, _ := web.arena.Database.GetAllRankings()
	assert.Empty(t, rankings)
	tournament.CalculateRankings(web.arena.Database, false)
	assert.Empty(t, rankings)
	alliances, _ := web.arena.Database.GetAllAlliances()
	assert.Empty(t, alliances)
}

func TestSetupSettingsBackupRestoreDb(t *testing.T) {
	web := setupTestWeb(t)

	// Modify a parameter so that we know when the database has been restored.
	web.arena.EventSettings.Name = "Chezy Champs"
	web.arena.Database.SaveEventSettings(web.arena.EventSettings)

	// Back up the database.
	recorder := web.getHttpResponse("/setup/db/save")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/octet-stream", recorder.HeaderMap["Content-Type"][0])
	backupBody := recorder.Body

	// Wipe the database to reset the defaults.
	web = setupTestWeb(t)
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with a missing file.
	recorder = web.postHttpResponse("/setup/db/restore", "")
	assert.Contains(t, recorder.Body.String(), "No database backup file was specified")
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with a corrupt file.
	recorder = web.postFileHttpResponse("/setup/db/restore", "databaseFile",
		bytes.NewBufferString("invalid"))
	assert.Contains(t, recorder.Body.String(), "Could not read uploaded database backup file")
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with the backup retrieved before.
	recorder = web.postFileHttpResponse("/setup/db/restore", "databaseFile", backupBody)
	assert.Equal(t, "Chezy Champs", web.arena.EventSettings.Name)

}

func (web *Web) postFileHttpResponse(path string, paramName string, file *bytes.Buffer) *httptest.ResponseRecorder {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(paramName, "file.ext")
	io.Copy(part, file)
	writer.Close()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	web.newHandler().ServeHTTP(recorder, req)
	return recorder
}
