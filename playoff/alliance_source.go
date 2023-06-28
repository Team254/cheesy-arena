// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents how the alliance is determined to fill a given spot at a given stage in a playoff tournament.

package playoff

import "fmt"

type allianceSource interface {
	// AllianceId returns the alliance number that will fill this spot, or zero if it is not yet determined.
	AllianceId() int

	// displayName returns a human-readable name for the source of this alliance.
	displayName() string

	// setDestination passes back the match group filled by this alliance source to the source.
	setDestination(destination MatchGroup)

	// update updates the state of each match group based on the results of the given played matches.
	update(playoffMatchResults map[int]playoffMatchResult)

	// traverse performs a depth-first traversal of the playoff graph and invokes the given function before visiting
	// each match group's children.
	traverse(visitFunction func(MatchGroup) error) error
}

// Represents a playoff spot that is filled directly from the alliance selection.
type allianceSelectionSource struct {
	allianceId int
}

func (source allianceSelectionSource) AllianceId() int {
	return source.allianceId
}

func (source allianceSelectionSource) displayName() string {
	return fmt.Sprintf("A %d", source.allianceId)
}

func (source allianceSelectionSource) setDestination(destination MatchGroup) {
	// Do nothing as there are no child match groups.
}

func (source allianceSelectionSource) update(playoffMatchResults map[int]playoffMatchResult) {
	// Do nothing as there are no child match groups.
}

func (source allianceSelectionSource) traverse(visitFunction func(MatchGroup) error) error {
	// Do nothing as there are no child match groups.
	return nil
}

// Represents a playoff spot that is filled by the winner or loser of a given earlier matchup.
type matchupSource struct {
	matchup   *Matchup
	useWinner bool
}

func (source matchupSource) AllianceId() int {
	if source.useWinner {
		return source.matchup.WinningAllianceId()
	} else {
		return source.matchup.LosingAllianceId()
	}
}

func (source matchupSource) displayName() string {
	if source.useWinner {
		return "W " + source.matchup.Id()
	}
	return "L " + source.matchup.Id()
}

func (source matchupSource) setDestination(destination MatchGroup) {
	if source.useWinner {
		source.matchup.winningAllianceDestination = destination
	} else {
		source.matchup.losingAllianceDestination = destination
	}

	// Recurse down through the playoff tournament tree.
	source.matchup.setSourceDestinations()
}

func (source matchupSource) update(playoffMatchResults map[int]playoffMatchResult) {
	// Only update if this source is for the winner, to avoid visiting the same match group more than once.
	if source.useWinner {
		source.matchup.update(playoffMatchResults)
	}
}

func (source matchupSource) traverse(visitFunction func(MatchGroup) error) error {
	// Only traverse if this source is for the winner, to avoid visiting the same match group more than once.
	if source.useWinner {
		return source.matchup.traverse(visitFunction)
	}
	return nil
}
