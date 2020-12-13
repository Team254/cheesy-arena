// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific fields by which teams are ranked and the logic for sorting rankings.

package game

import "math/rand"

type RankingFields struct {
	RankingPoints     int
	AutoPoints        int
	EndgamePoints     int
	TeleopPoints      int
	Random            float64
	Wins              int
	Losses            int
	Ties              int
	Disqualifications int
	Played            int
}

type Ranking struct {
	TeamId       int
	Rank         int
	PreviousRank int
	RankingFields
}

type Rankings []Ranking

func (fields *RankingFields) AddScoreSummary(ownScore *ScoreSummary, opponentScore *ScoreSummary, disqualified bool) {
	fields.Played += 1

	// Store a random value to be used as the last tiebreaker if necessary.
	fields.Random = rand.Float64()

	if disqualified {
		// Don't award any points.
		fields.Disqualifications += 1
		return
	}

	// Assign ranking points and wins/losses/ties.
	if ownScore.Score > opponentScore.Score {
		fields.RankingPoints += 2
		fields.Wins += 1
	} else if ownScore.Score == opponentScore.Score {
		fields.RankingPoints += 1
		fields.Ties += 1
	} else {
		fields.Losses += 1
	}
	if ownScore.ControlPanelRankingPoint {
		fields.RankingPoints += 1
	}
	if ownScore.EndgameRankingPoint {
		fields.RankingPoints += 1
	}

	// Assign tiebreaker points.
	fields.AutoPoints += ownScore.AutoPoints
	fields.EndgamePoints += ownScore.EndgamePoints
	fields.TeleopPoints += ownScore.TeleopPowerCellPoints + ownScore.ControlPanelPoints
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Len() int {
	return len(rankings)
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Less(i, j int) bool {
	a := rankings[i]
	b := rankings[j]

	// Use cross-multiplication to keep it in integer math.
	if a.RankingPoints*b.Played == b.RankingPoints*a.Played {
		if a.AutoPoints*b.Played == b.AutoPoints*a.Played {
			if a.EndgamePoints*b.Played == b.EndgamePoints*a.Played {
				if a.TeleopPoints*b.Played == b.TeleopPoints*a.Played {
					return a.Random > b.Random
				}
				return a.TeleopPoints*b.Played > b.TeleopPoints*a.Played
			}
			return a.EndgamePoints*b.Played > b.EndgamePoints*a.Played
		}
		return a.AutoPoints*b.Played > b.AutoPoints*a.Played
	}
	return a.RankingPoints*b.Played > b.RankingPoints*a.Played
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Swap(i, j int) {
	rankings[i], rankings[j] = rankings[j], rankings[i]
}
