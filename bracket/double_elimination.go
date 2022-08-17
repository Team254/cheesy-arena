// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Defines the tournament structure for a double-elimination bracket culminating in a best-of-three final.

package bracket

import "fmt"

// Creates an unpopulated double-elimination bracket. Only supports having exactly eight alliances.
func NewDoubleEliminationBracket(numAlliances int) (*Bracket, error) {
	if numAlliances != 8 {
		return nil, fmt.Errorf("Must have exactly 8 alliances")
	}
	return newBracket(doubleEliminationBracketMatchupTemplates, newMatchupKey(6, 1), numAlliances)
}

var doubleEliminationBracketMatchupTemplates = []matchupTemplate{
	{
		matchupKey:         newMatchupKey(1, 1),
		displayNameFormat:  "1",
		numWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 8},
	},
	{
		matchupKey:         newMatchupKey(1, 2),
		displayNameFormat:  "2",
		numWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 5},
	},
	{
		matchupKey:         newMatchupKey(1, 3),
		displayNameFormat:  "3",
		numWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 6},
	},
	{
		matchupKey:         newMatchupKey(1, 4),
		displayNameFormat:  "4",
		numWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 7},
	},
	{
		matchupKey:         newMatchupKey(2, 1),
		displayNameFormat:  "5",
		numWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 1),
		blueAllianceSource: newLoserAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 2),
		displayNameFormat:  "6",
		numWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 3),
		blueAllianceSource: newLoserAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(2, 3),
		displayNameFormat:  "7",
		numWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 1),
		blueAllianceSource: newWinnerAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 4),
		displayNameFormat:  "8",
		numWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 3),
		blueAllianceSource: newWinnerAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(3, 1),
		displayNameFormat:  "9",
		numWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	},
	{
		matchupKey:         newMatchupKey(3, 2),
		displayNameFormat:  "10",
		numWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 4),
		blueAllianceSource: newWinnerAllianceSource(2, 1),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		displayNameFormat:  "11",
		numWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(3, 1),
		blueAllianceSource: newWinnerAllianceSource(3, 2),
	},
	{
		matchupKey:         newMatchupKey(4, 2),
		displayNameFormat:  "12",
		numWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 4),
	},
	{
		matchupKey:         newMatchupKey(5, 1),
		displayNameFormat:  "13",
		numWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(4, 1),
	},
	{
		matchupKey:         newMatchupKey(6, 1),
		displayNameFormat:  "F-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(5, 1),
	},
}
