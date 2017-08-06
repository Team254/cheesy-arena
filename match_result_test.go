// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentMatchResult(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match, err := db.GetMatchResultForMatch(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchResultCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	matchResult := buildTestMatchResult(254, 5)
	db.CreateMatchResult(matchResult)
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	matchResult.BlueScore.AutoMobility = 12
	db.SaveMatchResult(matchResult)
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	db.DeleteMatchResult(matchResult)
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestTruncateMatchResults(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	matchResult := buildTestMatchResult(254, 1)
	db.CreateMatchResult(matchResult)
	db.TruncateMatchResults()
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestGetMatchResultForMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	matchResult := buildTestMatchResult(254, 2)
	db.CreateMatchResult(matchResult)
	matchResult2 := buildTestMatchResult(254, 5)
	db.CreateMatchResult(matchResult2)
	matchResult3 := buildTestMatchResult(254, 4)
	db.CreateMatchResult(matchResult3)

	// Should return the match result with the highest play number (i.e. the most recent).
	matchResult4, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult2, matchResult4)
}

func buildTestMatchResult(matchId int, playNumber int) *MatchResult {
	matchResult := &MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = game.TestScore1()
	matchResult.BlueScore = game.TestScore2()
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}
