// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	AutoFuelPoints                int
	AutoTowerPoints               int
	TeleopFuelPoints              int
	TeleopTowerPoints             int
	NumFuel                       int
	NumFuelPostMatch              int
	NumFuelGoal                   int
	MatchPoints                   int
	PostMatchPoints               int
	FoulPoints                    int
	Score                         int
	PlayoffDq                     bool
	EnergizedBonusRankingPoint    bool
	SuperchargedBonusRankingPoint bool
	TraversalBonusRankingPoint    bool
	BonusRankingPoints            int
	NumOpponentMajorFouls         int
}

type MatchStatus int

const (
	MatchScheduled MatchStatus = iota
	MatchHidden
	RedWonMatch
	BlueWonMatch
	TieMatch
)

func (t MatchStatus) Get() MatchStatus {
	return t
}

// Determines the winner of the match given the score summaries for both alliances, and returns a display string
// indicating the playoff tiebreaker criterion used if the primary score is tied.
func DetermineMatchStatus(
	redScoreSummary, blueScoreSummary *ScoreSummary,
	applyPlayoffTiebreakers bool,
) (MatchStatus, string) {
	if redScoreSummary.PlayoffDq != blueScoreSummary.PlayoffDq {
		if redScoreSummary.PlayoffDq {
			return BlueWonMatch, ""
		}
		return RedWonMatch, ""
	}

	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status, ""
	}

	if applyPlayoffTiebreakers {
		// Check scoring breakdowns to resolve playoff ties.
		if status := comparePoints(
			redScoreSummary.NumOpponentMajorFouls, blueScoreSummary.NumOpponentMajorFouls,
		); status != TieMatch {
			return status, "TIEBREAK: MAJOR FOULS"
		}
		status := comparePoints(redScoreSummary.AutoFuelPoints, blueScoreSummary.AutoFuelPoints)
		if status != TieMatch {
			return status, "TIEBREAK: AUTO FUEL"
		}
		if status = comparePoints(
			redScoreSummary.AutoTowerPoints+redScoreSummary.TeleopTowerPoints,
			blueScoreSummary.AutoTowerPoints+blueScoreSummary.TeleopTowerPoints,
		); status != TieMatch {
			return status, "TIEBREAK: TOWER POINTS"
		}
		return TieMatch, "TRUE TIE"
	}

	return TieMatch, ""
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
