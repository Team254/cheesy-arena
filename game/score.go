// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	TaxiStatuses     [3]bool
	AutoCargoLower   [4]int
	AutoCargoUpper   [4]int
	TeleopCargoLower [4]int
	TeleopCargoUpper [4]int
	EndgameStatuses  [3]EndgameStatus
	Fouls            []Foul
	ElimDq           bool
}

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
}

var QuintetThreshold = 5
var CargoBonusRankingPointThresholdWithoutQuintet = 20
var CargoBonusRankingPointThresholdWithQuintet = 18
var HangarBonusRankingPointThreshold = 16

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameLow
	EndgameMid
	EndgameHigh
	EndgameTraversal
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, taxied := range score.TaxiStatuses {
		if taxied {
			summary.TaxiPoints += 2
		}
	}
	for i := 0; i < 4; i++ {
		summary.AutoCargoCount += score.AutoCargoLower[i] + score.AutoCargoUpper[i]
		summary.AutoCargoPoints += 2 * score.AutoCargoLower[i]
		summary.AutoCargoPoints += 4 * score.AutoCargoUpper[i]
	}

	// Calculate teleoperated period cargo points.
	summary.CargoCount = summary.AutoCargoCount
	summary.CargoPoints = summary.AutoCargoPoints
	for i := 0; i < 4; i++ {
		summary.CargoCount += score.TeleopCargoLower[i] + score.TeleopCargoUpper[i]
		summary.CargoPoints += 1 * score.TeleopCargoLower[i]
		summary.CargoPoints += 2 * score.TeleopCargoUpper[i]
	}

	// Calculate endgame points.
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameLow:
			summary.HangarPoints += 4
		case EndgameMid:
			summary.HangarPoints += 6
		case EndgameHigh:
			summary.HangarPoints += 10
		case EndgameTraversal:
			summary.HangarPoints += 15
		}
	}

	// Calculate bonus ranking points.
	var cargoBonusRankingPointThreshold int
	if summary.AutoCargoCount >= QuintetThreshold {
		cargoBonusRankingPointThreshold = CargoBonusRankingPointThresholdWithQuintet
		summary.AutoCargoRemaining = 0
		summary.QuintetAchieved = true
	} else {
		cargoBonusRankingPointThreshold = CargoBonusRankingPointThresholdWithoutQuintet
		summary.AutoCargoRemaining = QuintetThreshold - summary.AutoCargoCount
	}
	if summary.CargoCount >= cargoBonusRankingPointThreshold {
		summary.TeleopCargoRemaining = 0
		summary.CargoBonusRankingPoint = true
	} else {
		summary.TeleopCargoRemaining = cargoBonusRankingPointThreshold - summary.CargoCount
	}
	summary.HangarBonusRankingPoint = summary.HangarPoints >= HangarBonusRankingPointThreshold

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	// Check for the opponent fouls that automatically trigger a ranking point.
	// Note: There are no such fouls in the 2022 game; leaving this comment for future years.

	summary.MatchPoints = summary.TaxiPoints + summary.CargoPoints + summary.HangarPoints
	summary.Score = summary.MatchPoints + summary.FoulPoints

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.TaxiStatuses != other.TaxiStatuses ||
		score.AutoCargoLower != other.AutoCargoLower ||
		score.AutoCargoUpper != other.AutoCargoUpper ||
		score.TeleopCargoLower != other.TeleopCargoLower ||
		score.TeleopCargoUpper != other.TeleopCargoUpper ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.ElimDq != other.ElimDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
