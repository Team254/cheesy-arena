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

var QuintetThreshold = 5
var CargoBonusRankingPointThresholdWithoutQuintet = 20
var CargoBonusRankingPointThresholdWithQuintet = 18
var HangarBonusRankingPointThreshold = 16
var DoubleBonusRankingPointThreshold = 0

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
	summary.CargoGoal = CargoBonusRankingPointThresholdWithoutQuintet
	// A QuintetThreshold of 0 disables the Quintet.
	if QuintetThreshold > 0 && summary.AutoCargoCount >= QuintetThreshold {
		summary.CargoGoal = CargoBonusRankingPointThresholdWithQuintet
		summary.QuintetAchieved = true
	}
	if summary.CargoCount >= summary.CargoGoal {
		summary.CargoBonusRankingPoint = true
	}
	summary.HangarBonusRankingPoint = summary.HangarPoints >= HangarBonusRankingPointThreshold

	// The "double bonus" ranking point is an offseason-only addition which grants an additional RP if either the total
	// cargo count or the hangar points is over the certain threshold. A threshold of 0 disables this RP.
	if DoubleBonusRankingPointThreshold > 0 {
		summary.DoubleBonusRankingPoint = summary.CargoCount >= DoubleBonusRankingPointThreshold ||
			summary.HangarPoints >= DoubleBonusRankingPointThreshold
	}

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
