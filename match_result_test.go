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

	matchResult.BlueScore.Cycles[0].Truss = !matchResult.BlueScore.Cycles[0].Truss
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

func buildTestMatchResult(matchId int, playNumber int) MatchResult {
	cycle1 := Cycle{3, true, true, true, false, false}
	cycle2 := Cycle{2, false, false, false, true, false}
	cycle3 := Cycle{1, true, false, false, false, true}
	fouls := Fouls{[]Foul{Foul{25, "G22", 25.2}, Foul{25, "G18", 150}}, []Foul{Foul{1868, "G20", 0}}}
	matchResult := MatchResult{MatchId: matchId, PlayNumber: playNumber}
	matchResult.RedScore = Score{1, 2, 3, 4, 5, 6, 7, []Cycle{cycle1, cycle2, cycle3}}
	matchResult.BlueScore = Score{7, 6, 5, 4, 3, 2, 1, []Cycle{cycle3, cycle1, cycle1, cycle1}}
	matchResult.RedFouls = fouls
	matchResult.BlueFouls = Fouls{}
	matchResult.Cards = Cards{[]int{1868}, []int{}}
	return matchResult
}
