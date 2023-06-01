// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and logic encapsulating a playoff elimination bracket.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"sort"
	"time"
)

type Bracket struct {
	database     *model.Database
	finalMatchup *Matchup
	matchupMap   map[matchupKey]*Matchup
}

const PlayoffMatchSpacingSec = 600

// Creates an unpopulated bracket with a format that is defined by the given matchup templates and number of alliances.
func newBracket(
	database *model.Database, matchupTemplates []matchupTemplate, finalsMatchupKey matchupKey, numAlliances int,
) (*Bracket, error) {
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

	return &Bracket{database: database, finalMatchup: finalsMatchup, matchupMap: matchupMap}, nil
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

func (bracket *Bracket) IsComplete() bool {
	return bracket.finalMatchup.IsComplete()
}

func (bracket *Bracket) WinningAlliance() int {
	return bracket.finalMatchup.Winner()
}

func (bracket *Bracket) FinalistAlliance() int {
	return bracket.finalMatchup.Loser()
}

func (bracket *Bracket) GetAllMatchups() []*Matchup {
	var matchups []*Matchup
	for _, matchup := range bracket.matchupMap {
		matchups = append(matchups, matchup)
	}
	sort.Slice(matchups, func(i, j int) bool {
		if matchups[i].Round == matchups[j].Round {
			return matchups[i].Group < matchups[j].Group
		}
		return matchups[i].Round < matchups[j].Round
	})
	return matchups
}

func (bracket *Bracket) GetMatchup(round, group int) (*Matchup, error) {
	matchupKey := newMatchupKey(round, group)
	if matchup, ok := bracket.matchupMap[matchupKey]; ok {
		return matchup, nil
	}
	return nil, fmt.Errorf("bracket does not contain matchup for key %+v", matchupKey)
}

func (bracket *Bracket) FinalMatchup() *Matchup {
	return bracket.finalMatchup
}

func (bracket *Bracket) Update(startTime *time.Time) error {
	if err := bracket.finalMatchup.update(bracket.database); err != nil {
		return err
	}

	if startTime != nil {
		// Update the scheduled time for all matches that have yet to be run.
		matches, err := bracket.database.GetMatchesByType(model.Playoff)
		if err != nil {
			return err
		}
		matchIndex := 0
		for _, match := range matches {
			if match.IsComplete() {
				continue
			}
			match.Time = startTime.Add(time.Duration(matchIndex*PlayoffMatchSpacingSec) * time.Second)
			if err = bracket.database.UpdateMatch(&match); err != nil {
				return err
			}
			matchIndex++
		}
	}

	return nil
}

func (bracket *Bracket) ReverseRoundOrderTraversal(visitFunction func(*Matchup)) {
	matchupQueue := []*Matchup{bracket.finalMatchup}
	for len(matchupQueue) > 0 {
		// Reorder the queue since graph depth doesn't necessarily equate to round.
		sort.Slice(matchupQueue, func(i, j int) bool {
			if matchupQueue[i].Round == matchupQueue[j].Round {
				return matchupQueue[i].Group < matchupQueue[j].Group
			}
			return matchupQueue[i].Round > matchupQueue[j].Round
		})
		matchup := matchupQueue[0]
		visitFunction(matchup)
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

// Prints out each matchup within the bracket in level order, backwards from finals to earlier rounds, for debugging.
func (bracket *Bracket) print() {
	bracket.ReverseRoundOrderTraversal(func(matchup *Matchup) {
		fmt.Printf("%+v\n\n", matchup)
	})
}
