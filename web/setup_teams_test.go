// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetupTeams(t *testing.T) {
	web := setupTestWeb(t)

	// Check that there are no teams to start.
	recorder := web.getHttpResponse("/setup/teams")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "0 teams")

	// Mock the URL to download team info from.
	teamInfoBody := `{
		"website": "http://www.team254.com",
		"name": "NASA Ames Research Center",
		"city": "San Jose",
		"rookie_year": 1999,
		"state_prov": "CA",
		"team_number": 254,
		"location": "San Jose, CA, USA",
		"key": "frc254",
		"country": "USA",
		"nickname": "The Cheesy Poofs"
	}`
	teamRobotsBody := `[
		{
			"team_key": "frc33",
			"name": "Buzz 22",
			"key": "frc33_2017",
			"year": 2017
		}
	]`
	teamAwardsBody := `[{
		"event_key": "2014cmp",
		"award_type": 1,
		"name": "Championship Winners",
		"recipient_list": [
		{
			"team_number": 254,
			"awardee": null
		},
		{
			"team_number": 2848,
			"awardee": null
		},
		{
			"team_number": 469,
			"awardee": null
		},
		{
			"team_number": 74,
			"awardee": null
		}
		],
		"year": 2014
	}]`
	eventBody := `{ "name": "Championship" }`
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "robots") {
			fmt.Fprintln(w, teamRobotsBody)
		} else if strings.Contains(r.RequestURI, "awards") {
			fmt.Fprintln(w, teamAwardsBody)
		} else if strings.Contains(r.RequestURI, "team") {
			fmt.Fprintln(w, teamInfoBody)
		} else if strings.Contains(r.RequestURI, "event") {
			fmt.Fprintln(w, eventBody)
		} else {
			http.Error(w, "Unexpected request during test", 500)
		}
	}))
	defer tbaServer.Close()
	web.arena.TbaClient.BaseUrl = tbaServer.URL

	// Add some teams.
	recorder = web.postHttpResponse("/setup/teams", "teamNumbers=254\r\nnotateam\r\n1114\r\n")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "2 teams")
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")
	assert.Contains(t, recorder.Body.String(), "1114")

	// Add another team.
	recorder = web.postHttpResponse("/setup/teams", "teamNumbers=33")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "3 teams")
	assert.Contains(t, recorder.Body.String(), "33")

	// Edit a team.
	recorder = web.getHttpResponse("/setup/teams/254/edit")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")
	recorder = web.postHttpResponse("/setup/teams/254/edit", "nickname=Teh Chezy Pofs")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "Teh Chezy Pofs")

	// Re-download team info from TBA.
	recorder = web.getHttpResponse("/setup/teams/refresh")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")
	assert.NotContains(t, recorder.Body.String(), "Teh Chezy Pofs")

	// Delete a team.
	recorder = web.postHttpResponse("/setup/teams/1114/delete", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "2 teams")

	// Test clearing all teams.
	recorder = web.postHttpResponse("/setup/teams/clear", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.Contains(t, recorder.Body.String(), "0 teams")
}

func TestSetupTeamsDisallowModification(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateTeam(&model.Team{Id: 254, Nickname: "The Cheesy Poofs"})
	web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification})

	// Disallow adding teams.
	recorder := web.postHttpResponse("/setup/teams", "teamNumbers=33")
	assert.Contains(t, recorder.Body.String(), "can't modify")
	assert.Contains(t, recorder.Body.String(), "1 teams")
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")

	// Disallow deleting team.
	recorder = web.postHttpResponse("/setup/teams/254/delete", "")
	assert.Contains(t, recorder.Body.String(), "can't modify")
	assert.Contains(t, recorder.Body.String(), "1 teams")
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")

	// Disallow clearing all teams.
	recorder = web.postHttpResponse("/setup/teams/clear", "")
	assert.Contains(t, recorder.Body.String(), "can't modify")
	assert.Contains(t, recorder.Body.String(), "1 teams")
	assert.Contains(t, recorder.Body.String(), "The Cheesy Poofs")

	// Allow editing a team.
	recorder = web.postHttpResponse("/setup/teams/254/edit", "nickname=Teh Chezy Pofs")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/teams")
	assert.NotContains(t, recorder.Body.String(), "can't modify")
	assert.Contains(t, recorder.Body.String(), "1 teams")
	assert.Contains(t, recorder.Body.String(), "Teh Chezy Pofs")
}

func TestSetupTeamsBadReqest(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/setup/teams/254/edit")
	assert.Equal(t, 400, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such team")
	recorder = web.postHttpResponse("/setup/teams/254/edit", "")
	assert.Equal(t, 400, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such team")
	recorder = web.postHttpResponse("/setup/teams/254/delete", "")
	assert.Equal(t, 400, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such team")
}

func TestSetupTeamsWpaKeys(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.EventSettings.NetworkSecurityEnabled = true

	team1 := &model.Team{Id: 254, WpaKey: "aaaaaaaa"}
	team2 := &model.Team{Id: 1114}
	web.arena.Database.CreateTeam(team1)
	web.arena.Database.CreateTeam(team2)

	recorder := web.getHttpResponse("/setup/teams/generate_wpa_keys?all=false")
	assert.Equal(t, 303, recorder.Code)
	team1, _ = web.arena.Database.GetTeamById(254)
	team2, _ = web.arena.Database.GetTeamById(1114)
	assert.Equal(t, "aaaaaaaa", team1.WpaKey)
	assert.Equal(t, 8, len(team2.WpaKey))

	recorder = web.getHttpResponse("/setup/teams/generate_wpa_keys?all=true")
	assert.Equal(t, 303, recorder.Code)
	team1, _ = web.arena.Database.GetTeamById(254)
	team3, _ := web.arena.Database.GetTeamById(1114)
	assert.NotEqual(t, "aaaaaaaa", team1.WpaKey)
	assert.Equal(t, 8, len(team1.WpaKey))
	assert.NotEqual(t, team2.WpaKey, team3.WpaKey)
	assert.Equal(t, 8, len(team3.WpaKey))

	// Disallow invalid manual WPA keys.
	recorder = web.postHttpResponse("/setup/teams/254/edit", "wpa_key=1234567")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "WPA key must be between 8 and 63 characters")
}
