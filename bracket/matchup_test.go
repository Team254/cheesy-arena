// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package bracket

import (
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchupDisplayNames(t *testing.T) {
	database := setupTestDb(t)
	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)

	assert.Equal(t, "Finals", bracket.FinalsMatchup.LongDisplayName())
	assert.Equal(t, "F-1", bracket.FinalsMatchup.matchDisplayName(1))
	assert.Equal(t, "W 11", bracket.FinalsMatchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W 13", bracket.FinalsMatchup.BlueAllianceSourceDisplayName())

	match13, err := bracket.GetMatchup(5, 1)
	assert.Nil(t, err)
	assert.Equal(t, "Match 13", match13.LongDisplayName())
	assert.Equal(t, "13", match13.matchDisplayName(1))
	assert.Equal(t, "13-2", match13.matchDisplayName(2))
	assert.Equal(t, "L 11", match13.RedAllianceSourceDisplayName())
	assert.Equal(t, "W 12", match13.BlueAllianceSourceDisplayName())

	bracket, err = NewSingleEliminationBracket(8)
	assert.Nil(t, err)

	assert.Equal(t, "Finals", bracket.FinalsMatchup.LongDisplayName())
	assert.Equal(t, "F-1", bracket.FinalsMatchup.matchDisplayName(1))
	assert.Equal(t, "W SF1", bracket.FinalsMatchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W SF2", bracket.FinalsMatchup.BlueAllianceSourceDisplayName())

	matchSf2, err := bracket.GetMatchup(3, 2)
	assert.Nil(t, err)
	assert.Equal(t, "SF2", matchSf2.LongDisplayName())
	assert.Equal(t, "SF2-1", matchSf2.matchDisplayName(1))
	assert.Equal(t, "SF2-3", matchSf2.matchDisplayName(3))
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

	matchup.displayName = "F"
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
