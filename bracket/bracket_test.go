// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package bracket

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewBracketErrors(t *testing.T) {
	_, err := newBracket([]matchupTemplate{}, newMatchupKey(33, 12), 8)
	if assert.NotNil(t, err) {
		assert.Equal(t, "could not find template for matchup {Round:33 Group:12} in the list of templates", err.Error())
	}

	matchTemplate := matchupTemplate{
		matchupKey:         newMatchupKey(1, 1),
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	}
	_, err = newBracket([]matchupTemplate{matchTemplate}, newMatchupKey(1, 1), 8)
	if assert.NotNil(t, err) {
		assert.Equal(t, "both alliances must be populated either from selection or a lower round", err.Error())
	}
}

func TestNewBracketInverseSeeding(t *testing.T) {
	database := setupTestDb(t)
	matchupTemplates := []matchupTemplate{
		{
			matchupKey:         newMatchupKey(1, 1),
			LongName:           "Playoff QF1",
			ShortName:          "QF1",
			NumWinsToAdvance:   2,
			redAllianceSource:  allianceSource{allianceId: 8},
			blueAllianceSource: allianceSource{allianceId: 1},
		},
		{
			matchupKey:         newMatchupKey(1, 2),
			LongName:           "Playoff QF2",
			ShortName:          "QF2",
			NumWinsToAdvance:   2,
			redAllianceSource:  allianceSource{allianceId: 5},
			blueAllianceSource: allianceSource{allianceId: 4},
		},
		{
			matchupKey:         newMatchupKey(2, 1),
			LongName:           "Playoff SF1",
			ShortName:          "SF1",
			NumWinsToAdvance:   2,
			redAllianceSource:  newWinnerAllianceSource(1, 2),
			blueAllianceSource: newWinnerAllianceSource(1, 1),
		},
		{
			matchupKey:         newMatchupKey(2, 2),
			LongName:           "Playoff SF2",
			ShortName:          "SF2",
			NumWinsToAdvance:   2,
			redAllianceSource:  allianceSource{allianceId: 3},
			blueAllianceSource: allianceSource{allianceId: 2},
		},
		{
			matchupKey:         newMatchupKey(3, 1),
			LongName:           "Playoff F",
			ShortName:          "F",
			NumWinsToAdvance:   2,
			redAllianceSource:  newWinnerAllianceSource(2, 1),
			blueAllianceSource: newWinnerAllianceSource(2, 2),
		},
	}

	tournament.CreateTestAlliances(database, 2)
	bracket, err := newBracket(matchupTemplates, newMatchupKey(3, 1), 2)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assertMatch(t, matches[0], "Playoff F-1", "F-1", "", 1, 2)
		assertMatch(t, matches[1], "Playoff F-2", "F-2", "", 1, 2)
	}
}

func TestBracketUpdateTiming(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	startTime := time.Unix(1000, 0)
	assert.Nil(t, bracket.Update(database, &startTime))
	matches, err := database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assert.Equal(t, int64(1000), matches[0].Time.Unix())
		assert.Equal(t, int64(1600), matches[1].Time.Unix())
		assert.Equal(t, int64(2200), matches[2].Time.Unix())
		assert.Equal(t, int64(2800), matches[3].Time.Unix())
	}
	scoreMatch(database, "SF1-1", game.RedWonMatch)
	scoreMatch(database, "SF1-2", game.BlueWonMatch)
	startTime = time.Unix(5000, 0)
	assert.Nil(t, bracket.Update(database, &startTime))
	matches, err = database.GetMatchesByType(model.Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 5, len(matches)) {
		assert.Equal(t, int64(1000), matches[0].Time.Unix())
		assert.Equal(t, int64(5000), matches[1].Time.Unix())
		assert.Equal(t, int64(2200), matches[2].Time.Unix())
		assert.Equal(t, int64(5600), matches[3].Time.Unix())
		assert.Equal(t, int64(6200), matches[4].Time.Unix())
	}
}

func TestBracketUpdateTeamPositions(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ := database.GetMatchesByType(model.Playoff)
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
	scoreMatch(database, "SF1-1", game.RedWonMatch)
	scoreMatch(database, "SF2-1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ = database.GetMatchesByType(model.Playoff)
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
	scoreMatch(database, "SF1-2", game.RedWonMatch)
	scoreMatch(database, "SF2-2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, _ = database.GetMatchesByType(model.Playoff)
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

func TestBracketGetAllMatchups(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 5)
	bracket, err := NewSingleEliminationBracket(5)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))

	matchups := bracket.GetAllMatchups()
	if assert.Equal(t, 4, len(matchups)) {
		assert.Equal(t, newMatchupKey(2, 2), matchups[0].matchupKey)
		assert.Equal(t, newMatchupKey(3, 1), matchups[1].matchupKey)
		assert.Equal(t, newMatchupKey(3, 2), matchups[2].matchupKey)
		assert.Equal(t, newMatchupKey(4, 1), matchups[3].matchupKey)
	}
}

func TestBracketGetMatchup(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 4)
	bracket, err := NewSingleEliminationBracket(4)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))

	matchup, err := bracket.GetMatchup(4, 1)
	assert.Nil(t, err)
	assert.Equal(t, newMatchupKey(4, 1), matchup.matchupKey)

	matchup, err = bracket.GetMatchup(3, 2)
	assert.Nil(t, err)
	assert.Equal(t, newMatchupKey(3, 2), matchup.matchupKey)

	matchup, err = bracket.GetMatchup(2, 1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "bracket does not contain matchup for key {Round:2 Group:1}", err.Error())
	}
	assert.Nil(t, matchup)
}

func TestBracketLevelOrderTraversal(t *testing.T) {
	database := setupTestDb(t)
	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewSingleEliminationBracket(8)
	assert.Nil(t, err)

	var shortNames []string
	bracket.ReverseRoundOrderTraversal(func(matchup *Matchup) {
		shortNames = append(shortNames, matchup.ShortName)
	})
	assert.Equal(t, []string{"F", "SF1", "SF2", "QF1", "QF2", "QF3", "QF4"}, shortNames)

	bracket, err = NewDoubleEliminationBracket(8)
	assert.Nil(t, err)

	shortNames = nil
	bracket.ReverseRoundOrderTraversal(func(matchup *Matchup) {
		shortNames = append(shortNames, matchup.ShortName)
	})
	assert.Equal(t, []string{"F", "13", "11", "12", "9", "10", "5", "6", "7", "8", "1", "2", "3", "4"}, shortNames)
}
