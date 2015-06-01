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

func TestScoreSummary(t *testing.T) {
	matchResult := buildTestMatchResult(1, 1)
	redSummary := matchResult.RedScoreSummary()
	assert.Equal(t, 40, redSummary.CoopertitionPoints)
	assert.Equal(t, 28, redSummary.AutoPoints)
	assert.Equal(t, 24, redSummary.ContainerPoints)
	assert.Equal(t, 12, redSummary.TotePoints)
	assert.Equal(t, 6, redSummary.LitterPoints)
	assert.Equal(t, 18, redSummary.FoulPoints)
	assert.Equal(t, 92, redSummary.Score)

	blueSummary := matchResult.BlueScoreSummary()
	assert.Equal(t, 40, blueSummary.CoopertitionPoints)
	assert.Equal(t, 10, blueSummary.AutoPoints)
	assert.Equal(t, 24, blueSummary.ContainerPoints)
	assert.Equal(t, 24, blueSummary.TotePoints)
	assert.Equal(t, 6, blueSummary.LitterPoints)
	assert.Equal(t, 0, blueSummary.FoulPoints)
	assert.Equal(t, 104, blueSummary.Score)
}

func TestCorrectEliminationScore(t *testing.T) {
	// TODO(patrick): Test proper calculation of DQ.
	matchResult := MatchResult{}
	matchResult.CorrectEliminationScore()

	// TODO(patrick): Put back elim tiebreaker tests if the game calls for it.
}

func buildTestMatchResult(matchId int, playNumber int) MatchResult {
	fouls := []Foul{Foul{25, "G22", 25.2}, Foul{25, "G18", 150}, Foul{1868, "G20", 0}}
	stacks1 := []Stack{Stack{6, true, true}, Stack{0, false, false}, Stack{0, true, false}}
	stacks2 := []Stack{Stack{5, true, false}, Stack{6, false, false}, Stack{1, true, true}}
	matchResult := MatchResult{MatchId: matchId, PlayNumber: playNumber}
	matchResult.RedScore = Score{false, true, false, true, stacks1, false, true, fouls, false}
	matchResult.BlueScore = Score{true, false, true, false, stacks2, false, true, []Foul{}, false}
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}
