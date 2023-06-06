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

func TestNewPlayoffTournamentErrors(t *testing.T) {
	_, err := NewPlayoffTournament(5, 8)
	if assert.NotNil(t, err) {
		assert.Equal(t, "invalid playoff type: 5", err.Error())
	}
}

func TestPlayoffTournamentGetters(t *testing.T) {
	playoffTournament, err := NewPlayoffTournament(model.SingleEliminationPlayoff, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(playoffTournament.MatchGroups()))
	assert.Contains(t, playoffTournament.MatchGroups(), "F")
	assert.Equal(t, playoffTournament.FinalMatchup(), playoffTournament.MatchGroups()["F"])
	assert.False(t, playoffTournament.IsComplete())
	assert.Equal(t, 0, playoffTournament.WinningAllianceId())
	assert.Equal(t, 0, playoffTournament.FinalistAllianceId())

	playoffTournament.FinalMatchup().update(
		map[int]playoffMatchResult{43: {game.BlueWonMatch}, 44: {game.BlueWonMatch}},
	)
	assert.True(t, playoffTournament.IsComplete())
	assert.Equal(t, 2, playoffTournament.WinningAllianceId())
	assert.Equal(t, 1, playoffTournament.FinalistAllianceId())
}

func TestPlayoffTournamentCreateMatches(t *testing.T) {
	database := setupTestDb(t)
	tournament.CreateTestAlliances(database, 8)

	// Test double-elimination.
	playoffTournament, err := NewPlayoffTournament(model.DoubleEliminationPlayoff, 8)
	assert.Nil(t, err)

	startTime := time.Unix(5000, 0)
	err = playoffTournament.CreateMatches(database, startTime)
	assert.Nil(t, err)
	err = playoffTournament.CreateMatches(database, startTime)
	if assert.NotNil(t, err) {
		assert.Equal(t, "cannot create playoff matches; 19 matches already exist", err.Error())
	}

	matches, _ := database.GetMatchesByType(model.Playoff)
	if assert.Equal(t, 19, len(matches)) {
		assertMatch(t, matches[0], 1, 5000, "Match 1", "M1", "Round 1 Upper", "M1", 1, 8, true, "sf", 1, 1)
		assertMatch(t, matches[1], 2, 5600, "Match 2", "M2", "Round 1 Upper", "M2", 4, 5, true, "sf", 2, 1)
		assertMatch(t, matches[2], 3, 6200, "Match 3", "M3", "Round 1 Upper", "M3", 2, 7, true, "sf", 3, 1)
		assertMatch(t, matches[3], 4, 6800, "Match 4", "M4", "Round 1 Upper", "M4", 3, 6, true, "sf", 4, 1)
		assertMatch(t, matches[4], 5, 7400, "Match 5", "M5", "Round 2 Lower", "M5", 0, 0, true, "sf", 5, 1)
		assertMatch(t, matches[5], 6, 8000, "Match 6", "M6", "Round 2 Lower", "M6", 0, 0, true, "sf", 6, 1)
		assertMatch(t, matches[6], 7, 8600, "Match 7", "M7", "Round 2 Upper", "M7", 0, 0, true, "sf", 7, 1)
		assertMatch(t, matches[7], 8, 9200, "Match 8", "M8", "Round 2 Upper", "M8", 0, 0, true, "sf", 8, 1)
		assertMatch(t, matches[8], 9, 9800, "Match 9", "M9", "Round 3 Lower", "M9", 0, 0, true, "sf", 9, 1)
		assertMatch(t, matches[9], 10, 10400, "Match 10", "M10", "Round 3 Lower", "M10", 0, 0, true, "sf", 10, 1)
		assertMatch(t, matches[10], 11, 11000, "Match 11", "M11", "Round 4 Upper", "M11", 0, 0, true, "sf", 11, 1)
		assertMatch(t, matches[11], 12, 11600, "Match 12", "M12", "Round 4 Lower", "M12", 0, 0, true, "sf", 12, 1)
		assertMatch(t, matches[12], 13, 12200, "Match 13", "M13", "Round 5 Lower", "M13", 0, 0, true, "sf", 13, 1)
		assertMatch(t, matches[13], 14, 12800, "Final 1", "F1", "", "F", 0, 0, false, "f", 1, 1)
		assertMatch(t, matches[14], 15, 13400, "Final 2", "F2", "", "F", 0, 0, false, "f", 1, 2)
		assertMatch(t, matches[15], 16, 14000, "Final 3", "F3", "", "F", 0, 0, false, "f", 1, 3)
		assertMatch(t, matches[16], 17, 14600, "Overtime 1", "O1", "", "F", 0, 0, true, "f", 1, 4)
		assertMatch(t, matches[17], 18, 15200, "Overtime 2", "O2", "", "F", 0, 0, true, "f", 1, 5)
		assertMatch(t, matches[18], 19, 15800, "Overtime 3", "O3", "", "F", 0, 0, true, "f", 1, 6)
	}
	for i := 0; i < 16; i++ {
		assert.Equal(t, game.MatchScheduled, matches[i].Status)
	}
	for i := 17; i < 19; i++ {
		assert.Equal(t, game.MatchHidden, matches[i].Status)
	}

	// Test single-elimination.
	assert.Nil(t, database.TruncateMatches())
	playoffTournament, err = NewPlayoffTournament(model.SingleEliminationPlayoff, 3)
	assert.Nil(t, err)

	startTime = time.Unix(1000, 0)
	err = playoffTournament.CreateMatches(database, startTime)
	assert.Nil(t, err)

	matches, _ = database.GetMatchesByType(model.Playoff)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[0], 38, 1000, "Semifinal 2-1", "SF2-1", "", "SF2", 2, 3, true, "sf", 2, 1)
		assertMatch(t, matches[1], 40, 1600, "Semifinal 2-2", "SF2-2", "", "SF2", 2, 3, true, "sf", 2, 2)
		assertMatch(t, matches[2], 42, 2200, "Semifinal 2-3", "SF2-3", "", "SF2", 2, 3, true, "sf", 2, 3)
		assertMatch(t, matches[3], 43, 2800, "Final 1", "F1", "", "F", 1, 0, false, "f", 1, 1)
		assertMatch(t, matches[4], 44, 3400, "Final 2", "F2", "", "F", 1, 0, false, "f", 1, 2)
		assertMatch(t, matches[5], 45, 4000, "Final 3", "F3", "", "F", 1, 0, false, "f", 1, 3)
		assertMatch(t, matches[6], 46, 4600, "Overtime 1", "O1", "", "F", 1, 0, true, "f", 1, 4)
		assertMatch(t, matches[7], 47, 5200, "Overtime 2", "O2", "", "F", 1, 0, true, "f", 1, 5)
		assertMatch(t, matches[8], 48, 5800, "Overtime 3", "O3", "", "F", 1, 0, true, "f", 1, 6)
	}
	for i := 0; i < 6; i++ {
		assert.Equal(t, game.MatchScheduled, matches[i].Status)
	}
	for i := 6; i < 9; i++ {
		assert.Equal(t, game.MatchHidden, matches[i].Status)
	}
}

func TestPlayoffTournamentUpdateMatches(t *testing.T) {
	database := setupTestDb(t)
	tournament.CreateTestAlliances(database, 4)

	playoffTournament, err := NewPlayoffTournament(model.SingleEliminationPlayoff, 4)
	assert.Nil(t, err)

	err = playoffTournament.UpdateMatches(database)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "no matches exist")
	}

	err = playoffTournament.CreateMatches(database, time.Unix(0, 0))
	assert.Nil(t, err)

	matches, _ := database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 102, matches[0].Red1)
	assert.Equal(t, 101, matches[0].Red2)
	assert.Equal(t, 103, matches[0].Red3)
	assert.Equal(t, 402, matches[0].Blue1)
	assert.Equal(t, 401, matches[0].Blue2)
	assert.Equal(t, 403, matches[0].Blue3)

	matches[0].Status = game.BlueWonMatch
	err = database.UpdateMatch(&matches[0])
	assert.Nil(t, err)
	err = database.UpdateAllianceFromMatch(1, [3]int{103, 102, 101})
	assert.Nil(t, err)
	err = database.UpdateAllianceFromMatch(4, [3]int{404, 405, 406})
	assert.Nil(t, err)

	err = playoffTournament.UpdateMatches(database)
	assert.Nil(t, err)

	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 102, matches[0].Red1)
	assert.Equal(t, 101, matches[0].Red2)
	assert.Equal(t, 103, matches[0].Red3)
	assert.Equal(t, 402, matches[0].Blue1)
	assert.Equal(t, 401, matches[0].Blue2)
	assert.Equal(t, 403, matches[0].Blue3)
	assert.Equal(t, 103, matches[2].Red1)
	assert.Equal(t, 102, matches[2].Red2)
	assert.Equal(t, 101, matches[2].Red3)
	assert.Equal(t, 404, matches[2].Blue1)
	assert.Equal(t, 405, matches[2].Blue2)
	assert.Equal(t, 406, matches[2].Blue3)
	assert.Equal(t, 103, matches[4].Red1)
	assert.Equal(t, 102, matches[4].Red2)
	assert.Equal(t, 101, matches[4].Red3)
	assert.Equal(t, 404, matches[4].Blue1)
	assert.Equal(t, 405, matches[4].Blue2)
	assert.Equal(t, 406, matches[4].Blue3)

	matches[1].Status = game.BlueWonMatch
	err = database.UpdateMatch(&matches[1])
	assert.Nil(t, err)
	matches[2].Status = game.BlueWonMatch
	err = database.UpdateMatch(&matches[2])
	assert.Nil(t, err)
	matches[3].Status = game.BlueWonMatch
	err = database.UpdateMatch(&matches[3])
	assert.Nil(t, err)
	err = database.UpdateAllianceFromMatch(4, [3]int{403, 402, 406})
	assert.Nil(t, err)

	err = playoffTournament.UpdateMatches(database)
	assert.Nil(t, err)

	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 103, matches[2].Red1)
	assert.Equal(t, 102, matches[2].Red2)
	assert.Equal(t, 101, matches[2].Red3)
	assert.Equal(t, 404, matches[2].Blue1)
	assert.Equal(t, 405, matches[2].Blue2)
	assert.Equal(t, 406, matches[2].Blue3)
	assert.Equal(t, 4, matches[6].PlayoffRedAlliance)
	assert.Equal(t, 3, matches[6].PlayoffBlueAlliance)
	assert.Equal(t, 403, matches[6].Red1)
	assert.Equal(t, 402, matches[6].Red2)
	assert.Equal(t, 406, matches[6].Red3)
	assert.Equal(t, 302, matches[6].Blue1)
	assert.Equal(t, 301, matches[6].Blue2)
	assert.Equal(t, 303, matches[6].Blue3)

	// Change the outcome of some matches and verify that the teams in the finals are wiped out.
	matches[1].Status = game.RedWonMatch
	err = database.UpdateMatch(&matches[1])
	assert.Nil(t, err)
	matches[2].Status = game.RedWonMatch
	err = database.UpdateMatch(&matches[2])
	assert.Nil(t, err)

	err = playoffTournament.UpdateMatches(database)
	assert.Nil(t, err)

	matches, _ = database.GetMatchesByType(model.Playoff)
	assert.Equal(t, 0, matches[6].PlayoffRedAlliance)
	assert.Equal(t, 0, matches[6].PlayoffBlueAlliance)
	assert.Equal(t, 0, matches[6].Red1)
	assert.Equal(t, 0, matches[6].Red2)
	assert.Equal(t, 0, matches[6].Red3)
	assert.Equal(t, 0, matches[6].Blue1)
	assert.Equal(t, 0, matches[6].Blue2)
	assert.Equal(t, 0, matches[6].Blue3)
}
