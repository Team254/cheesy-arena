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
		LongName:           "Playoff 1",
		ShortName:          "1",
		nameDetail:         "Round 1 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 8},
	},
	{
		matchupKey:         newMatchupKey(1, 2),
		LongName:           "Playoff 2",
		ShortName:          "2",
		nameDetail:         "Round 1 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 5},
	},
	{
		matchupKey:         newMatchupKey(1, 3),
		LongName:           "Playoff 3",
		ShortName:          "3",
		nameDetail:         "Round 1 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 7},
	},
	{
		matchupKey:         newMatchupKey(1, 4),
		LongName:           "Playoff 4",
		ShortName:          "4",
		nameDetail:         "Round 1 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 6},
	},
	{
		matchupKey:         newMatchupKey(2, 1),
		LongName:           "Playoff 5",
		ShortName:          "5",
		nameDetail:         "Round 2 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 1),
		blueAllianceSource: newLoserAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 2),
		LongName:           "Playoff 6",
		ShortName:          "6",
		nameDetail:         "Round 2 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 3),
		blueAllianceSource: newLoserAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(2, 3),
		LongName:           "Playoff 7",
		ShortName:          "7",
		nameDetail:         "Round 2 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 1),
		blueAllianceSource: newWinnerAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 4),
		LongName:           "Playoff 8",
		ShortName:          "8",
		nameDetail:         "Round 2 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 3),
		blueAllianceSource: newWinnerAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(3, 1),
		LongName:           "Playoff 9",
		ShortName:          "9",
		nameDetail:         "Round 3 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	},
	{
		matchupKey:         newMatchupKey(3, 2),
		LongName:           "Playoff 10",
		ShortName:          "10",
		nameDetail:         "Round 3 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 4),
		blueAllianceSource: newWinnerAllianceSource(2, 1),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		LongName:           "Playoff 11",
		ShortName:          "11",
		nameDetail:         "Round 4 Upper",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 4),
	},
	{
		matchupKey:         newMatchupKey(4, 2),
		LongName:           "Playoff 12",
		ShortName:          "12",
		nameDetail:         "Round 4 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(3, 2),
		blueAllianceSource: newWinnerAllianceSource(3, 1),
	},
	{
		matchupKey:         newMatchupKey(5, 1),
		LongName:           "Playoff 13",
		ShortName:          "13",
		nameDetail:         "Round 5 Lower",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(4, 1),
		blueAllianceSource: newWinnerAllianceSource(4, 2),
	},
	{
		matchupKey:         newMatchupKey(6, 1),
		LongName:           "Playoff F",
		ShortName:          "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(4, 1),
		blueAllianceSource: newWinnerAllianceSource(5, 1),
	},
}
