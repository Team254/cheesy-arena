// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchReview(t *testing.T) {
	web := setupTestWeb(t)

	match1 := model.Match{Type: "practice", DisplayName: "1", Status: "complete", Winner: "R"}
	match2 := model.Match{Type: "practice", DisplayName: "2"}
	match3 := model.Match{Type: "qualification", DisplayName: "1", Status: "complete", Winner: "B"}
	match4 := model.Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete", Winner: "T"}
	match5 := model.Match{Type: "elimination", DisplayName: "SF1-2"}
	web.arena.Database.CreateMatch(&match1)
	web.arena.Database.CreateMatch(&match2)
	web.arena.Database.CreateMatch(&match3)
	web.arena.Database.CreateMatch(&match4)
	web.arena.Database.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">P1<")
	assert.Contains(t, recorder.Body.String(), ">P2<")
	assert.Contains(t, recorder.Body.String(), ">Q1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-2<")
}

func TestMatchReviewEditExistingResult(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 1001,
		Red2: 1002, Red3: 1003, Blue1: 1004, Blue2: 1005, Blue3: 1006, ElimRedAlliance: 1, ElimBlueAlliance: 2}
	web.arena.Database.CreateMatch(&match)
	matchResult := model.BuildTestMatchResult(match.Id, 1)
	matchResult.MatchType = match.Type
	web.arena.Database.CreateMatchResult(matchResult)
	tournament.CreateTestAlliances(web.arena.Database, 2)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Contains(t, recorder.Body.String(), ">71<") // The red score
	assert.Contains(t, recorder.Body.String(), ">72<") // The blue score

	// Check response for non-existent match.
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " QF4-3 ")

	// Update the score to something else.
	postBody := "redScoreJson={\"RobotEndLevels\":[0,3,0]}&blueScoreJson={\"CargoBays\":[0,2,1,2,2,0,1]," +
		"\"Fouls\":[{\"TeamId\":973,\"Rule\":\"G22\"}]}&redCardsJson={\"105\":\"yellow\"}&blueCardsJson={}"
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code)

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Contains(t, recorder.Body.String(), ">15<") // The red score
	assert.Contains(t, recorder.Body.String(), ">19<") // The blue score
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 1001,
		Red2: 1002, Red3: 1003, Blue1: 1004, Blue2: 1005, Blue3: 1006, ElimRedAlliance: 1, ElimBlueAlliance: 2}
	web.arena.Database.CreateMatch(&match)
	tournament.CreateTestAlliances(web.arena.Database, 2)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.NotContains(t, recorder.Body.String(), ">71<") // The red score
	assert.NotContains(t, recorder.Body.String(), ">72<") // The blue score

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " QF4-3 ")

	// Update the score to something else.
	postBody := "redScoreJson={\"RocketNearLeftBays\":[1,0,2]}&blueScoreJson={\"RocketFarRightBays\":[2,2,2]," +
		"\"Fouls\":[{\"TeamId\":973,\"Rule\":\"G22\"}]}&redCardsJson={\"105\":\"yellow\"}&blueCardsJson={}"
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code)

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Contains(t, recorder.Body.String(), ">10<") // The red score
	assert.Contains(t, recorder.Body.String(), ">15<") // The blue score
}
