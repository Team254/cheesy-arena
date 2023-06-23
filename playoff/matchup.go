// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Models and logic encapsulating a group of one or more matches between the same two alliances at a given point in a
// playoff tournament.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
)

type Matchup struct {
	id                 string
	NumWinsToAdvance   int
	redAllianceSource  allianceSource
	blueAllianceSource allianceSource
	matchSpecs         []*matchSpec
	RedAllianceId      int
	BlueAllianceId     int
	RedAllianceWins    int
	BlueAllianceWins   int
	NumMatchesPlayed   int
}

func (matchup *Matchup) Id() string {
	return matchup.id
}

func (matchup *Matchup) MatchSpecs() []*matchSpec {
	return matchup.matchSpecs
}

func (matchup *Matchup) update(playoffMatchResults map[int]playoffMatchResult) {
	// Update child matchups first.
	matchup.redAllianceSource.update(playoffMatchResults)
	matchup.blueAllianceSource.update(playoffMatchResults)

	// Populate the alliance IDs from the lower matchups (or with a zero value if they are not yet complete).
	matchup.RedAllianceId = matchup.redAllianceSource.AllianceId()
	matchup.BlueAllianceId = matchup.blueAllianceSource.AllianceId()

	for _, match := range matchup.matchSpecs {
		match.redAllianceId = matchup.RedAllianceId
		match.blueAllianceId = matchup.BlueAllianceId
	}

	matchup.RedAllianceWins = 0
	matchup.BlueAllianceWins = 0
	matchup.NumMatchesPlayed = 0
	for _, match := range matchup.matchSpecs {
		if matchResult, ok := playoffMatchResults[match.order]; ok {
			switch matchResult.status {
			case game.RedWonMatch:
				matchup.RedAllianceWins++
				matchup.NumMatchesPlayed++
			case game.BlueWonMatch:
				matchup.BlueAllianceWins++
				matchup.NumMatchesPlayed++
			case game.TieMatch:
				matchup.NumMatchesPlayed++
			}
		}
	}
}

func (matchup *Matchup) traverse(visitFunction func(MatchGroup) error) error {
	if err := visitFunction(matchup); err != nil {
		return err
	}
	if err := matchup.redAllianceSource.traverse(visitFunction); err != nil {
		return err
	}
	if err := matchup.blueAllianceSource.traverse(visitFunction); err != nil {
		return err
	}
	return nil
}

// RedAllianceSourceDisplayName returns the display name for the linked matchup from which the red alliance is
// populated.
func (matchup *Matchup) RedAllianceSourceDisplayName() string {
	return matchup.redAllianceSource.displayName()
}

// BlueAllianceSourceDisplayName returns the display name for the linked matchup from which the blue alliance is
// populated.
func (matchup *Matchup) BlueAllianceSourceDisplayName() string {
	return matchup.blueAllianceSource.displayName()
}

// StatusText returns a pair of strings indicating the leading alliance and a readable status of the matchup.
func (matchup *Matchup) StatusText() (string, string) {
	var leader, status string
	winText := "Advances"
	if matchup.isFinal() {
		winText = "Wins"
	}
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		leader = "red"
		status = fmt.Sprintf("Red %s %d-%d", winText, matchup.RedAllianceWins, matchup.BlueAllianceWins)
	} else if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		leader = "blue"
		status = fmt.Sprintf("Blue %s %d-%d", winText, matchup.BlueAllianceWins, matchup.RedAllianceWins)
	} else if matchup.RedAllianceWins > matchup.BlueAllianceWins {
		leader = "red"
		status = fmt.Sprintf("Red Leads %d-%d", matchup.RedAllianceWins, matchup.BlueAllianceWins)
	} else if matchup.BlueAllianceWins > matchup.RedAllianceWins {
		leader = "blue"
		status = fmt.Sprintf("Blue Leads %d-%d", matchup.BlueAllianceWins, matchup.RedAllianceWins)
	} else if matchup.RedAllianceWins > 0 {
		status = fmt.Sprintf("Series Tied %d-%d", matchup.RedAllianceWins, matchup.BlueAllianceWins)
	}
	return leader, status
}

// WinningAllianceId returns the winning alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) WinningAllianceId() int {
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.RedAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.BlueAllianceId
	}
	return 0
}

// LosingAllianceId returns the losing alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) LosingAllianceId() int {
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.BlueAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.RedAllianceId
	}
	return 0
}

// IsComplete returns true if the matchup has been won, and false if it is still to be determined.
func (matchup *Matchup) IsComplete() bool {
	return matchup.WinningAllianceId() > 0
}

// isFinal returns true if the matchup represents the final matchup in the playoff tournament.
func (matchup *Matchup) isFinal() bool {
	return matchup.id == "F"
}
