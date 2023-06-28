// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Models and logic encapsulating the common aspects of all supported playoff tournament formats.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"time"
)

type PlayoffTournament struct {
	matchGroups  map[string]MatchGroup
	matchSpecs   []*matchSpec
	breakSpecs   []breakSpec
	finalMatchup *Matchup
}

// NewPlayoffTournament creates a new playoff tournament of the given type and number of alliances, or returns an error
// if the number of alliances is invalid for the given tournament type.
func NewPlayoffTournament(playoffType model.PlayoffType, numPlayoffAlliances int) (*PlayoffTournament, error) {
	var finalMatchup *Matchup
	var breakSpecs []breakSpec
	var err error
	switch playoffType {
	case model.DoubleEliminationPlayoff:
		finalMatchup, breakSpecs, err = newDoubleEliminationBracket(numPlayoffAlliances)
	case model.SingleEliminationPlayoff:
		finalMatchup, breakSpecs, err = newSingleEliminationBracket(numPlayoffAlliances)
	default:
		err = fmt.Errorf("invalid playoff type: %v", playoffType)
	}
	if err != nil {
		return nil, err
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	if err != nil {
		return nil, err
	}
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	if err != nil {
		return nil, err
	}

	// Doubly link the match group tree in order to populate alliance destinations.
	finalMatchup.setSourceDestinations()

	// Trigger an initial update to populate the alliances.
	finalMatchup.update(map[int]playoffMatchResult{})

	return &PlayoffTournament{
		finalMatchup: finalMatchup,
		matchGroups:  matchGroups,
		matchSpecs:   matchSpecs,
		breakSpecs:   breakSpecs,
	}, nil
}

// MatchGroups returns a map of all match groups in the tournament keyed by ID.
func (tournament *PlayoffTournament) MatchGroups() map[string]MatchGroup {
	return tournament.matchGroups
}

// FinalMatchup returns the matchup representing the tournament's final round.
func (tournament *PlayoffTournament) FinalMatchup() *Matchup {
	return tournament.finalMatchup
}

// IsComplete returns true if the tournament has been won and false if it is still in progress.
func (tournament *PlayoffTournament) IsComplete() bool {
	return tournament.finalMatchup.IsComplete()
}

// WinningAllianceId returns the number of the alliance that won the tournament, or 0 if the tournament is not yet
// complete.
func (tournament *PlayoffTournament) WinningAllianceId() int {
	return tournament.finalMatchup.WinningAllianceId()
}

// FinalistAllianceId returns the number of the alliance that were tournament finalists, or 0 if the tournament is not
// yet complete.
func (tournament *PlayoffTournament) FinalistAllianceId() int {
	return tournament.finalMatchup.LosingAllianceId()
}

// Traverse calls the given function on each match group in the tournament, in reverse round order of play.
func (tournament *PlayoffTournament) Traverse(visitFunction func(MatchGroup) error) error {
	return tournament.finalMatchup.traverse(visitFunction)
}

// CreateMatchesAndBreaks creates all the playoff matches and scheduled breaks in the database, as a one-time action at
// the beginning of the playoff tournament.
func (tournament *PlayoffTournament) CreateMatchesAndBreaks(database *model.Database, startTime time.Time) error {
	matches, err := database.GetMatchesByType(model.Playoff, true)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		return fmt.Errorf("cannot create playoff matches; %d matches already exist", len(matches))
	}
	scheduledBreaks, err := database.GetScheduledBreaksByMatchType(model.Playoff)
	if err != nil {
		return err
	}
	if len(scheduledBreaks) > 0 {
		return fmt.Errorf("cannot create playoff breaks; %d breaks already exist", len(scheduledBreaks))
	}

	alliances, err := database.GetAllAlliances()
	if err != nil {
		return err
	}

	breakIndex := 0
	matchIndex := 0
	nextEventTime := startTime

	for matchIndex < len(tournament.matchSpecs) {
		// Advance the break index past any nonexistent matches.
		for breakIndex < len(tournament.breakSpecs) &&
			tournament.breakSpecs[breakIndex].orderBefore < tournament.matchSpecs[matchIndex].order {
			breakIndex++
		}

		if breakIndex < len(tournament.breakSpecs) &&
			tournament.breakSpecs[breakIndex].orderBefore == tournament.matchSpecs[matchIndex].order {
			// Create the break that is scheduled before the next match.
			breakSpec := tournament.breakSpecs[breakIndex]
			scheduledBreak := model.ScheduledBreak{
				MatchType:       model.Playoff,
				TypeOrderBefore: breakSpec.orderBefore,
				Time:            nextEventTime,
				DurationSec:     breakSpec.durationSec,
				Description:     breakSpec.description,
			}
			if err := database.CreateScheduledBreak(&scheduledBreak); err != nil {
				return err
			}
			breakIndex++
			nextEventTime = nextEventTime.Add(time.Duration(breakSpec.durationSec) * time.Second)
		}

		matchSpec := tournament.matchSpecs[matchIndex]
		match := model.Match{
			Type:                model.Playoff,
			TypeOrder:           matchSpec.order,
			Time:                nextEventTime,
			LongName:            matchSpec.longName,
			ShortName:           matchSpec.shortName,
			NameDetail:          matchSpec.nameDetail,
			PlayoffMatchGroupId: matchSpec.matchGroupId,
			PlayoffRedAlliance:  matchSpec.redAllianceId,
			PlayoffBlueAlliance: matchSpec.blueAllianceId,
			UseTiebreakCriteria: matchSpec.useTiebreakCriteria,
			TbaMatchKey:         matchSpec.tbaMatchKey,
		}
		if match.PlayoffRedAlliance > 0 && len(alliances) >= match.PlayoffRedAlliance {
			positionRedTeams(&match, &alliances[match.PlayoffRedAlliance-1])
		}
		if match.PlayoffBlueAlliance > 0 && len(alliances) >= match.PlayoffBlueAlliance {
			positionBlueTeams(&match, &alliances[match.PlayoffBlueAlliance-1])
		}
		if matchSpec.isHidden {
			match.Status = game.MatchHidden
		} else {
			match.Status = game.MatchScheduled
		}

		if err := database.CreateMatch(&match); err != nil {
			return err
		}

		matchIndex++
		nextEventTime = nextEventTime.Add(time.Duration(matchSpec.durationSec) * time.Second)
	}

	return nil
}

// UpdateMatches updates the playoff matches in the database to assign teams based on the results of the playoff
// tournament so far.
func (tournament *PlayoffTournament) UpdateMatches(database *model.Database) error {
	matches, err := database.GetMatchesByType(model.Playoff, true)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("cannot update playoff matches; no matches exist")
	}

	playoffMatchResults := make(map[int]playoffMatchResult)
	for _, match := range matches {
		switch match.Status {
		case game.RedWonMatch, game.BlueWonMatch, game.TieMatch:
			playoffMatchResults[match.TypeOrder] = playoffMatchResult{status: match.Status}
		}
	}

	tournament.finalMatchup.update(playoffMatchResults)

	// Update all unplayed matches to assign any alliances that have been newly populated into or removed from matches.
	matchesByTypeOrder := make(map[int]*model.Match)
	for i, match := range matches {
		matchesByTypeOrder[match.TypeOrder] = &matches[i]
	}
	alliances, err := database.GetAllAlliances()
	if err != nil {
		return err
	}

	for _, spec := range tournament.matchSpecs {
		match, ok := matchesByTypeOrder[spec.order]
		if !ok {
			return fmt.Errorf("cannot update playoff matches; match with order %d does not exist", spec.order)
		}
		if match.IsComplete() {
			continue
		}

		if spec.isHidden {
			match.Status = game.MatchHidden
		} else {
			match.Status = game.MatchScheduled
		}
		match.PlayoffRedAlliance = spec.redAllianceId
		match.PlayoffBlueAlliance = spec.blueAllianceId
		if match.Status == game.MatchScheduled && match.PlayoffRedAlliance > 0 &&
			len(alliances) >= match.PlayoffRedAlliance {
			positionRedTeams(match, &alliances[match.PlayoffRedAlliance-1])
		} else {
			// Zero out the teams.
			positionRedTeams(match, &model.Alliance{})
		}
		if match.Status == game.MatchScheduled && match.PlayoffBlueAlliance > 0 &&
			len(alliances) >= match.PlayoffBlueAlliance {
			positionBlueTeams(match, &alliances[match.PlayoffBlueAlliance-1])
		} else {
			// Zero out the teams.
			positionBlueTeams(match, &model.Alliance{})
		}
		if err = database.UpdateMatch(match); err != nil {
			return err
		}
	}

	return nil
}

// Assigns the lineup from the alliance into the red team slots for the match.
func positionRedTeams(match *model.Match, alliance *model.Alliance) {
	match.Red1 = alliance.Lineup[0]
	match.Red2 = alliance.Lineup[1]
	match.Red3 = alliance.Lineup[2]
}

// Assigns the lineup from the alliance into the blue team slots for the match.
func positionBlueTeams(match *model.Match, alliance *model.Alliance) {
	match.Blue1 = alliance.Lineup[0]
	match.Blue2 = alliance.Lineup[1]
	match.Blue3 = alliance.Lineup[2]
}
