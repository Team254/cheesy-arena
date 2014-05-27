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

func TestCycleScores(t *testing.T) {
	matchResult := MatchResult{}
	matchResult.RedScore.Cycles = []Cycle{Cycle{}}
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)
	cycle := &matchResult.RedScore.Cycles[0]

	*cycle = Cycle{Assists: 3, Truss: false, Catch: false, ScoredHigh: false, ScoredLow: false, DeadBall: true}
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: true, Catch: false, ScoredHigh: false, ScoredLow: false, DeadBall: true}
	assert.Equal(t, 10, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: false, Catch: true, ScoredHigh: false, ScoredLow: false, DeadBall: true}
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: true, Catch: true, ScoredHigh: false, ScoredLow: false, DeadBall: true}
	assert.Equal(t, 20, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 0, Truss: false, Catch: false, ScoredHigh: true, ScoredLow: false, DeadBall: false}
	assert.Equal(t, 10, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 1, Truss: false, Catch: false, ScoredHigh: true, ScoredLow: false, DeadBall: false}
	assert.Equal(t, 10, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 2, Truss: false, Catch: false, ScoredHigh: true, ScoredLow: false, DeadBall: false}
	assert.Equal(t, 20, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: false, Catch: true, ScoredHigh: true, ScoredLow: false, DeadBall: false}
	assert.Equal(t, 40, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 1, Truss: false, Catch: false, ScoredHigh: false, ScoredLow: true, DeadBall: false}
	assert.Equal(t, 1, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 2, Truss: false, Catch: true, ScoredHigh: false, ScoredLow: true, DeadBall: false}
	assert.Equal(t, 11, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: false, Catch: false, ScoredHigh: false, ScoredLow: true, DeadBall: false}
	assert.Equal(t, 31, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: true, Catch: true, ScoredHigh: false, ScoredLow: true, DeadBall: false}
	assert.Equal(t, 51, matchResult.RedScoreSummary().Score)

	*cycle = Cycle{Assists: 3, Truss: true, Catch: true, ScoredHigh: true, ScoredLow: true, DeadBall: false}
	assert.Equal(t, 60, matchResult.RedScoreSummary().Score)
}

func TestScoreSummary(t *testing.T) {
	matchResult := buildTestMatchResult(1, 1)
	redSummary := matchResult.RedScoreSummary()
	assert.Equal(t, 164, redSummary.AutoPoints)
	assert.Equal(t, 40, redSummary.AssistPoints)
	assert.Equal(t, 30, redSummary.TrussCatchPoints)
	assert.Equal(t, 78, redSummary.GoalPoints)
	assert.Equal(t, 148, redSummary.TeleopPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 312, redSummary.Score)

	blueSummary := matchResult.BlueScoreSummary()
	assert.Equal(t, 292, blueSummary.AutoPoints)
	assert.Equal(t, 90, blueSummary.AssistPoints)
	assert.Equal(t, 70, blueSummary.TrussCatchPoints)
	assert.Equal(t, 51, blueSummary.GoalPoints)
	assert.Equal(t, 211, blueSummary.TeleopPoints)
	assert.Equal(t, 90, blueSummary.FoulPoints)
	assert.Equal(t, 593, blueSummary.Score)
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
