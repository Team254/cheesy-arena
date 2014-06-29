// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchPlay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	match1 := Match{Type: "practice", DisplayName: "1", Status: "complete", Winner: "R"}
	match2 := Match{Type: "practice", DisplayName: "2"}
	match3 := Match{Type: "qualification", DisplayName: "1", Status: "complete", Winner: "B"}
	match4 := Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete", Winner: "T"}
	match5 := Match{Type: "elimination", DisplayName: "SF1-2"}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	db.CreateMatch(&match3)
	db.CreateMatch(&match4)
	db.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "P1")
	assert.Contains(t, recorder.Body.String(), "P2")
	assert.Contains(t, recorder.Body.String(), "Q1")
	assert.Contains(t, recorder.Body.String(), "SF1-1")
	assert.Contains(t, recorder.Body.String(), "SF1-2")
}

func TestMatchPlayQueue(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	match := Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 101,
		Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	recorder := getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101")
	assert.NotContains(t, recorder.Body.String(), "102")
	assert.NotContains(t, recorder.Body.String(), "103")
	assert.NotContains(t, recorder.Body.String(), "104")
	assert.NotContains(t, recorder.Body.String(), "105")
	assert.NotContains(t, recorder.Body.String(), "106")

	// Queue the match and check for the team numbers again.
	recorder = getHttpResponse(fmt.Sprintf("/match_play/%d/queue", match.Id))
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "101")
	assert.Contains(t, recorder.Body.String(), "102")
	assert.Contains(t, recorder.Body.String(), "103")
	assert.Contains(t, recorder.Body.String(), "104")
	assert.Contains(t, recorder.Body.String(), "105")
	assert.Contains(t, recorder.Body.String(), "106")
}

func TestMatchPlayErrors(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Queue an invalid match.
	recorder := getHttpResponse("/match_play/1114/queue")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
}
