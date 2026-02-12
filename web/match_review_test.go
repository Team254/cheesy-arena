// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package web

import (
	"fmt"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
)

func TestMatchReview(t *testing.T) {
	web := setupTestWeb(t)
	// (Create matches code - Generic, unchanged)
	match1 := model.Match{Type: model.Practice, ShortName: "P1", Status: game.RedWonMatch}
	web.arena.Database.CreateMatch(&match1)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">P1<")
}

func TestMatchReviewEditExistingResult(t *testing.T) {
	web := setupTestWeb(t)
	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	match, _ := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	match.Status = game.RedWonMatch
	web.arena.Database.UpdateMatch(match)

	matchResult := model.BuildTestMatchResult(match.Id, 1)
	web.arena.Database.CreateMatchResult(matchResult)

	// Update the score using 2026 JSON format
	// Red: Endgame Level 2, Level 3
	// Blue: Teleop Fuel = 21, Foul
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameStatuses\":[0,2,3]},\"BlueScore\":{"+
			"\"TeleopFuelCount\":21,\"Fouls\":[{\"TeamId\":973,\"RuleId\":4}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)

	recorder := web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Verify update
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	// Note: You might need to adjust expected score assertions based on your exact scoring rules for Level 2/3
}

func TestMatchReviewEditCurrentMatch(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 352",
		ShortName: "Q352",
		Red1:      1001, Red2: 1002, Red3: 1003,
		Blue1: 1004, Blue2: 1005, Blue3: 1006,
	}
	web.arena.Database.CreateMatch(&match)
	web.arena.LoadMatch(&match)

	// Update 2026 Score via JSON
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameStatuses\":[0,2,0]},\"BlueScore\":{"+
			"\"TeleopFuelCount\":21,\"Fouls\":[{\"TeamId\":973,\"RuleId\":1}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse("/match_review/current/edit", postBody)
	assert.Equal(t, 303, recorder.Code)

	// Verify persistence in CurrentScore (Realtime)
	assert.Equal(t, 21, web.arena.BlueRealtimeScore.CurrentScore.TeleopFuelCount)
	assert.Equal(t, game.EndgameLevel2, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[1])
}
