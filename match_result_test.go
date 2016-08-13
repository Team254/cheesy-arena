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

	matchResult.BlueScore.AutoDefensesReached = 12
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
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	matchResult := buildTestMatchResult(1, 1)
	redSummary := matchResult.RedScoreSummary()
	assert.Equal(t, 55, redSummary.AutoPoints)
	assert.Equal(t, 60, redSummary.DefensePoints)
	assert.Equal(t, 86, redSummary.GoalPoints)
	assert.Equal(t, 10, redSummary.ScaleChallengePoints)
	assert.Equal(t, 101, redSummary.TeleopPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 0, redSummary.BonusPoints)
	assert.Equal(t, 156, redSummary.Score)
	assert.Equal(t, true, redSummary.Breached)
	assert.Equal(t, false, redSummary.Captured)
	assert.Equal(t, -5, redSummary.TowerStrength)

	blueSummary := matchResult.BlueScoreSummary()
	assert.Equal(t, 22, blueSummary.AutoPoints)
	assert.Equal(t, 25, blueSummary.DefensePoints)
	assert.Equal(t, 36, blueSummary.GoalPoints)
	assert.Equal(t, 35, blueSummary.ScaleChallengePoints)
	assert.Equal(t, 76, blueSummary.TeleopPoints)
	assert.Equal(t, 15, blueSummary.FoulPoints)
	assert.Equal(t, 0, blueSummary.BonusPoints)
	assert.Equal(t, 113, blueSummary.Score)
	assert.Equal(t, false, blueSummary.Breached)
	assert.Equal(t, false, blueSummary.Captured)
	assert.Equal(t, 1, blueSummary.TowerStrength)

	// Test breach boundary conditions.
	matchResult.RedScore.DefensesCrossed[4] = 2
	assert.Equal(t, true, matchResult.RedScoreSummary().Breached)
	matchResult.RedScore.AutoDefensesCrossed[0] = 0
	assert.Equal(t, true, matchResult.RedScoreSummary().Breached)
	matchResult.RedScore.DefensesCrossed[1] = 1
	assert.Equal(t, false, matchResult.RedScoreSummary().Breached)

	// Test capture boundary conditions.
	matchResult.BlueScore.AutoHighGoals = 1
	assert.Equal(t, 0, matchResult.BlueScoreSummary().TowerStrength)
	assert.Equal(t, true, matchResult.BlueScoreSummary().Captured)
	matchResult.BlueScore.HighGoals = 5
	assert.Equal(t, true, matchResult.BlueScoreSummary().Captured)
	matchResult.BlueScore.Challenges = 0
	assert.Equal(t, false, matchResult.BlueScoreSummary().Captured)

	// Test elimination bonus.
	matchResult.MatchType = "elimination"
	matchResult.RedScore.DefensesCrossed[1] = 2
	assert.Equal(t, 20, matchResult.RedScoreSummary().BonusPoints)
	assert.Equal(t, 171, matchResult.RedScoreSummary().Score)
	matchResult.BlueScore.Challenges = 1
	assert.Equal(t, 25, matchResult.BlueScoreSummary().BonusPoints)
	assert.Equal(t, 153, matchResult.BlueScoreSummary().Score)
	matchResult.RedScore.Scales = 1
	assert.Equal(t, 45, matchResult.RedScoreSummary().BonusPoints)
	matchResult.MatchType = "qualification"
	assert.Equal(t, 0, matchResult.RedScoreSummary().BonusPoints)
	assert.Equal(t, 0, matchResult.BlueScoreSummary().BonusPoints)
}

func buildTestMatchResult(matchId int, playNumber int) MatchResult {
	fouls := []Foul{Foul{25, "G22", false, 25.2}, Foul{25, "G18", true, 150}, Foul{1868, "G20", true, 0}}
	matchResult := MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = Score{0, [5]int{1, 0, 1, 1, 0}, 1, 2, [5]int{1, 2, 1, 1, 1}, 3, 11, 2, 0, fouls, false}
	matchResult.BlueScore = Score{1, [5]int{0, 1, 0, 0, 0}, 2, 0, [5]int{1, 1, 0, 0, 1}, 3, 4, 1, 2, []Foul{}, false}
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}
