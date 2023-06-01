// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchupDisplayNames(t *testing.T) {
	database := setupTestDb(t)
	tournament.CreateTestAlliances(database, 8)
	bracket, err := newDoubleEliminationBracket(database, 8)
	assert.Nil(t, err)

	assert.Equal(t, "Playoff F", bracket.finalMatchup.LongName)
	assert.Equal(t, "F", bracket.finalMatchup.ShortName)
	assert.Equal(t, "-1", bracket.finalMatchup.matchNameSuffix(1))
	assert.Equal(t, "W 11", bracket.finalMatchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W 13", bracket.finalMatchup.BlueAllianceSourceDisplayName())

	match13, err := bracket.GetMatchup(5, 1)
	assert.Nil(t, err)
	assert.Equal(t, "Playoff 13", match13.LongName)
	assert.Equal(t, "13", match13.ShortName)
	assert.Equal(t, "", match13.matchNameSuffix(1))
	assert.Equal(t, "-2", match13.matchNameSuffix(2))
	assert.Equal(t, "L 11", match13.RedAllianceSourceDisplayName())
	assert.Equal(t, "W 12", match13.BlueAllianceSourceDisplayName())

	bracket, err = newSingleEliminationBracket(database, 8)
	assert.Nil(t, err)

	assert.Equal(t, "Playoff F", bracket.finalMatchup.LongName)
	assert.Equal(t, "F", bracket.finalMatchup.ShortName)
	assert.Equal(t, "-1", bracket.finalMatchup.matchNameSuffix(1))
	assert.Equal(t, "W SF1", bracket.finalMatchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W SF2", bracket.finalMatchup.BlueAllianceSourceDisplayName())

	matchSf2, err := bracket.GetMatchup(3, 2)
	assert.Nil(t, err)
	assert.Equal(t, "Playoff SF2", matchSf2.LongName)
	assert.Equal(t, "SF2", matchSf2.ShortName)
	assert.Equal(t, "-1", matchSf2.matchNameSuffix(1))
	assert.Equal(t, "-3", matchSf2.matchNameSuffix(3))
	assert.Equal(t, "W QF3", matchSf2.RedAllianceSourceDisplayName())
	assert.Equal(t, "W QF4", matchSf2.BlueAllianceSourceDisplayName())
}

func TestMatchupStatusText(t *testing.T) {
	matchup := Matchup{matchupTemplate: matchupTemplate{NumWinsToAdvance: 1}}

	leader, status := matchup.StatusText()
	assert.Equal(t, "", leader)
	assert.Equal(t, "", status)

	matchup.RedAllianceWins = 1
	leader, status = matchup.StatusText()
	assert.Equal(t, "red", leader)
	assert.Equal(t, "Red Advances 1-0", status)

	matchup.RedAllianceWins = 0
	matchup.BlueAllianceWins = 2
	leader, status = matchup.StatusText()
	assert.Equal(t, "blue", leader)
	assert.Equal(t, "Blue Advances 2-0", status)

	matchup.NumWinsToAdvance = 3
	matchup.BlueAllianceWins = 2
	leader, status = matchup.StatusText()
	assert.Equal(t, "blue", leader)
	assert.Equal(t, "Blue Leads 2-0", status)

	matchup.RedAllianceWins = 2
	leader, status = matchup.StatusText()
	assert.Equal(t, "", leader)
	assert.Equal(t, "Series Tied 2-2", status)

	matchup.BlueAllianceWins = 1
	leader, status = matchup.StatusText()
	assert.Equal(t, "red", leader)
	assert.Equal(t, "Red Leads 2-1", status)

	matchup.ShortName = "F"
	matchup.RedAllianceWins = 3
	leader, status = matchup.StatusText()
	assert.Equal(t, "red", leader)
	assert.Equal(t, "Red Wins 3-1", status)

	matchup.RedAllianceWins = 2
	matchup.BlueAllianceWins = 4
	leader, status = matchup.StatusText()
	assert.Equal(t, "blue", leader)
	assert.Equal(t, "Blue Wins 4-2", status)

	matchup.RedAllianceWins = 0
	matchup.BlueAllianceWins = 0
	leader, status = matchup.StatusText()
	assert.Equal(t, "", leader)
	assert.Equal(t, "", status)
}
