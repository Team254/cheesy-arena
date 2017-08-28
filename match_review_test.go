// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchReview(t *testing.T) {
	setupTest(t)

	match1 := model.Match{Type: "practice", DisplayName: "1", Status: "complete", Winner: "R"}
	match2 := model.Match{Type: "practice", DisplayName: "2"}
	match3 := model.Match{Type: "qualification", DisplayName: "1", Status: "complete", Winner: "B"}
	match4 := model.Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete", Winner: "T"}
	match5 := model.Match{Type: "elimination", DisplayName: "SF1-2"}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	db.CreateMatch(&match3)
	db.CreateMatch(&match4)
	db.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "P1")
	assert.Contains(t, recorder.Body.String(), "P2")
	assert.Contains(t, recorder.Body.String(), "Q1")
	assert.Contains(t, recorder.Body.String(), "SF1-1")
	assert.Contains(t, recorder.Body.String(), "SF1-2")
}

func TestMatchReviewEditExistingResult(t *testing.T) {
	setupTest(t)

	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 1001,
		Red2: 1002, Red3: 1003, Blue1: 1004, Blue2: 1005, Blue3: 1006}
	db.CreateMatch(&match)
	matchResult := model.BuildTestMatchResult(match.Id, 1)
	matchResult.MatchType = match.Type
	db.CreateMatchResult(matchResult)
	tournament.CreateTestAlliances(db, 2)

	recorder := getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")
	assert.Contains(t, recorder.Body.String(), "210") // The red score
	assert.Contains(t, recorder.Body.String(), "533") // The blue score

	// Check response for non-existent match.
	recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")

	// Update the score to something else.
	postBody := "redScoreJson={\"AutoMobility\":3}&blueScoreJson={\"Rotors\":3," +
		"\"Fouls\":[{\"TeamId\":973,\"Rule\":\"G22\"}]}&redCardsJson={\"105\":\"yellow\"}&blueCardsJson={}"
	recorder = postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 302, recorder.Code)

	// Check for the updated scores back on the match list page.
	recorder = getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")
	assert.Contains(t, recorder.Body.String(), "20")  // The red score
	assert.Contains(t, recorder.Body.String(), "120") // The blue score
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	setupTest(t)

	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 1001,
		Red2: 1002, Red3: 1003, Blue1: 1004, Blue2: 1005, Blue3: 1006}
	db.CreateMatch(&match)
	tournament.CreateTestAlliances(db, 2)

	recorder := getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")
	assert.NotContains(t, recorder.Body.String(), "210") // The red score
	assert.NotContains(t, recorder.Body.String(), "533") // The blue score

	recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")

	// Update the score to something else.
	postBody := "redScoreJson={\"AutoRotors\":1}&blueScoreJson={\"FuelHigh\":30," +
		"\"Fouls\":[{\"TeamId\":973,\"Rule\":\"G22\"}]}&redCardsJson={\"105\":\"yellow\"}&blueCardsJson={}"
	recorder = postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 302, recorder.Code)

	// Check for the updated scores back on the match list page.
	recorder = getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")
	assert.Contains(t, recorder.Body.String(), "65") // The red score
	assert.Contains(t, recorder.Body.String(), "10") // The blue score
}
