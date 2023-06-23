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
	recorder = web.postHttpResponse("/setup/settings", "name=Chezy Champs&code=CC&playoffType=single&"+
		"numPlayoffAlliances=16&tbaPublishingEnabled=on&tbaEventCode=2014cc&tbaSecretId=secretId&tbaSecret=tbasec")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/settings")
	assert.Contains(t, recorder.Body.String(), "Chezy Champs")
	assert.Contains(t, recorder.Body.String(), "16")
	assert.Contains(t, recorder.Body.String(), "tbaPublishingEnabled\" checked")
	assert.Contains(t, recorder.Body.String(), "2014cc")
	assert.Contains(t, recorder.Body.String(), "secretId")
	assert.Contains(t, recorder.Body.String(), "tbasec")
}

func TestSetupSettingsDoubleElimination(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.postHttpResponse("/setup/settings", "playoffType=DoubleEliminationPlayoff&numPlayoffAlliances=3")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, model.DoubleEliminationPlayoff, web.arena.EventSettings.PlayoffType)
	assert.Equal(t, 8, web.arena.EventSettings.NumPlayoffAlliances)
}

func TestSetupSettingsInvalidValues(t *testing.T) {
	web := setupTestWeb(t)

	// Invalid number of alliances.
	recorder := web.postHttpResponse("/setup/settings", "playoffType=SingleEliminationPlayoff&numAlliances=1")
	assert.Contains(t, recorder.Body.String(), "must be between 2 and 16")

	// Changing the playoff type after alliance selection is finalized.
	assert.Nil(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1}))
	recorder = web.postHttpResponse("/setup/settings", "playoffType=SingleEliminationPlayoff&numPlayoffAlliances=8")
	assert.Contains(t, recorder.Body.String(), "Cannot change playoff type after alliance selection")
}

func TestSetupSettingsClearDb(t *testing.T) {
	web := setupTestWeb(t)

	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 254}))
	assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification}))
	assert.Nil(t, web.arena.Database.CreateMatchResult(new(model.MatchResult)))
	assert.Nil(t, web.arena.Database.CreateRanking(&game.Ranking{TeamId: 254}))
	assert.Nil(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1}))
	recorder := web.postHttpResponse("/setup/db/clear", "")
	assert.Equal(t, 303, recorder.Code)

	teams, _ := web.arena.Database.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ := web.arena.Database.GetMatchesByType(model.Qualification, true)
	assert.Empty(t, matches)
	rankings, _ := web.arena.Database.GetAllRankings()
	assert.Empty(t, rankings)
	tournament.CalculateRankings(web.arena.Database, false)
	assert.Empty(t, rankings)
	alliances, _ := web.arena.Database.GetAllAlliances()
	assert.Empty(t, alliances)
	assert.Empty(t, web.arena.AllianceSelectionAlliances)
}

func TestSetupSettingsBackupRestoreDb(t *testing.T) {
	web := setupTestWeb(t)

	// Modify a parameter so that we know when the database has been restored.
	web.arena.EventSettings.Name = "Chezy Champs"
	assert.Nil(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))

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

func TestSetupSettingsPublishToTba(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.TbaClient.BaseUrl = "fakeurl"
	web.arena.EventSettings.TbaPublishingEnabled = true

	recorder := web.getHttpResponse("/setup/settings/publish_alliances")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish alliances")

	recorder = web.getHttpResponse("/setup/settings/publish_awards")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish awards")

	recorder = web.getHttpResponse("/setup/settings/publish_matches")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to delete published matches")

	recorder = web.getHttpResponse("/setup/settings/publish_rankings")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish rankings")

	recorder = web.getHttpResponse("/setup/settings/publish_teams")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish teams")
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
