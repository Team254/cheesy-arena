// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/Team254/cheesy-arena/game"
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
	bracket, err := newSingleEliminationBracket(database, 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "Playoff F-1", "F-1", "", 1, 2)
		assertMatch(t, matches[1], "Playoff F-2", "F-2", "", 1, 2)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 3)
	bracket, err = newSingleEliminationBracket(database, 3)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "Playoff SF2-1", "SF2-1", "", 2, 3)
		assertMatch(t, matches[1], "Playoff SF2-2", "SF2-2", "", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 4)
	bracket, err = newSingleEliminationBracket(database, 4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "Playoff SF1-1", "SF1-1", "", 1, 4)
		assertMatch(t, matches[1], "Playoff SF2-1", "SF2-1", "", 2, 3)
		assertMatch(t, matches[2], "Playoff SF1-2", "SF1-2", "", 1, 4)
		assertMatch(t, matches[3], "Playoff SF2-2", "SF2-2", "", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 5)
	bracket, err = newSingleEliminationBracket(database, 5)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[1], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[2], "Playoff SF2-1", "SF2-1", "", 2, 3)
		assertMatch(t, matches[3], "Playoff SF2-2", "SF2-2", "", 2, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 6)
	bracket, err = newSingleEliminationBracket(database, 6)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[1], "Playoff QF4-1", "QF4-1", "", 3, 6)
		assertMatch(t, matches[2], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[3], "Playoff QF4-2", "QF4-2", "", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 7)
	bracket, err = newSingleEliminationBracket(database, 7)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[0], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[1], "Playoff QF3-1", "QF3-1", "", 2, 7)
		assertMatch(t, matches[2], "Playoff QF4-1", "QF4-1", "", 3, 6)
		assertMatch(t, matches[3], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[4], "Playoff QF3-2", "QF3-2", "", 2, 7)
		assertMatch(t, matches[5], "Playoff QF4-2", "QF4-2", "", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 8)
	bracket, err = newSingleEliminationBracket(database, 8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "Playoff QF1-1", "QF1-1", "", 1, 8)
		assertMatch(t, matches[1], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[2], "Playoff QF3-1", "QF3-1", "", 2, 7)
		assertMatch(t, matches[3], "Playoff QF4-1", "QF4-1", "", 3, 6)
		assertMatch(t, matches[4], "Playoff QF1-2", "QF1-2", "", 1, 8)
		assertMatch(t, matches[5], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[6], "Playoff QF3-2", "QF3-2", "", 2, 7)
		assertMatch(t, matches[7], "Playoff QF4-2", "QF4-2", "", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 9)
	bracket, err = newSingleEliminationBracket(database, 9)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[2], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[3], "Playoff QF3-1", "QF3-1", "", 2, 7)
		assertMatch(t, matches[4], "Playoff QF4-1", "QF4-1", "", 3, 6)
		assertMatch(t, matches[5], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[6], "Playoff QF3-2", "QF3-2", "", 2, 7)
		assertMatch(t, matches[7], "Playoff QF4-2", "QF4-2", "", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 10)
	bracket, err = newSingleEliminationBracket(database, 10)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[2], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[3], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[4], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[5], "Playoff QF4-1", "QF4-1", "", 3, 6)
		assertMatch(t, matches[6], "Playoff QF2-2", "QF2-2", "", 4, 5)
		assertMatch(t, matches[7], "Playoff QF4-2", "QF4-2", "", 3, 6)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 11)
	bracket, err = newSingleEliminationBracket(database, 11)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[2], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[3], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[4], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[5], "Playoff EF8-2", "EF8-2", "", 6, 11)
		assertMatch(t, matches[6], "Playoff QF2-1", "QF2-1", "", 4, 5)
		assertMatch(t, matches[7], "Playoff QF2-2", "QF2-2", "", 4, 5)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 12)
	bracket, err = newSingleEliminationBracket(database, 12)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF4-1", "EF4-1", "", 5, 12)
		assertMatch(t, matches[2], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[3], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[4], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[5], "Playoff EF4-2", "EF4-2", "", 5, 12)
		assertMatch(t, matches[6], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[7], "Playoff EF8-2", "EF8-2", "", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 13)
	bracket, err = newSingleEliminationBracket(database, 13)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 10, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF3-1", "EF3-1", "", 4, 13)
		assertMatch(t, matches[2], "Playoff EF4-1", "EF4-1", "", 5, 12)
		assertMatch(t, matches[3], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[4], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[5], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[6], "Playoff EF3-2", "EF3-2", "", 4, 13)
		assertMatch(t, matches[7], "Playoff EF4-2", "EF4-2", "", 5, 12)
		assertMatch(t, matches[8], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[9], "Playoff EF8-2", "EF8-2", "", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 14)
	bracket, err = newSingleEliminationBracket(database, 14)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF3-1", "EF3-1", "", 4, 13)
		assertMatch(t, matches[2], "Playoff EF4-1", "EF4-1", "", 5, 12)
		assertMatch(t, matches[3], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[4], "Playoff EF7-1", "EF7-1", "", 3, 14)
		assertMatch(t, matches[5], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[6], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[7], "Playoff EF3-2", "EF3-2", "", 4, 13)
		assertMatch(t, matches[8], "Playoff EF4-2", "EF4-2", "", 5, 12)
		assertMatch(t, matches[9], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[10], "Playoff EF7-2", "EF7-2", "", 3, 14)
		assertMatch(t, matches[11], "Playoff EF8-2", "EF8-2", "", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 15)
	bracket, err = newSingleEliminationBracket(database, 15)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 14, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[1], "Playoff EF3-1", "EF3-1", "", 4, 13)
		assertMatch(t, matches[2], "Playoff EF4-1", "EF4-1", "", 5, 12)
		assertMatch(t, matches[3], "Playoff EF5-1", "EF5-1", "", 2, 15)
		assertMatch(t, matches[4], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[5], "Playoff EF7-1", "EF7-1", "", 3, 14)
		assertMatch(t, matches[6], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[7], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[8], "Playoff EF3-2", "EF3-2", "", 4, 13)
		assertMatch(t, matches[9], "Playoff EF4-2", "EF4-2", "", 5, 12)
		assertMatch(t, matches[10], "Playoff EF5-2", "EF5-2", "", 2, 15)
		assertMatch(t, matches[11], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[12], "Playoff EF7-2", "EF7-2", "", 3, 14)
		assertMatch(t, matches[13], "Playoff EF8-2", "EF8-2", "", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()

	tournament.CreateTestAlliances(database, 16)
	bracket, err = newSingleEliminationBracket(database, 16)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 16, len(matches)) {
		assertMatch(t, matches[0], "Playoff EF1-1", "EF1-1", "", 1, 16)
		assertMatch(t, matches[1], "Playoff EF2-1", "EF2-1", "", 8, 9)
		assertMatch(t, matches[2], "Playoff EF3-1", "EF3-1", "", 4, 13)
		assertMatch(t, matches[3], "Playoff EF4-1", "EF4-1", "", 5, 12)
		assertMatch(t, matches[4], "Playoff EF5-1", "EF5-1", "", 2, 15)
		assertMatch(t, matches[5], "Playoff EF6-1", "EF6-1", "", 7, 10)
		assertMatch(t, matches[6], "Playoff EF7-1", "EF7-1", "", 3, 14)
		assertMatch(t, matches[7], "Playoff EF8-1", "EF8-1", "", 6, 11)
		assertMatch(t, matches[8], "Playoff EF1-2", "EF1-2", "", 1, 16)
		assertMatch(t, matches[9], "Playoff EF2-2", "EF2-2", "", 8, 9)
		assertMatch(t, matches[10], "Playoff EF3-2", "EF3-2", "", 4, 13)
		assertMatch(t, matches[11], "Playoff EF4-2", "EF4-2", "", 5, 12)
		assertMatch(t, matches[12], "Playoff EF5-2", "EF5-2", "", 2, 15)
		assertMatch(t, matches[13], "Playoff EF6-2", "EF6-2", "", 7, 10)
		assertMatch(t, matches[14], "Playoff EF7-2", "EF7-2", "", 3, 14)
		assertMatch(t, matches[15], "Playoff EF8-2", "EF8-2", "", 6, 11)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
}

func TestSingleEliminationErrors(t *testing.T) {
	_, err := newSingleEliminationBracket(nil, 1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at least 2 alliances", err.Error())
	}

	_, err = newSingleEliminationBracket(nil, 17)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at most 16 alliances", err.Error())
	}
}

func TestSingleEliminationPopulatePartialMatch(t *testing.T) {
	database := setupTestDb(t)

	// Final should be updated after semifinal is concluded.
	tournament.CreateTestAlliances(database, 3)
	bracket, err := newSingleEliminationBracket(database, 3)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF2-1", game.BlueWonMatch)
	scoreMatch(database, "SF2-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[2], "Playoff F-1", "F-1", "", 1, 3)
		assertMatch(t, matches[3], "Playoff F-2", "F-2", "", 1, 3)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Final should be generated and populated as both semifinals conclude.
	tournament.CreateTestAlliances(database, 4)
	bracket, err = newSingleEliminationBracket(database, 4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF2-1", game.RedWonMatch)
	scoreMatch(database, "SF2-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-1", game.RedWonMatch)
	scoreMatch(database, "SF1-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "Playoff F-1", "F-1", "", 1, 2)
		assertMatch(t, matches[5], "Playoff F-2", "F-2", "", 1, 2)
	}
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()
}

func TestSingleEliminationCreateNextRound(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := newSingleEliminationBracket(database, 4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF1-1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, _ := database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF1-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 4, len(matches))
	scoreMatch(database, "SF2-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, _ = database.GetMatchesByType(model.Playoff)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "Playoff F-1", "F-1", "", 4, 3)
		assertMatch(t, matches[5], "Playoff F-2", "F-2", "", 4, 3)
	}
}

func TestSingleEliminationDetermineWinner(t *testing.T) {
	database := setupTestDb(t)

	// Round with one tie and a sweep.
	tournament.CreateTestAlliances(database, 2)
	bracket, err := newSingleEliminationBracket(database, 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "F-1", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.WinningAlliance())
	assert.Equal(t, 0, bracket.FinalistAlliance())
	matches, _ := database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 2, bracket.WinningAlliance())
	assert.Equal(t, 1, bracket.FinalistAlliance())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with one tie and a split.
	tournament.CreateTestAlliances(database, 2)
	bracket, err = newSingleEliminationBracket(database, 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "F-1", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 2, len(matches))
	scoreMatch(database, "F-2", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 4, len(matches))
	assert.Equal(t, "F-4", matches[3].ShortName)
	scoreMatch(database, "F-4", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	scoreMatch(database, "F-5", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 1, bracket.WinningAlliance())
	assert.Equal(t, 2, bracket.FinalistAlliance())
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with two ties.
	tournament.CreateTestAlliances(database, 2)
	bracket, err = newSingleEliminationBracket(database, 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "F-1", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))
	scoreMatch(database, "F-3", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, "F-4", matches[3].ShortName)
	}
	scoreMatch(database, "F-4", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.True(t, bracket.IsComplete())
	database.TruncateAlliances()
	database.TruncateMatches()
	database.TruncateMatchResults()

	// Round with repeated ties.
	tournament.CreateTestAlliances(database, 2)
	updateAndAssertSchedule := func(expectedNumMatches int, expectedWon bool) {
		assert.Nil(t, bracket.Update(&dummyStartTime))
		assert.Equal(t, expectedWon, bracket.IsComplete())
		matches, _ = database.GetMatchesByType(model.Playoff)
		assert.Equal(t, expectedNumMatches, len(matches))
	}
	updateAndAssertSchedule(2, false)
	scoreMatch(database, "F-1", game.TieMatch)
	updateAndAssertSchedule(3, false)
	scoreMatch(database, "F-2", game.TieMatch)
	updateAndAssertSchedule(4, false)
	scoreMatch(database, "F-3", game.TieMatch)
	updateAndAssertSchedule(5, false)
	scoreMatch(database, "F-4", game.TieMatch)
	updateAndAssertSchedule(6, false)
	scoreMatch(database, "F-5", game.TieMatch)
	updateAndAssertSchedule(7, false)
	scoreMatch(database, "F-6", game.TieMatch)
	updateAndAssertSchedule(8, false)
	scoreMatch(database, "F-7", game.RedWonMatch)
	updateAndAssertSchedule(8, false)
	scoreMatch(database, "F-8", game.BlueWonMatch)
	updateAndAssertSchedule(9, false)
	scoreMatch(database, "F-9", game.RedWonMatch)
	updateAndAssertSchedule(9, true)
}

func TestSingleEliminationRemoveUnneededMatches(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 2)
	bracket, err := newSingleEliminationBracket(database, 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "F-1", game.RedWonMatch)
	scoreMatch(database, "F-2", game.TieMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, _ := database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 3, len(matches))

	// Check that the third match is deleted if the score is changed.
	scoreMatch(database, "F-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.True(t, bracket.IsComplete())

	// Check that the deleted match is recreated if the score is changed.
	scoreMatch(database, "F-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	assert.False(t, bracket.IsComplete())
	matches, _ = database.GetMatchesByType(model.Playoff)
	if assert.Equal(t, 3, len(matches)) {
		assert.Equal(t, "F-3", matches[2].ShortName)
	}
}

func TestSingleEliminationChangePreviousRoundResult(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := newSingleEliminationBracket(database, 4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF2-1", game.RedWonMatch)
	scoreMatch(database, "SF2-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF2-3", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF2-3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(matches))

	scoreMatch(database, "SF1-1", game.RedWonMatch)
	scoreMatch(database, "SF1-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF1-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	scoreMatch(database, "SF1-3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[6], "Playoff F-1", "F-1", "", 4, 3)
		assertMatch(t, matches[7], "Playoff F-2", "F-2", "", 4, 3)
	}

	scoreMatch(database, "SF2-3", game.MatchNotPlayed)
	assert.Nil(t, bracket.Update(&dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))
}
