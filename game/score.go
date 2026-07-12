// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoTowerStatuses    [3]TowerStatus
	Hub                  Hub
	EndgameTowerStatuses [3]TowerStatus
	Fouls                []Foul
	PlayoffDq            bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedBonusThreshold = 100
var SuperchargedBonusThreshold = 360
var TraversalBonusThreshold = 50

// Represents the state of a robot on the Tower, at the end of auto or teleop.
type TowerStatus int

const (
	TowerNone TowerStatus = iota
	// Park indicates the robot is parked (partial scoring).
	TowerPark
	// Complete indicates the robot completed the action (higher scoring in auto).
	TowerComplete
)

// Backwards-compatible aliases for older code/tests that referenced level names.
const (
	TowerLevel1 = TowerPark
	TowerLevel2 = TowerComplete
	TowerLevel3 = TowerComplete
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)
	summary.PlayoffDq = score.PlayoffDq

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	summary.AutoFuelPoints = score.Hub.GetShiftCount(ShiftAuto, true)
	summary.NumFuel += summary.AutoFuelPoints
	numAutoRobots := 0
	for _, status := range score.AutoTowerStatuses {
		if status == TowerComplete {
			// Complete during auto: +12 points (cap two robots).
			summary.AutoTowerPoints += 12
			numAutoRobots++
		} else if status == TowerPark {
			// Park during auto: +5 points.
			summary.AutoTowerPoints += 5
			numAutoRobots++
		}
		if numAutoRobots == 2 {
			break
		}
	}

	// Calculate teleoperated period points.
	summary.TeleopFuelPoints = score.Hub.GetTeleopActiveFuelCount()
	summary.NumFuelPostMatch = score.Hub.GetShiftCount(ShiftPostMatch, true)
	summary.NumFuel += summary.TeleopFuelPoints
	// Endgame (post-match) parking/complete scoring. Both Park and Complete count as +5.
	for _, status := range score.EndgameTowerStatuses {
		switch status {
		case TowerPark, TowerComplete:
			summary.TeleopTowerPoints += 5
		default:
		}
	}

	summary.MatchPoints = summary.AutoFuelPoints + summary.AutoTowerPoints +
		summary.TeleopFuelPoints + summary.TeleopTowerPoints
	summary.PostMatchPoints = summary.TeleopTowerPoints + summary.NumFuelPostMatch

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Fuel bonus ranking points.
	summary.NumFuelGoal = EnergizedBonusThreshold
	if summary.NumFuel >= EnergizedBonusThreshold {
		summary.EnergizedBonusRankingPoint = true
		summary.NumFuelGoal = SuperchargedBonusThreshold
	}
	summary.SuperchargedBonusRankingPoint = summary.NumFuel >= SuperchargedBonusThreshold

	// Tower bonus ranking point.
	summary.TraversalBonusRankingPoint = summary.AutoTowerPoints+summary.TeleopTowerPoints >= TraversalBonusThreshold

	// Check for G206 violation.
	for _, foul := range score.Fouls {
		if foul.Rule() != nil && foul.Rule().RuleNumber == "G206" {
			summary.EnergizedBonusRankingPoint = false
			summary.SuperchargedBonusRankingPoint = false
			summary.TraversalBonusRankingPoint = false
			break
		}
	}

	// Add up the bonus ranking points.
	if summary.EnergizedBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.SuperchargedBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.TraversalBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.AutoTowerStatuses != other.AutoTowerStatuses ||
		score.Hub != other.Hub ||
		score.EndgameTowerStatuses != other.EndgameTowerStatuses ||
		score.PlayoffDq != other.PlayoffDq ||
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
