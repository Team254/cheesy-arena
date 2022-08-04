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
	return newBracket(singleEliminationBracketMatchupTemplates, numAlliances)
}

var singleEliminationBracketMatchupTemplates = []matchupTemplate{
	{
		matchupKey:         newMatchupKey(1, 1),
		displayNameFormat:  "F-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(2, 1),
		blueAllianceSource: newMatchupAllianceSource(2, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 1),
		displayNameFormat:  "SF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(4, 1),
		blueAllianceSource: newMatchupAllianceSource(4, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 2),
		displayNameFormat:  "SF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(4, 3),
		blueAllianceSource: newMatchupAllianceSource(4, 4),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		displayNameFormat:  "QF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(8, 1),
		blueAllianceSource: newMatchupAllianceSource(8, 2),
	},
	{
		matchupKey:         newMatchupKey(4, 2),
		displayNameFormat:  "QF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(8, 3),
		blueAllianceSource: newMatchupAllianceSource(8, 4),
	},
	{
		matchupKey:         newMatchupKey(4, 3),
		displayNameFormat:  "QF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(8, 5),
		blueAllianceSource: newMatchupAllianceSource(8, 6),
	},
	{
		matchupKey:         newMatchupKey(4, 4),
		displayNameFormat:  "QF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  newMatchupAllianceSource(8, 7),
		blueAllianceSource: newMatchupAllianceSource(8, 8),
	},
	{
		matchupKey:         newMatchupKey(8, 1),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 16},
	},
	{
		matchupKey:         newMatchupKey(8, 2),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 8},
		blueAllianceSource: allianceSource{allianceId: 9},
	},
	{
		matchupKey:         newMatchupKey(8, 3),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 13},
	},
	{
		matchupKey:         newMatchupKey(8, 4),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 5},
		blueAllianceSource: allianceSource{allianceId: 12},
	},
	{
		matchupKey:         newMatchupKey(8, 5),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 15},
	},
	{
		matchupKey:         newMatchupKey(8, 6),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 7},
		blueAllianceSource: allianceSource{allianceId: 10},
	},
	{
		matchupKey:         newMatchupKey(8, 7),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 14},
	},
	{
		matchupKey:         newMatchupKey(8, 8),
		displayNameFormat:  "EF${group}-${instance}",
		numWinsToAdvance:   2,
		redAllianceSource:  allianceSource{allianceId: 6},
		blueAllianceSource: allianceSource{allianceId: 11},
	},
}
