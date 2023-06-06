// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchupAllianceSourceDisplayNames(t *testing.T) {
	// Test double-elimination.
	matchup, err := newDoubleEliminationBracket(8)
	assert.Nil(t, err)

	assert.Equal(t, "W M11", matchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W M13", matchup.BlueAllianceSourceDisplayName())

	matchGroups, err := collectMatchGroups(matchup)
	assert.Nil(t, err)
	match13 := matchGroups["M13"].(*Matchup)
	assert.Equal(t, "L M11", match13.RedAllianceSourceDisplayName())
	assert.Equal(t, "W M12", match13.BlueAllianceSourceDisplayName())

	// Test single-elimination.
	matchup, err = newSingleEliminationBracket(5)
	assert.Nil(t, err)

	assert.Equal(t, "W SF1", matchup.RedAllianceSourceDisplayName())
	assert.Equal(t, "W SF2", matchup.BlueAllianceSourceDisplayName())

	matchGroups, err = collectMatchGroups(matchup)
	assert.Nil(t, err)
	sf1 := matchGroups["SF1"].(*Matchup)
	assert.Nil(t, err)
	assert.Equal(t, "A 1", sf1.RedAllianceSourceDisplayName())
	assert.Equal(t, "W QF2", sf1.BlueAllianceSourceDisplayName())
}

func TestMatchupStatusText(t *testing.T) {
	matchup := Matchup{NumWinsToAdvance: 1}

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

	matchup.id = "F"
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
