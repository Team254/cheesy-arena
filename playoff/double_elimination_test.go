// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoubleEliminationInitial(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "Playoff 1", "1", "Round 1 Upper", 1, 8)
		assertMatch(t, matches[1], "Playoff 2", "2", "Round 1 Upper", 4, 5)
		assertMatch(t, matches[2], "Playoff 3", "3", "Round 1 Upper", 2, 7)
		assertMatch(t, matches[3], "Playoff 4", "4", "Round 1 Upper", 3, 6)
	}
}

func TestDoubleEliminationErrors(t *testing.T) {
	_, err := NewDoubleEliminationBracket(7)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have exactly 8 alliances", err.Error())
	}

	_, err = NewDoubleEliminationBracket(9)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have exactly 8 alliances", err.Error())
	}
}

func TestDoubleEliminationProgression(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "Playoff 5", "5", "Round 2 Lower", 1, 5)
		assertMatch(t, matches[5], "Playoff 7", "7", "Round 2 Upper", 8, 4)
	}

	scoreMatch(database, "3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))

	scoreMatch(database, "4", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[4], "Playoff 5", "5", "Round 2 Lower", 1, 5)
		assertMatch(t, matches[5], "Playoff 6", "6", "Round 2 Lower", 2, 6)
		assertMatch(t, matches[6], "Playoff 7", "7", "Round 2 Upper", 8, 4)
		assertMatch(t, matches[7], "Playoff 8", "8", "Round 2 Upper", 7, 3)
	}

	scoreMatch(database, "5", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(matches))

	scoreMatch(database, "6", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(matches))

	scoreMatch(database, "7", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[8], "Playoff 9", "9", "Round 3 Lower", 8, 2)
	}

	scoreMatch(database, "8", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 11, len(matches)) {
		assertMatch(t, matches[9], "Playoff 10", "10", "Round 3 Lower", 7, 5)
		assertMatch(t, matches[10], "Playoff 11", "11", "Round 4 Upper", 4, 3)
	}

	scoreMatch(database, "9", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 11, len(matches))

	scoreMatch(database, "10", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[10], "Playoff 11", "11", "Round 4 Upper", 4, 3)
		assertMatch(t, matches[11], "Playoff 12", "12", "Round 4 Lower", 7, 8)
	}

	scoreMatch(database, "11", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 12, len(matches))

	scoreMatch(database, "12", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 13, len(matches)) {
		assertMatch(t, matches[12], "Playoff 13", "13", "Round 5 Lower", 3, 7)
	}

	scoreMatch(database, "13", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 15, len(matches)) {
		assertMatch(t, matches[13], "Playoff F-1", "F-1", "", 4, 7)
		assertMatch(t, matches[14], "Playoff F-2", "F-2", "", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 15, len(matches))
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 16, len(matches)) {
		assertMatch(t, matches[15], "Playoff F-3", "F-3", "", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-3", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 17, len(matches)) {
		assertMatch(t, matches[16], "Playoff F-4", "F-4", "", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-4", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 18, len(matches)) {
		assertMatch(t, matches[17], "Playoff F-5", "F-5", "", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-5", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 18, len(matches))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 7, bracket.Winner())
	assert.Equal(t, 4, bracket.Finalist())
}

func TestDoubleEliminationTie(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 5, len(matches)) {
		assertMatch(t, matches[4], "Playoff 1-2", "1-2", "Round 1 Upper", 1, 8)
	}

	scoreMatch(database, "1-2", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[5], "Playoff 1-3", "1-3", "Round 1 Upper", 1, 8)
	}

	scoreMatch(database, "1-3", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))
}

func TestDoubleEliminationChangeResult(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "Playoff 5", "5", "Round 2 Lower", 1, 5)
		assertMatch(t, matches[5], "Playoff 7", "7", "Round 2 Upper", 8, 4)
	}

	scoreMatch(database, "2", game.MatchNotPlayed)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "Playoff 5", "5", "Round 2 Lower", 1, 4)
		assertMatch(t, matches[5], "Playoff 7", "7", "Round 2 Upper", 8, 5)
	}
}
