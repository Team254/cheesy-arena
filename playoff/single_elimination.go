// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Defines the tournament structure for a single-elimination, best-of-three bracket.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"strings"
)

// Creates a single-elimination bracket containing only the required matchups for the given number of alliances, and
// returns the root matchup comprising the tournament finals along with scheduled breaks.
func newSingleEliminationBracket(numAlliances int) (*Matchup, []breakSpec, error) {
	if numAlliances < 2 {
		return nil, nil, fmt.Errorf("single-elimination bracket must have at least 2 alliances")
	}
	if numAlliances > 16 {
		return nil, nil, fmt.Errorf("single-elimination bracket must have at most 16 alliances")
	}

	// Define eighthfinal matches.
	ef1 := Matchup{
		id:                 "EF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{1},
		blueAllianceSource: allianceSelectionSource{16},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 1, 1, 1),
			newSingleEliminationMatch("Eighthfinal", "EF", 1, 2, 9),
			newSingleEliminationMatch("Eighthfinal", "EF", 1, 3, 17),
		},
	}
	ef2 := Matchup{
		id:                 "EF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{8},
		blueAllianceSource: allianceSelectionSource{9},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 2, 1, 2),
			newSingleEliminationMatch("Eighthfinal", "EF", 2, 2, 10),
			newSingleEliminationMatch("Eighthfinal", "EF", 2, 3, 18),
		},
	}
	ef3 := Matchup{
		id:                 "EF3",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{4},
		blueAllianceSource: allianceSelectionSource{13},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 3, 1, 3),
			newSingleEliminationMatch("Eighthfinal", "EF", 3, 2, 11),
			newSingleEliminationMatch("Eighthfinal", "EF", 3, 3, 19),
		},
	}
	ef4 := Matchup{
		id:                 "EF4",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{5},
		blueAllianceSource: allianceSelectionSource{12},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 4, 1, 4),
			newSingleEliminationMatch("Eighthfinal", "EF", 4, 2, 12),
			newSingleEliminationMatch("Eighthfinal", "EF", 4, 3, 20),
		},
	}
	ef5 := Matchup{
		id:                 "EF5",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{2},
		blueAllianceSource: allianceSelectionSource{15},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 5, 1, 5),
			newSingleEliminationMatch("Eighthfinal", "EF", 5, 2, 13),
			newSingleEliminationMatch("Eighthfinal", "EF", 5, 3, 21),
		},
	}
	ef6 := Matchup{
		id:                 "EF6",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{7},
		blueAllianceSource: allianceSelectionSource{10},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 6, 1, 6),
			newSingleEliminationMatch("Eighthfinal", "EF", 6, 2, 14),
			newSingleEliminationMatch("Eighthfinal", "EF", 6, 3, 22),
		},
	}
	ef7 := Matchup{
		id:                 "EF7",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{3},
		blueAllianceSource: allianceSelectionSource{14},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 7, 1, 7),
			newSingleEliminationMatch("Eighthfinal", "EF", 7, 2, 15),
			newSingleEliminationMatch("Eighthfinal", "EF", 7, 3, 23),
		},
	}
	ef8 := Matchup{
		id:                 "EF8",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{6},
		blueAllianceSource: allianceSelectionSource{11},
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Eighthfinal", "EF", 8, 1, 8),
			newSingleEliminationMatch("Eighthfinal", "EF", 8, 2, 16),
			newSingleEliminationMatch("Eighthfinal", "EF", 8, 3, 24),
		},
	}

	// Define quarterfinal matches.
	qf1 := Matchup{
		id:                 "QF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&ef1, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&ef2, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Quarterfinal", "QF", 1, 1, 25),
			newSingleEliminationMatch("Quarterfinal", "QF", 1, 2, 29),
			newSingleEliminationMatch("Quarterfinal", "QF", 1, 3, 33),
		},
	}
	qf2 := Matchup{
		id:                 "QF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&ef3, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&ef4, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Quarterfinal", "QF", 2, 1, 26),
			newSingleEliminationMatch("Quarterfinal", "QF", 2, 2, 30),
			newSingleEliminationMatch("Quarterfinal", "QF", 2, 3, 34),
		},
	}
	qf3 := Matchup{
		id:                 "QF3",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&ef5, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&ef6, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Quarterfinal", "QF", 3, 1, 27),
			newSingleEliminationMatch("Quarterfinal", "QF", 3, 2, 31),
			newSingleEliminationMatch("Quarterfinal", "QF", 3, 3, 35),
		},
	}
	qf4 := Matchup{
		id:                 "QF4",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&ef7, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&ef8, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Quarterfinal", "QF", 4, 1, 28),
			newSingleEliminationMatch("Quarterfinal", "QF", 4, 2, 32),
			newSingleEliminationMatch("Quarterfinal", "QF", 4, 3, 36),
		},
	}

	// Define semifinal matches.
	sf1 := Matchup{
		id:                 "SF1",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&qf1, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&qf2, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Semifinal", "SF", 1, 1, 37),
			newSingleEliminationMatch("Semifinal", "SF", 1, 2, 39),
			newSingleEliminationMatch("Semifinal", "SF", 1, 3, 41),
		},
	}
	sf2 := Matchup{
		id:                 "SF2",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&qf3, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&qf4, numAlliances),
		matchSpecs: []*matchSpec{
			newSingleEliminationMatch("Semifinal", "SF", 2, 1, 38),
			newSingleEliminationMatch("Semifinal", "SF", 2, 2, 40),
			newSingleEliminationMatch("Semifinal", "SF", 2, 3, 42),
		},
	}

	// Define final matches.
	final := Matchup{
		id:                 "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  newSingleEliminationAllianceSource(&sf1, numAlliances),
		blueAllianceSource: newSingleEliminationAllianceSource(&sf2, numAlliances),
		matchSpecs:         newFinalMatches(43),
	}

	// Define scheduled breaks.
	breakSpecs := []breakSpec{
		{43, 480, "Field Break"},
		{44, 480, "Field Break"},
		{45, 480, "Field Break"},
	}

	return &final, breakSpecs, nil
}

// Helper method to create an allianceSource while pruning any unnecessary matchups due to the number of alliances.
func newSingleEliminationAllianceSource(matchup *Matchup, numAlliances int) allianceSource {
	redAllianceId := matchup.redAllianceSource.AllianceId()
	blueAllianceId := matchup.blueAllianceSource.AllianceId()

	if blueAllianceId > redAllianceId && blueAllianceId > numAlliances {
		return matchup.redAllianceSource
	}
	if redAllianceId > blueAllianceId && redAllianceId > numAlliances {
		return matchup.blueAllianceSource
	}
	return matchupSource{matchup: matchup, useWinner: true}
}

// Helper method to create a match spec for a pre-final single-elimination matchup.
func newSingleEliminationMatch(longRoundName, shortRoundName string, setNumber, matchNumber, order int) *matchSpec {
	return &matchSpec{
		longName:            fmt.Sprintf("%s %d-%d", longRoundName, setNumber, matchNumber),
		shortName:           fmt.Sprintf("%s%d-%d", shortRoundName, setNumber, matchNumber),
		order:               order,
		durationSec:         600,
		useTiebreakCriteria: true,
		tbaMatchKey:         model.TbaMatchKey{strings.ToLower(shortRoundName), setNumber, matchNumber},
	}
}

// Helper method to create the final matches for any tournament type.
func newFinalMatches(startingOrder int) []*matchSpec {
	return []*matchSpec{
		{
			longName:            "Final 1",
			shortName:           "F1",
			order:               startingOrder,
			durationSec:         300,
			useTiebreakCriteria: false,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 1},
		},
		{
			longName:            "Final 2",
			shortName:           "F2",
			order:               startingOrder + 1,
			durationSec:         300,
			useTiebreakCriteria: false,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 2},
		},
		{
			longName:            "Final 3",
			shortName:           "F3",
			order:               startingOrder + 2,
			durationSec:         300,
			useTiebreakCriteria: false,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 3},
		},
		{
			longName:            "Overtime 1",
			shortName:           "O1",
			order:               startingOrder + 3,
			durationSec:         600,
			useTiebreakCriteria: true,
			isHidden:            true,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 4},
		},
		{
			longName:            "Overtime 2",
			shortName:           "O2",
			order:               startingOrder + 4,
			durationSec:         600,
			useTiebreakCriteria: true,
			isHidden:            true,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 5},
		},
		{
			longName:            "Overtime 3",
			shortName:           "O3",
			order:               startingOrder + 5,
			durationSec:         600,
			useTiebreakCriteria: true,
			isHidden:            true,
			tbaMatchKey:         model.TbaMatchKey{"f", 1, 6},
		},
	}
}
