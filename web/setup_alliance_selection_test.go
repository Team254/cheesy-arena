// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupAllianceSelection(t *testing.T) {
	web := setupTestWeb(t)

	cachedAlliances = [][]*model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.NumElimAlliances = 15
	web.arena.EventSettings.SelectionRound3Order = "L"
	for i := 1; i <= 10; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Check that there are no alliance placeholders to start.
	recorder := web.getHttpResponse("/setup/alliance_selection")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Start the alliance selection.
	recorder = web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 15, len(cachedAlliances)) {
		assert.Equal(t, 4, len(cachedAlliances[0]))
	}
	recorder = web.getHttpResponse("/setup/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "Captain")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Reset the alliance selection.
	recorder = web.postHttpResponse("/setup/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")
	web.arena.EventSettings.NumElimAlliances = 3
	web.arena.EventSettings.SelectionRound3Order = ""
	recorder = web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 3, len(cachedAlliances)) {
		assert.Equal(t, 3, len(cachedAlliances[0]))
	}

	// Update one team at a time.
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=110")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 110, cachedAlliances[0][0].TeamId)
	recorder = web.getHttpResponse("/setup/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "\"110\"")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Update multiple teams at a time.
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=101&selection0_1=102&selection1_0=103")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 101, cachedAlliances[0][0].TeamId)
	assert.Equal(t, 102, cachedAlliances[0][1].TeamId)
	assert.Equal(t, 103, cachedAlliances[1][0].TeamId)
	recorder = web.getHttpResponse("/setup/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Update remainder of teams.
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=101&selection0_1=102&"+
		"selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106&selection2_0=107&selection2_1=108&"+
		"selection2_2=109")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Finalize alliance selection.
	web.arena.Database.CreateTeam(&model.Team{Id: 254, YellowCard: true})
	recorder = web.postHttpResponse("/setup/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, err := web.arena.Database.GetAllAlliances()
	assert.Nil(t, err)
	if assert.Equal(t, 3, len(alliances)) {
		assert.Equal(t, 101, alliances[0][0].TeamId)
		assert.Equal(t, 105, alliances[1][1].TeamId)
		assert.Equal(t, 109, alliances[2][2].TeamId)
	}
	matches, err := web.arena.Database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))
	team, _ := web.arena.Database.GetTeamById(254)
	assert.False(t, team.YellowCard)
}

func TestSetupAllianceSelectionErrors(t *testing.T) {
	web := setupTestWeb(t)

	cachedAlliances = [][]*model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.NumElimAlliances = 2
	for i := 1; i <= 6; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Start an alliance selection that is already underway.
	recorder := web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already in progress")

	// Select invalid teams.
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid team number")
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=100")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "not present at this event")
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=101&selection1_1=101")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already part of an alliance")

	// Finalize early and without required parameters.
	recorder = web.postHttpResponse("/setup/alliance_selection/finalize",
		"startTime=2014-01-01 01:00:00 PM&matchSpacingSec=360")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "until all spots have been filled")
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=101&selection0_1=102&"+
		"selection0_2=103&selection1_0=104&selection1_1=105&selection1_2=106")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/setup/alliance_selection/finalize", "startTime=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "valid start time")

	// Finalize for real and check that TBA publishing is triggered.
	web.arena.TbaClient.BaseUrl = "fakeurl"
	web.arena.EventSettings.TbaPublishingEnabled = true
	recorder = web.postHttpResponse("/setup/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish alliances")

	// Do other things after finalization.
	recorder = web.postHttpResponse("/setup/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	recorder = web.postHttpResponse("/setup/alliance_selection/reset", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	recorder = web.postHttpResponse("/setup/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	cachedAlliances = [][]*model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	recorder = web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
}

func TestSetupAllianceSelectionAutofocus(t *testing.T) {
	web := setupTestWeb(t)

	cachedAlliances = [][]*model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.NumElimAlliances = 2

	// Straight draft.
	web.arena.EventSettings.SelectionRound2Order = "F"
	web.arena.EventSettings.SelectionRound3Order = "F"
	recorder := web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j := web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	cachedAlliances[0][0].TeamId = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	cachedAlliances[0][1].TeamId = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	cachedAlliances[1][0].TeamId = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	cachedAlliances[1][1].TeamId = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	cachedAlliances[0][2].TeamId = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	cachedAlliances[1][2].TeamId = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	cachedAlliances[0][3].TeamId = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	cachedAlliances[1][3].TeamId = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)

	// Double-serpentine draft.
	web.arena.EventSettings.SelectionRound2Order = "L"
	web.arena.EventSettings.SelectionRound3Order = "L"
	recorder = web.postHttpResponse("/setup/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/setup/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	cachedAlliances[0][0].TeamId = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	cachedAlliances[0][1].TeamId = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	cachedAlliances[1][0].TeamId = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	cachedAlliances[1][1].TeamId = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	cachedAlliances[1][2].TeamId = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	cachedAlliances[0][2].TeamId = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	cachedAlliances[1][3].TeamId = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	cachedAlliances[0][3].TeamId = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)
}
