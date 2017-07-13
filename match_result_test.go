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

	matchResult.BlueScore.AutoMobility = 12
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
	assert.Equal(t, 0, redSummary.AutoMobilityPoints)
	assert.Equal(t, 80, redSummary.AutoPoints)
	assert.Equal(t, 100, redSummary.RotorPoints)
	assert.Equal(t, 50, redSummary.TakeoffPoints)
	assert.Equal(t, 40, redSummary.PressurePoints)
	assert.Equal(t, 0, redSummary.BonusPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 190, redSummary.Score)
	assert.Equal(t, true, redSummary.PressureGoalReached)
	assert.Equal(t, false, redSummary.RotorGoalReached)

	blueSummary := matchResult.BlueScoreSummary()
	assert.Equal(t, 10, blueSummary.AutoMobilityPoints)
	assert.Equal(t, 133, blueSummary.AutoPoints)
	assert.Equal(t, 200, blueSummary.RotorPoints)
	assert.Equal(t, 150, blueSummary.TakeoffPoints)
	assert.Equal(t, 18, blueSummary.PressurePoints)
	assert.Equal(t, 0, blueSummary.BonusPoints)
	assert.Equal(t, 55, blueSummary.FoulPoints)
	assert.Equal(t, 433, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PressureGoalReached)
	assert.Equal(t, true, blueSummary.RotorGoalReached)

	// Test pressure boundary conditions.
	matchResult.RedScore.AutoFuelHigh = 19
	assert.Equal(t, false, matchResult.RedScoreSummary().PressureGoalReached)
	matchResult.RedScore.FuelLow = 18
	assert.Equal(t, true, matchResult.RedScoreSummary().PressureGoalReached)
	matchResult.RedScore.AutoFuelLow = 1
	assert.Equal(t, false, matchResult.RedScoreSummary().PressureGoalReached)
	matchResult.RedScore.FuelHigh = 56
	assert.Equal(t, true, matchResult.RedScoreSummary().PressureGoalReached)

	// Test rotor boundary conditions.
	matchResult.BlueScore.AutoGears = 2
	assert.Equal(t, false, matchResult.BlueScoreSummary().RotorGoalReached)
	matchResult.BlueScore.Gears = 11
	assert.Equal(t, true, matchResult.BlueScoreSummary().RotorGoalReached)

	// Test elimination bonus.
	matchResult.MatchType = "elimination"
	assert.Equal(t, 20, matchResult.RedScoreSummary().BonusPoints)
	assert.Equal(t, 210, matchResult.RedScoreSummary().Score)
	assert.Equal(t, 100, matchResult.BlueScoreSummary().BonusPoints)
	assert.Equal(t, 513, matchResult.BlueScoreSummary().Score)
	matchResult.RedScore.Gears = 12
	assert.Equal(t, 120, matchResult.RedScoreSummary().BonusPoints)
	matchResult.MatchType = "qualification"
	assert.Equal(t, 0, matchResult.RedScoreSummary().BonusPoints)
	assert.Equal(t, 0, matchResult.BlueScoreSummary().BonusPoints)
}

func buildTestMatchResult(matchId int, playNumber int) MatchResult {
	fouls := []Foul{Foul{25, "G22", false, 25.2}, Foul{25, "G18", true, 150}, Foul{1868, "G20", true, 0}}
	matchResult := MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = Score{0, 1, 2, 20, 4, 12, 55, 1, fouls, false}
	matchResult.BlueScore = Score{2, 3, 10, 0, 10, 65, 24, 3, []Foul{}, false}
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}
