// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and logic encapsulating a playoff elimination bracket.

package bracket

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"time"
)

type Bracket struct {
	FinalsMatchup *Matchup
	matchupMap    map[matchupKey]*Matchup
}

const ElimMatchSpacingSec = 600

// Creates an unpopulated bracket with a format that is defined by the given matchup templates and number of alliances.
func newBracket(matchupTemplates []matchupTemplate, finalsMatchupKey matchupKey, numAlliances int) (*Bracket, error) {
	// Create a map of matchup templates by key for easy lookup while creating the bracket.
	matchupTemplateMap := make(map[matchupKey]matchupTemplate, len(matchupTemplates))
	for _, matchupTemplate := range matchupTemplates {
		matchupTemplateMap[matchupTemplate.matchupKey] = matchupTemplate
	}

	// Recursively build the bracket, starting with the finals matchup.
	matchupMap := make(map[matchupKey]*Matchup)
	finalsMatchup, _, err := createMatchupGraph(finalsMatchupKey, true, matchupTemplateMap, numAlliances, matchupMap)
	if err != nil {
		return nil, err
	}

	return &Bracket{FinalsMatchup: finalsMatchup, matchupMap: matchupMap}, nil
}

// Recursive helper method to create the current matchup node and all of its children.
func createMatchupGraph(
	matchupKey matchupKey,
	useWinner bool,
	matchupTemplateMap map[matchupKey]matchupTemplate,
	numAlliances int,
	matchupMap map[matchupKey]*Matchup,
) (*Matchup, int, error) {
	matchupTemplate, ok := matchupTemplateMap[matchupKey]
	if !ok {
		return nil, 0, fmt.Errorf("could not find template for matchup %+v in the list of templates", matchupKey)
	}

	redAllianceIdFromSelection := matchupTemplate.redAllianceSource.allianceId
	blueAllianceIdFromSelection := matchupTemplate.blueAllianceSource.allianceId
	if redAllianceIdFromSelection > 0 || blueAllianceIdFromSelection > 0 {
		// This is a leaf node in the matchup graph; the alliances will come from the alliance selection.
		if redAllianceIdFromSelection == 0 || blueAllianceIdFromSelection == 0 {
			return nil, 0, fmt.Errorf("both alliances must be populated either from selection or a lower round")
		}

		// Zero out alliance IDs that don't exist at this tournament to signal that this matchup doesn't need to be
		// played.
		if redAllianceIdFromSelection > numAlliances {
			redAllianceIdFromSelection = 0
		}
		if blueAllianceIdFromSelection > numAlliances {
			blueAllianceIdFromSelection = 0
		}

		if redAllianceIdFromSelection > 0 && blueAllianceIdFromSelection > 0 {
			// This is a real matchup that will be played out.
			matchup, ok := matchupMap[matchupKey]
			if !ok {
				matchup = &Matchup{
					matchupTemplate: matchupTemplate,
					RedAllianceId:   redAllianceIdFromSelection,
					BlueAllianceId:  blueAllianceIdFromSelection,
				}
				matchupMap[matchupKey] = matchup
			}
			return matchup, 0, nil
		}
		if redAllianceIdFromSelection == 0 && blueAllianceIdFromSelection == 0 {
			// This matchup should be pruned from the bracket since neither alliance has a valid source; this tournament
			// is too small for this matchup to be played.
			return nil, 0, nil
		}
		if useWinner {
			if redAllianceIdFromSelection > 0 {
				// The red alliance has a bye.
				return nil, redAllianceIdFromSelection, nil
			} else {
				// The blue alliance has a bye.
				return nil, blueAllianceIdFromSelection, nil
			}
		} else {
			// There is no losing alliance to return; prune this matchup.
			return nil, 0, nil
		}
	}

	// Recurse to determine the lower-round red and blue matchups that will feed into this one, or the alliances that
	// have a bye to this round.
	redAllianceSourceMatchup, redByeAllianceId, err := createMatchupGraph(
		matchupTemplate.redAllianceSource.matchupKey,
		matchupTemplate.redAllianceSource.useWinner,
		matchupTemplateMap,
		numAlliances,
		matchupMap,
	)
	if err != nil {
		return nil, 0, err
	}
	blueAllianceSourceMatchup, blueByeAllianceId, err := createMatchupGraph(
		matchupTemplate.blueAllianceSource.matchupKey,
		matchupTemplate.blueAllianceSource.useWinner,
		matchupTemplateMap,
		numAlliances,
		matchupMap,
	)
	if err != nil {
		return nil, 0, err
	}

	if redAllianceSourceMatchup == nil && redByeAllianceId == 0 &&
		blueAllianceSourceMatchup == nil && blueByeAllianceId == 0 {
		// This matchup should be pruned from the bracket since neither alliance has a valid source; this tournament is
		// too small for this matchup to be played.
		return nil, 0, nil
	}
	if redByeAllianceId > 0 && blueAllianceSourceMatchup == nil && blueByeAllianceId == 0 {
		if useWinner {
			// The red alliance has a bye.
			return nil, redByeAllianceId, nil
		} else {
			// There is no losing alliance to return; prune this matchup.
			return nil, 0, nil
		}
	}
	if blueByeAllianceId > 0 && redAllianceSourceMatchup == nil && redByeAllianceId == 0 {
		if useWinner {
			// The blue alliance has a bye.
			return nil, blueByeAllianceId, nil
		} else {
			// There is no losing alliance to return; prune this matchup.
			return nil, 0, nil
		}
	}

	// This is a real matchup that will be played out.
	matchup, ok := matchupMap[matchupKey]
	if !ok {
		matchup = &Matchup{
			matchupTemplate:           matchupTemplate,
			RedAllianceId:             redByeAllianceId,
			BlueAllianceId:            blueByeAllianceId,
			redAllianceSourceMatchup:  redAllianceSourceMatchup,
			blueAllianceSourceMatchup: blueAllianceSourceMatchup,
		}
		matchupMap[matchupKey] = matchup
	}
	return matchup, 0, nil
}

// Returns the winning alliance ID of the entire bracket, or 0 if it is not yet known.
func (bracket *Bracket) Winner() int {
	return bracket.FinalsMatchup.winner()
}

// Returns the finalist alliance ID of the entire bracket, or 0 if it is not yet known.
func (bracket *Bracket) Finalist() int {
	return bracket.FinalsMatchup.loser()
}

// Returns true if the bracket has been won, and false if it is still to be determined.
func (bracket *Bracket) IsComplete() bool {
	return bracket.FinalsMatchup.isComplete()
}

// Returns the matchup for the given round and group, or an error if it doesn't exist within the bracket.
func (bracket *Bracket) GetMatchup(round, group int) (*Matchup, error) {
	matchupKey := newMatchupKey(round, group)
	if matchup, ok := bracket.matchupMap[matchupKey]; ok {
		return matchup, nil
	}
	return nil, fmt.Errorf("bracket does not contain matchup for key %+v", matchupKey)
}

// Traverses the bracket to update the state of each matchup based on match results, counting wins and creating or
// deleting matches as required.
func (bracket *Bracket) Update(database *model.Database, startTime *time.Time) error {
	if err := bracket.FinalsMatchup.update(database); err != nil {
		return err
	}

	if startTime != nil {
		// Update the scheduled time for all matches that have yet to be run.
		matches, err := database.GetMatchesByType("elimination")
		if err != nil {
			return err
		}
		matchIndex := 0
		for _, match := range matches {
			if match.IsComplete() {
				continue
			}
			match.Time = startTime.Add(time.Duration(matchIndex*ElimMatchSpacingSec) * time.Second)
			if err = database.UpdateMatch(&match); err != nil {
				return err
			}
			matchIndex++
		}
	}

	return nil
}

// Prints out each matchup within the bracket in level order, backwards from finals to earlier rounds, for debugging.
func (bracket *Bracket) print() {
	matchupQueue := []*Matchup{bracket.FinalsMatchup}
	for len(matchupQueue) > 0 {
		matchup := matchupQueue[0]
		fmt.Printf("%+v\n\n", matchup)
		matchupQueue = matchupQueue[1:]
		if matchup != nil {
			if matchup.redAllianceSourceMatchup != nil && matchup.redAllianceSource.useWinner {
				matchupQueue = append(matchupQueue, matchup.redAllianceSourceMatchup)
			}
			if matchup.blueAllianceSourceMatchup != nil && matchup.blueAllianceSource.useWinner {
				matchupQueue = append(matchupQueue, matchup.blueAllianceSourceMatchup)
			}
		}
	}
}
