// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	TaxiPoints              int
	AutoCargoCount          int
	AutoCargoPoints         int
	CargoCount              int
	CargoPoints             int
	HangarPoints            int
	MatchPoints             int
	FoulPoints              int
	Score                   int
	AutoCargoRemaining      int
	TeleopCargoRemaining    int
	QuintetAchieved         bool
	CargoBonusRankingPoint  bool
	HangarBonusRankingPoint bool
	DoubleBonusRankingPoint bool
}

type MatchStatus string

const (
	RedWonMatch    MatchStatus = "R"
	BlueWonMatch   MatchStatus = "B"
	TieMatch       MatchStatus = "T"
	MatchNotPlayed MatchStatus = ""
)

// Determines the winner of the match given the score summaries for both alliances.
func DetermineMatchStatus(redScoreSummary, blueScoreSummary *ScoreSummary, applyElimTiebreakers bool) MatchStatus {
	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status
	}

	if applyElimTiebreakers {
		// Check scoring breakdowns to resolve playoff ties.
		if status := comparePoints(redScoreSummary.FoulPoints, blueScoreSummary.FoulPoints); status != TieMatch {
			return status
		}
		if status := comparePoints(redScoreSummary.HangarPoints, blueScoreSummary.HangarPoints); status != TieMatch {
			return status
		}
		if status := comparePoints(
			redScoreSummary.TaxiPoints+redScoreSummary.AutoCargoPoints,
			blueScoreSummary.TaxiPoints+blueScoreSummary.AutoCargoPoints,
		); status != TieMatch {
			return status
		}
	}

	return TieMatch
}

// Helper method to compare the red and blue alliance point totals and return the appropriate MatchStatus.
func comparePoints(redPoints, bluePoints int) MatchStatus {
	if redPoints > bluePoints {
		return RedWonMatch
	}
	if redPoints < bluePoints {
		return BlueWonMatch
	}
	return TieMatch
}
