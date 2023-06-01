// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Interface representing a generic playoff tournament of any format.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"time"
)

type PlayoffTournament interface {
	// GetAllMatchups returns all the matchups in the tournament.
	GetAllMatchups() []*Matchup

	// GetMatchup returns the matchup for the given round and group, or an error if it doesn't exist.
	GetMatchup(round, group int) (*Matchup, error)

	// FinalMatchup returns the matchup representing the tournament's final round.
	FinalMatchup() *Matchup

	// IsComplete returns true if the tournament has been won and false if it is still in progress.
	IsComplete() bool

	// WinningAlliance returns the number of the alliance that won the tournament, or 0 if the tournament is not yet
	// complete.
	WinningAlliance() int

	// FinalistAlliance returns the number of the alliance that were tournament finalists, or 0 if the tournament is not
	// yet complete.
	FinalistAlliance() int

	// Update traverses the tournament to update the state of each unplayed match based on the results of prior matches.
	Update(startTime *time.Time) error

	// ReverseRoundOrderTraversal performs a traversal of the tournament in reverse order of rounds and invokes the
	// given function for each visited matchup.
	// TODO(pat): Don't expose this method directly.
	ReverseRoundOrderTraversal(visitFunction func(*Matchup))
}

// NewPlayoffTournament creates a new playoff tournament of the given type and number of alliances, or returns an error
// if the number of alliances is invalid for the given tournament type.
func NewPlayoffTournament(
	database *model.Database, playoffType model.PlayoffType, numPlayoffAlliances int,
) (playoffTournament PlayoffTournament, err error) {
	switch playoffType {
	case model.DoubleEliminationPlayoff:
		playoffTournament, err = newDoubleEliminationBracket(database, numPlayoffAlliances)
	case model.SingleEliminationPlayoff:
		playoffTournament, err = newSingleEliminationBracket(database, numPlayoffAlliances)
	default:
		err = fmt.Errorf("invalid playoff type: %v", playoffType)
	}
	return playoffTournament, err
}
