// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific fields by which teams are ranked and the logic for sorting rankings.

package game

import "math/rand"

type RankingFields struct {
	RankingPoints        int
	CargoPoints          int
	HatchPanelPoints     int
	HabClimbPoints       int
	SandstormBonusPoints int
	Random               float64
	Wins                 int
	Losses               int
	Ties                 int
	Disqualifications    int
	Played               int
}

type Ranking struct {
	TeamId int
	Rank   int
	RankingFields
}

type Rankings []*Ranking

func (fields *RankingFields) AddScoreSummary(ownScore *ScoreSummary, opponentScore *ScoreSummary, disqualified bool) {
	fields.Played += 1

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
	if ownScore.CompleteRocket {
		fields.RankingPoints += 1
	}
	if ownScore.HabDocking {
		fields.RankingPoints += 1
	}

	// Assign tiebreaker points.
	fields.CargoPoints += ownScore.CargoPoints
	fields.HatchPanelPoints += ownScore.HatchPanelPoints
	fields.HabClimbPoints += ownScore.HabClimbPoints
	fields.SandstormBonusPoints += ownScore.SandstormBonusPoints

	// Store a random value to be used as the last tiebreaker if necessary.
	fields.Random = rand.Float64()
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
		if a.CargoPoints*b.Played == b.CargoPoints*a.Played {
			if a.HatchPanelPoints*b.Played == b.HatchPanelPoints*a.Played {
				if a.HabClimbPoints*b.Played == b.HabClimbPoints*a.Played {
					if a.SandstormBonusPoints*b.Played == b.SandstormBonusPoints*a.Played {
						return a.Random > b.Random
					}
					return a.SandstormBonusPoints*b.Played > b.SandstormBonusPoints*a.Played
				}
				return a.HabClimbPoints*b.Played > b.HabClimbPoints*a.Played
			}
			return a.HatchPanelPoints*b.Played > b.HatchPanelPoints*a.Played
		}
		return a.CargoPoints*b.Played > b.CargoPoints*a.Played
	}
	return a.RankingPoints*b.Played > b.RankingPoints*a.Played
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Swap(i, j int) {
	rankings[i], rankings[j] = rankings[j], rankings[i]
}
