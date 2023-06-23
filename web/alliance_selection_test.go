// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllianceSelection(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 15
	web.arena.EventSettings.SelectionRound3Order = "L"
	for i := 1; i <= 10; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Check that there are no alliance placeholders to start.
	recorder := web.getHttpResponse("/alliance_selection")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Start the alliance selection.
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 15, len(web.arena.AllianceSelectionAlliances)) {
		assert.Equal(t, 4, len(web.arena.AllianceSelectionAlliances[0].TeamIds))
	}
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "Captain")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Reset the alliance selection.
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Captain")
	assert.NotContains(t, recorder.Body.String(), ">110<")
	web.arena.EventSettings.NumPlayoffAlliances = 3
	web.arena.EventSettings.SelectionRound3Order = ""
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	if assert.Equal(t, 3, len(web.arena.AllianceSelectionAlliances)) {
		assert.Equal(t, 3, len(web.arena.AllianceSelectionAlliances[0].TeamIds))
	}

	// Update one team at a time.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=110")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 110, web.arena.AllianceSelectionAlliances[0].TeamIds[0])
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), "\"110\"")
	assert.NotContains(t, recorder.Body.String(), ">110<")

	// Update multiple teams at a time.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection1_0=103")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, 101, web.arena.AllianceSelectionAlliances[0].TeamIds[0])
	assert.Equal(t, 102, web.arena.AllianceSelectionAlliances[0].TeamIds[1])
	assert.Equal(t, 103, web.arena.AllianceSelectionAlliances[1].TeamIds[0])
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Update remainder of teams.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection0_2=103&"+
		"selection1_0=104&selection1_1=105&selection1_2=106&selection2_0=107&selection2_1=108&selection2_2=109")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/alliance_selection")
	assert.Contains(t, recorder.Body.String(), ">110<")

	// Finalize alliance selection.
	web.arena.Database.CreateTeam(&model.Team{Id: 254, YellowCard: true})
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, err := web.arena.Database.GetAllAlliances()
	assert.Nil(t, err)
	if assert.Equal(t, 3, len(alliances)) {
		assert.Equal(t, 101, alliances[0].TeamIds[0])
		assert.Equal(t, 105, alliances[1].TeamIds[1])
		assert.Equal(t, 109, alliances[2].TeamIds[2])

		// Check that the initial lineup is populated correctly.
		assert.Equal(t, 102, alliances[0].Lineup[0])
		assert.Equal(t, 101, alliances[0].Lineup[1])
		assert.Equal(t, 103, alliances[0].Lineup[2])
	}
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, false)
	assert.Nil(t, err)
	assert.Equal(t, 16, len(matches))
	team, _ := web.arena.Database.GetTeamById(254)
	assert.False(t, team.YellowCard)
}

func TestAllianceSelectionErrors(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	for i := 1; i <= 6; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Start an alliance selection that is already underway.
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already in progress")

	// Select invalid teams.
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid team number")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=100")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "ineligible for selection")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection1_1=101")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already part of an alliance")

	// Finalize early and without required parameters.
	recorder = web.postHttpResponse("/alliance_selection/finalize",
		"startTime=2014-01-01 01:00:00 PM&matchSpacingSec=360")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "until all spots have been filled")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection0_2=103&"+
		"selection1_0=104&selection1_1=105&selection1_2=106")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "valid start time")

	// Finalize for real and check that TBA publishing is triggered.
	web.arena.TbaClient.BaseUrl = "fakeurl"
	web.arena.EventSettings.TbaPublishingEnabled = true
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish alliances")

	// Do other things after finalization.
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=asdf")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	cachedRankedTeams = []*RankedTeam{}
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already been finalized")
}

func TestAllianceSelectionReset(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	for i := 1; i <= 6; i++ {
		web.arena.Database.CreateRanking(&game.Ranking{TeamId: 100 + i, Rank: i})
	}

	// Start, populate, and finalize the alliance selection.
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection0_2=103&"+
		"selection1_0=104&selection1_1=105&selection1_2=106")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ := web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ := web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)

	// Reset the alliance selection before any matches have been played.
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.Empty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.Empty(t, matches)

	// Start, populate, and finalize the alliance selection again.
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection", "selection0_0=101&selection0_1=102&selection0_2=103&"+
		"selection1_0=104&selection1_1=105&selection1_2=106")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/finalize", "startTime=2014-01-01 01:00:00 PM")
	assert.Equal(t, 303, recorder.Code)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)

	// Mark a match as played and verify that the alliance selection can no longer be reset.
	matches[0].Status = game.RedWonMatch
	assert.Nil(t, web.arena.Database.UpdateMatch(&matches[0]))
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "matches have already started")
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)
}

func TestAllianceSelectionAutofocus(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2

	// Straight draft.
	web.arena.EventSettings.SelectionRound2Order = "F"
	web.arena.EventSettings.SelectionRound3Order = "F"
	recorder := web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j := web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[0] = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[1] = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[0] = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[1] = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[2] = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[2] = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[3] = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[3] = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)

	// Double-serpentine draft.
	web.arena.EventSettings.SelectionRound2Order = "L"
	web.arena.EventSettings.SelectionRound3Order = "L"
	recorder = web.postHttpResponse("/alliance_selection/reset", "")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/alliance_selection/start", "")
	assert.Equal(t, 303, recorder.Code)
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[0] = 1
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[1] = 2
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 0, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[0] = 3
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 1, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[1] = 4
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[2] = 5
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 2, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[2] = 6
	i, j = web.determineNextCell()
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[1].TeamIds[3] = 7
	i, j = web.determineNextCell()
	assert.Equal(t, 0, i)
	assert.Equal(t, 3, j)
	web.arena.AllianceSelectionAlliances[0].TeamIds[3] = 8
	i, j = web.determineNextCell()
	assert.Equal(t, -1, i)
	assert.Equal(t, -1, j)
}
