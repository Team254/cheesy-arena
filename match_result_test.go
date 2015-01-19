// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
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
	db.CreateMatchResult(&matchResult)
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, *matchResult2)

	matchResult.BlueScore.CoopertitionSet = !matchResult.BlueScore.CoopertitionSet
	db.SaveMatchResult(&matchResult)
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, *matchResult2)

	db.DeleteMatchResult(&matchResult)
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
	db.CreateMatchResult(&matchResult)
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
	db.CreateMatchResult(&matchResult)
	matchResult2 := buildTestMatchResult(254, 5)
	db.CreateMatchResult(&matchResult2)
	matchResult3 := buildTestMatchResult(254, 4)
	db.CreateMatchResult(&matchResult3)

	// Should return the match result with the highest play number (i.e. the most recent).
	matchResult4, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult2, *matchResult4)
}

func TestScoring(t *testing.T) {
	matchResult := MatchResult{}
	score := &matchResult.RedScore
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)

	// TODO(pat): Test all scoring combinations.
	*score = Score{}
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)
}

func TestScoreSummary(t *testing.T) {
	matchResult := buildTestMatchResult(1, 1)
	redSummary := matchResult.RedScoreSummary()
	assert.Equal(t, 40, redSummary.CoopertitionPoints)
	assert.Equal(t, 28, redSummary.AutoPoints)
	assert.Equal(t, 36, redSummary.ContainerPoints)
	assert.Equal(t, 2, redSummary.TotePoints)
	assert.Equal(t, 64, redSummary.LitterPoints)
	assert.Equal(t, 18, redSummary.FoulPoints)
	assert.Equal(t, 152, redSummary.Score)

	blueSummary := matchResult.BlueScoreSummary()
	assert.Equal(t, 20, blueSummary.CoopertitionPoints)
	assert.Equal(t, 10, blueSummary.AutoPoints)
	assert.Equal(t, 36, blueSummary.ContainerPoints)
	assert.Equal(t, 12, blueSummary.TotePoints)
	assert.Equal(t, 24, blueSummary.LitterPoints)
	assert.Equal(t, 0, blueSummary.FoulPoints)
	assert.Equal(t, 102, blueSummary.Score)
}

func TestCorrectEliminationScore(t *testing.T) {
	// TODO(patrick): Test proper calculation of DQ.
	matchResult := MatchResult{}
	matchResult.CorrectEliminationScore()
}

func buildTestMatchResult(matchId int, playNumber int) MatchResult {
	fouls := []Foul{Foul{25, "G22", 25.2}, Foul{25, "G18", 150}, Foul{1868, "G20", 0}}
	matchResult := MatchResult{MatchId: matchId, PlayNumber: playNumber}
	matchResult.RedScore = Score{false, false, true, true, 1, []int{2, 3, 4}, 5, 6, 7, false, true, fouls, false}
	// 6 + 20 + 2 + 36 + 30 + 6 + 28 + 40
	matchResult.BlueScore = Score{true, true, false, false, 6, []int{5, 4}, 3, 2, 1, true, false, []Foul{}, false}
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}
