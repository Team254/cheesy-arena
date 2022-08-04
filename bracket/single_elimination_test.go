// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package bracket

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var dummyStartTime = time.Unix(0, 0)

func TestSingleEliminationInitial(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 2)
	bracket, err := NewSingleEliminationBracket(2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "F-1", 1, 2)
		assertMatch(t, matches[1], "F-2", 1, 2)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 3)
	bracket, err = NewSingleEliminationBracket(3)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "SF2-1", 2, 3)
		assertMatch(t, matches[1], "SF2-2", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 4)
	bracket, err = NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "SF1-1", 1, 4)
		assertMatch(t, matches[1], "SF2-1", 2, 3)
		assertMatch(t, matches[2], "SF1-2", 1, 4)
		assertMatch(t, matches[3], "SF2-2", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 5)
	bracket, err = NewSingleEliminationBracket(5)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF2-2", 4, 5)
		assertMatch(t, matches[2], "SF2-1", 2, 3)
		assertMatch(t, matches[3], "SF2-2", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 6)
	bracket, err = NewSingleEliminationBracket(6)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF4-1", 3, 6)
		assertMatch(t, matches[2], "QF2-2", 4, 5)
		assertMatch(t, matches[3], "QF4-2", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 7)
	bracket, err = NewSingleEliminationBracket(7)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF3-1", 2, 7)
		assertMatch(t, matches[2], "QF4-1", 3, 6)
		assertMatch(t, matches[3], "QF2-2", 4, 5)
		assertMatch(t, matches[4], "QF3-2", 2, 7)
		assertMatch(t, matches[5], "QF4-2", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 8)
	bracket, err = NewSingleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "QF1-1", 1, 8)
		assertMatch(t, matches[1], "QF2-1", 4, 5)
		assertMatch(t, matches[2], "QF3-1", 2, 7)
		assertMatch(t, matches[3], "QF4-1", 3, 6)
		assertMatch(t, matches[4], "QF1-2", 1, 8)
		assertMatch(t, matches[5], "QF2-2", 4, 5)
		assertMatch(t, matches[6], "QF3-2", 2, 7)
		assertMatch(t, matches[7], "QF4-2", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 9)
	bracket, err = NewSingleEliminationBracket(9)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF2-2", 8, 9)
		assertMatch(t, matches[2], "QF2-1", 4, 5)
		assertMatch(t, matches[3], "QF3-1", 2, 7)
		assertMatch(t, matches[4], "QF4-1", 3, 6)
		assertMatch(t, matches[5], "QF2-2", 4, 5)
		assertMatch(t, matches[6], "QF3-2", 2, 7)
		assertMatch(t, matches[7], "QF4-2", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 10)
	bracket, err = NewSingleEliminationBracket(10)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF6-1", 7, 10)
		assertMatch(t, matches[2], "EF2-2", 8, 9)
		assertMatch(t, matches[3], "EF6-2", 7, 10)
		assertMatch(t, matches[4], "QF2-1", 4, 5)
		assertMatch(t, matches[5], "QF4-1", 3, 6)
		assertMatch(t, matches[6], "QF2-2", 4, 5)
		assertMatch(t, matches[7], "QF4-2", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 11)
	bracket, err = NewSingleEliminationBracket(11)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF6-1", 7, 10)
		assertMatch(t, matches[2], "EF8-1", 6, 11)
		assertMatch(t, matches[3], "EF2-2", 8, 9)
		assertMatch(t, matches[4], "EF6-2", 7, 10)
		assertMatch(t, matches[5], "EF8-2", 6, 11)
		assertMatch(t, matches[6], "QF2-1", 4, 5)
		assertMatch(t, matches[7], "QF2-2", 4, 5)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 12)
	bracket, err = NewSingleEliminationBracket(12)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF4-1", 5, 12)
		assertMatch(t, matches[2], "EF6-1", 7, 10)
		assertMatch(t, matches[3], "EF8-1", 6, 11)
		assertMatch(t, matches[4], "EF2-2", 8, 9)
		assertMatch(t, matches[5], "EF4-2", 5, 12)
		assertMatch(t, matches[6], "EF6-2", 7, 10)
		assertMatch(t, matches[7], "EF8-2", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 13)
	bracket, err = NewSingleEliminationBracket(13)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 10, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF6-1", 7, 10)
		assertMatch(t, matches[4], "EF8-1", 6, 11)
		assertMatch(t, matches[5], "EF2-2", 8, 9)
		assertMatch(t, matches[6], "EF3-2", 4, 13)
		assertMatch(t, matches[7], "EF4-2", 5, 12)
		assertMatch(t, matches[8], "EF6-2", 7, 10)
		assertMatch(t, matches[9], "EF8-2", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 14)
	bracket, err = NewSingleEliminationBracket(14)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF6-1", 7, 10)
		assertMatch(t, matches[4], "EF7-1", 3, 14)
		assertMatch(t, matches[5], "EF8-1", 6, 11)
		assertMatch(t, matches[6], "EF2-2", 8, 9)
		assertMatch(t, matches[7], "EF3-2", 4, 13)
		assertMatch(t, matches[8], "EF4-2", 5, 12)
		assertMatch(t, matches[9], "EF6-2", 7, 10)
		assertMatch(t, matches[10], "EF7-2", 3, 14)
		assertMatch(t, matches[11], "EF8-2", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 15)
	bracket, err = NewSingleEliminationBracket(15)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 14, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF5-1", 2, 15)
		assertMatch(t, matches[4], "EF6-1", 7, 10)
		assertMatch(t, matches[5], "EF7-1", 3, 14)
		assertMatch(t, matches[6], "EF8-1", 6, 11)
		assertMatch(t, matches[7], "EF2-2", 8, 9)
		assertMatch(t, matches[8], "EF3-2", 4, 13)
		assertMatch(t, matches[9], "EF4-2", 5, 12)
		assertMatch(t, matches[10], "EF5-2", 2, 15)
		assertMatch(t, matches[11], "EF6-2", 7, 10)
		assertMatch(t, matches[12], "EF7-2", 3, 14)
		assertMatch(t, matches[13], "EF8-2", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 16)
	bracket, err = NewSingleEliminationBracket(16)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 16, len(matches)) {
		assertMatch(t, matches[0], "EF1-1", 1, 16)
		assertMatch(t, matches[1], "EF2-1", 8, 9)
		assertMatch(t, matches[2], "EF3-1", 4, 13)
		assertMatch(t, matches[3], "EF4-1", 5, 12)
		assertMatch(t, matches[4], "EF5-1", 2, 15)
		assertMatch(t, matches[5], "EF6-1", 7, 10)
		assertMatch(t, matches[6], "EF7-1", 3, 14)
		assertMatch(t, matches[7], "EF8-1", 6, 11)
		assertMatch(t, matches[8], "EF1-2", 1, 16)
		assertMatch(t, matches[9], "EF2-2", 8, 9)
		assertMatch(t, matches[10], "EF3-2", 4, 13)
		assertMatch(t, matches[11], "EF4-2", 5, 12)
		assertMatch(t, matches[12], "EF5-2", 2, 15)
		assertMatch(t, matches[13], "EF6-2", 7, 10)
		assertMatch(t, matches[14], "EF7-2", 3, 14)
		assertMatch(t, matches[15], "EF8-2", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
}

func TestSingleEliminationErrors(t *testing.T) {
	_, err := NewSingleEliminationBracket(1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at least 2 alliances", err.Error())
	}

	_, err = NewSingleEliminationBracket(17)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at most 16 alliances", err.Error())
	}
}

func TestSingleEliminationPopulatePartialMatch(t *testing.T) {
	database := setupTestDb(t)

	// Final should be updated after semifinal is concluded.
	tournament.CreateTestAlliances(database, 3)
	bracket, err := NewSingleEliminationBracket(3)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF2-1", model.BlueWonMatch)
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[2], "F-1", 1, 3)
		assertMatch(t, matches[3], "F-2", 1, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Final should be generated and populated as both semifinals conclude.
	tournament.CreateTestAlliances(database, 4)
	bracket, err = NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF2-1", model.RedWonMatch)
	scoreMatch(database, "SF2-2", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF1-2", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "F-1", 1, 2)
		assertMatch(t, matches[5], "F-2", 1, 2)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()
}

func TestSingleEliminationCreateNextRound(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF1-1", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-1", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "F-1", 4, 3)
		assertMatch(t, matches[5], "F-2", 4, 3)
	}
}

func TestSingleEliminationDetermineWinner(t *testing.T) {
	database := setupTestDb(t)

	// Round with one tie and a sweep.
	tournament.CreateTestAlliances(database, 2)
	bracket, err := NewSingleEliminationBracket(2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "F-1", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 2, bracket.Winner())
	assert.Equal(t, 1, bracket.Finalist())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with one tie and a split.
	tournament.CreateTestAlliances(database, 2)
	bracket, err = NewSingleEliminationBracket(2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "F-1", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 2, len(matches))
	scoreMatch(database, "F-2", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	assert.Equal(t, "F-4", matches[3].DisplayName)
	scoreMatch(database, "F-4", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	scoreMatch(database, "F-5", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 1, bracket.Winner())
	assert.Equal(t, 2, bracket.Finalist())
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with two ties.
	tournament.CreateTestAlliances(database, 2)
	bracket, err = NewSingleEliminationBracket(2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "F-1", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, "F-4", matches[3].DisplayName)
	}
	scoreMatch(database, "F-4", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.True(t, bracket.IsComplete())
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with repeated ties.
	tournament.CreateTestAlliances(database, 2)
	updateAndAssertSchedule := func(expectedNumMatches int, expectedWon bool) {
		assert.Nil(t, bracket.Update(database, &dummyStartTime))
		assert.Equal(t, expectedWon, bracket.IsComplete())
		matches, _ = database.GetMatchesByType("elimination")
		assert.Equal(t, expectedNumMatches, len(matches))
	}
	updateAndAssertSchedule(2, false)
	scoreMatch(database, "F-1", model.TieMatch)
	updateAndAssertSchedule(3, false)
	scoreMatch(database, "F-2", model.TieMatch)
	updateAndAssertSchedule(4, false)
	scoreMatch(database, "F-3", model.TieMatch)
	updateAndAssertSchedule(5, false)
	scoreMatch(database, "F-4", model.TieMatch)
	updateAndAssertSchedule(6, false)
	scoreMatch(database, "F-5", model.TieMatch)
	updateAndAssertSchedule(7, false)
	scoreMatch(database, "F-6", model.TieMatch)
	updateAndAssertSchedule(8, false)
	scoreMatch(database, "F-7", model.RedWonMatch)
	updateAndAssertSchedule(8, false)
	scoreMatch(database, "F-8", model.BlueWonMatch)
	updateAndAssertSchedule(9, false)
	scoreMatch(database, "F-9", model.RedWonMatch)
	updateAndAssertSchedule(9, true)
}

func TestSingleEliminationRemoveUnneededMatches(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 2)
	bracket, err := NewSingleEliminationBracket(2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "F-1", model.RedWonMatch)
	scoreMatch(database, "F-2", model.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))

	// Check that the third match is deleted if the score is changed.
	scoreMatch(database, "F-2", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.True(t, bracket.IsComplete())

	// Check that the deleted match is recreated if the score is changed.
	scoreMatch(database, "F-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 3, len(matches)) {
		assert.Equal(t, "F-3", matches[2].DisplayName)
	}
}

func TestSingleEliminationChangePreviousRoundResult(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF2-1", model.RedWonMatch)
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF2-3", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF2-3", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(matches))

	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF1-2", model.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF1-2", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	scoreMatch(database, "SF1-3", model.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[6], "F-1", 4, 3)
		assertMatch(t, matches[7], "F-2", 4, 3)
	}
}
