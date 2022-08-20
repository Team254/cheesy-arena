// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Defines the tournament structure for a single-elimination, best-of-three bracket.

package bracket

import "fmt"

// Creates an unpopulated single-elimination bracket containing only the required matchups for the given number of
// alliances.
func NewSingleEliminationBracket(numAlliances int) (*Bracket, error) {
	if numAlliances < 2 {
		return nil, fmt.Errorf("Must have at least 2 alliances")
	}
	if numAlliances > 16 {
		return nil, fmt.Errorf("Must have at most 16 alliances")
	}
	return newBracket(singleEliminationBracketMatchupTemplates, newMatchupKey(4, 1), numAlliances)
}

var singleEliminationBracketMatchupTemplates = []matchupTemplate{
	{
		matchupKey:         newMatchupKey(1, 1),
		displayName:        "EF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 16},
	},
	{
		matchupKey:         newMatchupKey(1, 2),
		displayName:        "EF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 8},
		blueAllianceSource: allianceSource{allianceId: 9},
	},
	{
		matchupKey:         newMatchupKey(1, 3),
		displayName:        "EF3",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 13},
	},
	{
		matchupKey:         newMatchupKey(1, 4),
		displayName:        "EF4",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 5},
		blueAllianceSource: allianceSource{allianceId: 12},
	},
	{
		matchupKey:         newMatchupKey(1, 5),
		displayName:        "EF5",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 15},
	},
	{
		matchupKey:         newMatchupKey(1, 6),
		displayName:        "EF6",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 7},
		blueAllianceSource: allianceSource{allianceId: 10},
	},
	{
		matchupKey:         newMatchupKey(1, 7),
		displayName:        "EF7",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 14},
	},
	{
		matchupKey:         newMatchupKey(1, 8),
		displayName:        "EF8",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 6},
		blueAllianceSource: allianceSource{allianceId: 11},
	},
	{
		matchupKey:         newMatchupKey(2, 1),
		displayName:        "QF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(1, 1),
		blueAllianceSource: newWinnerAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 2),
		displayName:        "QF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(1, 3),
		blueAllianceSource: newWinnerAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(2, 3),
		displayName:        "QF3",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(1, 5),
		blueAllianceSource: newWinnerAllianceSource(1, 6),
	},
	{
		matchupKey:         newMatchupKey(2, 4),
		displayName:        "QF4",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(1, 7),
		blueAllianceSource: newWinnerAllianceSource(1, 8),
	},
	{
		matchupKey:         newMatchupKey(3, 1),
		displayName:        "SF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(2, 1),
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	},
	{
		matchupKey:         newMatchupKey(3, 2),
		displayName:        "SF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 4),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		displayName:        "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(3, 1),
		blueAllianceSource: newWinnerAllianceSource(3, 2),
	},
}
