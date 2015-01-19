// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchReview(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	match1 := Match{Type: "practice", DisplayName: "1", Status: "complete"}
	match2 := Match{Type: "practice", DisplayName: "2"}
	match3 := Match{Type: "qualification", DisplayName: "1", Status: "complete"}
	match4 := Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete"}
	match5 := Match{Type: "elimination", DisplayName: "SF1-2"}
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
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	match := Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Red1: 101, Red2: 102,
		Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	matchResult := buildTestMatchResult(match.Id, 1)
	db.CreateMatchResult(&matchResult)
	createTestAlliances(db, 2)

	recorder := getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")
	assert.Contains(t, recorder.Body.String(), "152") // The red score
	assert.Contains(t, recorder.Body.String(), "102") // The blue score

	// Check response for non-existent match.
	recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "QF4-3")

	// Update the score to something else.
	// TODO(pat): Update for 2015.
	/*
		postBody := "redScoreJson={\"AutoMobilityBonuses\":3}&blueScoreJson={\"Cycles\":[{\"ScoredHigh\":true}]}&" +
			"redFoulsJson=[{\"TeamId\":103,\"IsTechnical\":false}]&blueFoulsJson=[{\"TeamId\":104,\"IsTechnical\":" +
			"true}]&redCardsJson={\"105\":\"yellow\"}&blueCardsJson={}"
		recorder = postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
		assert.Equal(t, 302, recorder.Code)

		// Check for the updated scores back on the match list page.
		recorder = getHttpResponse("/match_review")
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "QF4-3")
		assert.Contains(t, recorder.Body.String(), "65") // The red score
		assert.Contains(t, recorder.Body.String(), "30") // The blue score
	*/
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// TODO(pat): Update for 2015.
	/*
		match := Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Red1: 101, Red2: 102,
			Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
		db.CreateMatch(&match)
		createTestAlliances(db, 2)

		recorder := getHttpResponse("/match_review")
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "QF4-3")
		assert.NotContains(t, recorder.Body.String(), "312") // The red score
		assert.NotContains(t, recorder.Body.String(), "593") // The blue score

		recorder = getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "QF4-3")

		// Update the score to something else.
		postBody := "redScoreJson={\"AutoHighHot\":4}&blueScoreJson={\"Cycles\":[{\"Assists\":3," +
			"\"ScoredLow\":true}]}&redFoulsJson=[]&blueFoulsJson=[]&redCardsJson={}&blueCardsJson={}"
		recorder = postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
		assert.Equal(t, 302, recorder.Code)

		// Check for the updated scores back on the match list page.
		recorder = getHttpResponse("/match_review")
		assert.Equal(t, 200, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "QF4-3")
		assert.Contains(t, recorder.Body.String(), "80") // The red score
		assert.Contains(t, recorder.Body.String(), "31") // The blue score
	*/
}
