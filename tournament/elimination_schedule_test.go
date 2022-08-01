// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package tournament

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEliminationScheduleInitial(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 2)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "F-1", 1, 2)
		assertMatch(t, matches[1], "F-2", 1, 2)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	CreateTestAlliances(database, 3)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "SF2-1", 2, 3)
		assertMatch(t, matches[1], "SF2-2", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	CreateTestAlliances(database, 4)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 5)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 6)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 7)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 8)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 9)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 10)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 11)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 12)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 13)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 14)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 15)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

	CreateTestAlliances(database, 16)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

func TestEliminationScheduleErrors(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 1)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at least 2 alliances", err.Error())
	}
	database.TruncateAlliances()

	CreateTestAlliances(database, 17)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Round of depth 32 is not supported", err.Error())
	}
	database.TruncateAlliances()
}

func TestEliminationSchedulePopulatePartialMatch(t *testing.T) {
	database := setupTestDb(t)

	// Final should be updated after semifinal is concluded.
	CreateTestAlliances(database, 3)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "SF2-1", model.BlueWonMatch)
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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
	CreateTestAlliances(database, 4)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "SF2-1", model.RedWonMatch)
	scoreMatch(database, "SF2-2", model.RedWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF1-2", model.RedWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
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

func TestEliminationScheduleCreateNextRound(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 4)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "SF1-1", model.BlueWonMatch)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-1", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-2", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "F-1", 4, 3)
		assertMatch(t, matches[5], "F-2", 4, 3)
	}
}

func TestEliminationScheduleDetermineWinner(t *testing.T) {
	database := setupTestDb(t)

	// Round with one tie and a sweep.
	CreateTestAlliances(database, 2)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "F-1", model.TieMatch)
	won, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	if assert.Nil(t, err) {
		assert.True(t, won)
	}
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with one tie and a split.
	CreateTestAlliances(database, 2)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "F-1", model.RedWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 2, len(matches))
	scoreMatch(database, "F-2", model.TieMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	assert.Equal(t, "F-4", matches[3].DisplayName)
	scoreMatch(database, "F-4", model.TieMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	scoreMatch(database, "F-5", model.RedWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	if assert.Nil(t, err) {
		assert.True(t, won)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with two ties.
	CreateTestAlliances(database, 2)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "F-1", model.TieMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", model.TieMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, "F-4", matches[3].DisplayName)
	}
	scoreMatch(database, "F-4", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	if assert.Nil(t, err) {
		assert.True(t, won)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with repeated ties.
	CreateTestAlliances(database, 2)
	updateAndAssertSchedule := func(expectedNumMatches int, expectedWon bool) {
		won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
		assert.Nil(t, err)
		assert.Equal(t, expectedWon, won)
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

func TestEliminationScheduleRemoveUnneededMatches(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 2)
	UpdateEliminationSchedule(database, time.Unix(0, 0))
	scoreMatch(database, "F-1", model.RedWonMatch)
	scoreMatch(database, "F-2", model.TieMatch)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ := database.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))

	// Check that the third match is deleted if the score is changed.
	scoreMatch(database, "F-2", model.RedWonMatch)
	won, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.True(t, won)

	// Check that the deleted match is recreated if the score is changed.
	scoreMatch(database, "F-2", model.BlueWonMatch)
	won, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	assert.False(t, won)
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 3, len(matches)) {
		assert.Equal(t, "F-3", matches[2].DisplayName)
	}
}

func TestEliminationScheduleChangePreviousRoundResult(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 4)
	_, err := UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(database, "SF2-1", model.RedWonMatch)
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(database, "SF2-3", model.RedWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(database, "SF2-3", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(matches))

	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF1-2", model.RedWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(database, "SF1-2", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(database, "SF1-3", model.BlueWonMatch)
	_, err = UpdateEliminationSchedule(database, time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[6], "F-1", 4, 3)
		assertMatch(t, matches[7], "F-2", 4, 3)
	}
}

func TestEliminationScheduleTiming(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 4)
	UpdateEliminationSchedule(database, time.Unix(1000, 0))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, int64(1000), matches[0].Time.Unix())
		assert.Equal(t, int64(1600), matches[1].Time.Unix())
		assert.Equal(t, int64(2200), matches[2].Time.Unix())
		assert.Equal(t, int64(2800), matches[3].Time.Unix())
	}
	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF1-2", model.BlueWonMatch)
	UpdateEliminationSchedule(database, time.Unix(5000, 0))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 5, len(matches)) {
		assert.Equal(t, int64(1000), matches[0].Time.Unix())
		assert.Equal(t, int64(5000), matches[1].Time.Unix())
		assert.Equal(t, int64(2200), matches[2].Time.Unix())
		assert.Equal(t, int64(5600), matches[3].Time.Unix())
		assert.Equal(t, int64(6200), matches[4].Time.Unix())
	}
}

func TestEliminationScheduleTeamPositions(t *testing.T) {
	database := setupTestDb(t)

	CreateTestAlliances(database, 4)
	UpdateEliminationSchedule(database, time.Unix(1000, 0))
	matches, _ := database.GetMatchesByType("elimination")
	match1 := matches[0]
	match2 := matches[1]
	assert.Equal(t, 102, match1.Red1)
	assert.Equal(t, 101, match1.Red2)
	assert.Equal(t, 103, match1.Red3)
	assert.Equal(t, 302, match2.Blue1)
	assert.Equal(t, 301, match2.Blue2)
	assert.Equal(t, 303, match2.Blue3)

	// Shuffle the team positions and check that the subsequent matches in the same round have the same ones.
	match1.Red1, match1.Red2 = match1.Red2, 104
	match2.Blue1, match2.Blue3 = 305, match2.Blue1
	database.UpdateMatch(&match1)
	database.UpdateMatch(&match2)
	scoreMatch(database, "SF1-1", model.RedWonMatch)
	scoreMatch(database, "SF2-1", model.BlueWonMatch)
	UpdateEliminationSchedule(database, time.Unix(1000, 0))
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, match1.Red1, matches[0].Red1)
		assert.Equal(t, match1.Red2, matches[0].Red2)
		assert.Equal(t, match1.Red3, matches[0].Red3)
		assert.Equal(t, match1.Red1, matches[2].Red1)
		assert.Equal(t, match1.Red2, matches[2].Red2)
		assert.Equal(t, match1.Red3, matches[2].Red3)

		assert.Equal(t, match2.Blue1, matches[1].Blue1)
		assert.Equal(t, match2.Blue2, matches[1].Blue2)
		assert.Equal(t, match2.Blue3, matches[1].Blue3)
		assert.Equal(t, match2.Blue1, matches[3].Blue1)
		assert.Equal(t, match2.Blue2, matches[3].Blue2)
		assert.Equal(t, match2.Blue3, matches[3].Blue3)
	}

	// Advance them to the finals and verify that the team position updates have been propagated.
	scoreMatch(database, "SF1-2", model.RedWonMatch)
	scoreMatch(database, "SF2-2", model.BlueWonMatch)
	UpdateEliminationSchedule(database, time.Unix(5000, 0))
	matches, _ = database.GetMatchesByType("elimination")
	if assert.Equal(t, 6, len(matches)) {
		for i := 4; i < 6; i++ {
			assert.Equal(t, match1.Red1, matches[i].Red1)
			assert.Equal(t, match1.Red2, matches[i].Red2)
			assert.Equal(t, match1.Red3, matches[i].Red3)
			assert.Equal(t, match2.Blue1, matches[i].Blue1)
			assert.Equal(t, match2.Blue2, matches[i].Blue2)
			assert.Equal(t, match2.Blue3, matches[i].Blue3)
		}
	}
}

func assertMatch(t *testing.T, match model.Match, displayName string, redAlliance int, blueAlliance int) {
	assert.Equal(t, displayName, match.DisplayName)
	assert.Equal(t, redAlliance, match.ElimRedAlliance)
	assert.Equal(t, blueAlliance, match.ElimBlueAlliance)
	assert.Equal(t, 100*redAlliance+2, match.Red1)
	assert.Equal(t, 100*redAlliance+1, match.Red2)
	assert.Equal(t, 100*redAlliance+3, match.Red3)
	assert.Equal(t, 100*blueAlliance+2, match.Blue1)
	assert.Equal(t, 100*blueAlliance+1, match.Blue2)
	assert.Equal(t, 100*blueAlliance+3, match.Blue3)
}

func scoreMatch(database *model.Database, displayName string, winner model.MatchStatus) {
	match, _ := database.GetMatchByName("elimination", displayName)
	match.Status = winner
	database.UpdateMatch(match)
	UpdateAlliance(database, [3]int{match.Red1, match.Red2, match.Red3}, match.ElimRedAlliance)
	UpdateAlliance(database, [3]int{match.Blue1, match.Blue2, match.Blue3}, match.ElimBlueAlliance)
}
