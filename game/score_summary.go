// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	// --- Common Fields ---
	MatchPoints           int
	FoulPoints            int
	Score                 int
	AutoPoints            int
	BonusRankingPoints    int
	NumOpponentMajorFouls int

	// --- 2026 New Fields ---
	// Fuel Points (Balls)
	AutoFuelPoints   int
	TeleopFuelPoints int
	TotalFuelPoints  int

	// Tower Points (Climbing)
	AutoTowerPoints    int
	EndgameTowerPoints int
	TotalTowerPoints   int

	// Ranking Points Status
	EnergizedRankingPoint    bool // Based on Fuel count
	SuperchargedRankingPoint bool // Based on higher Fuel count
	TraversalRankingPoint    bool // Based on Tower points
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

// Determines the winner of the match given the score summaries for both alliances.
func DetermineMatchStatus(redScoreSummary, blueScoreSummary *ScoreSummary, applyPlayoffTiebreakers bool) MatchStatus {
	// 1. Compare Total Score
	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status
	}

	if applyPlayoffTiebreakers {
		// Check scoring breakdowns to resolve playoff ties.

		// 2. Cumulative Fouls (Lower fouls is better, but here we count opponent fouls added to us, so higher is better for us)
		// logic: points given to us by opponent fouls.
		if status := comparePoints(
			redScoreSummary.NumOpponentMajorFouls, blueScoreSummary.NumOpponentMajorFouls,
		); status != TieMatch {
			return status
		}

		// 3. Auto Points
		if status := comparePoints(redScoreSummary.AutoPoints, blueScoreSummary.AutoPoints); status != TieMatch {
			return status
		}

		// 4. Tower Points (Climbing) - Replaces BargePoints from 2025
		if status := comparePoints(redScoreSummary.TotalTowerPoints, blueScoreSummary.TotalTowerPoints); status != TieMatch {
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
