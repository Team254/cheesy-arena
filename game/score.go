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
	TowerLevel1
	TowerLevel2
	TowerLevel3
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	summary.AutoFuelPoints = score.Hub.GetAutoFuelCount()
	summary.NumFuel += summary.AutoFuelPoints
	for _, status := range score.AutoTowerStatuses {
		if status == TowerLevel1 || status == TowerLevel2 || status == TowerLevel3 {
			summary.AutoTowerPoints += 15
			break
		}
	}

	// Calculate teleoperated period points.
	summary.TeleopFuelPoints = score.Hub.GetTeleopActiveFuelCount()
	summary.NumFuel += summary.TeleopFuelPoints
	for _, status := range score.EndgameTowerStatuses {
		switch status {
		case TowerLevel1:
			summary.TeleopTowerPoints += 10
		case TowerLevel2:
			summary.TeleopTowerPoints += 20
		case TowerLevel3:
			summary.TeleopTowerPoints += 30
		default:
		}
	}

	summary.MatchPoints = summary.AutoFuelPoints + summary.AutoTowerPoints +
		summary.TeleopFuelPoints + summary.TeleopTowerPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}

		// TODO: Update for 2026.
		// rule := foul.Rule()
		// if rule != nil {
		// 	// Check for the opponent fouls that automatically trigger a ranking point.
		// 	if rule.IsRankingPoint {
		// 		switch rule.RuleNumber {
		// 		case "G410":
		// 			summary.CoralBonusRankingPoint = true
		// 		case "G418":
		// 			summary.BargeBonusRankingPoint = true
		// 		case "G428":
		// 			summary.BargeBonusRankingPoint = true
		// 		}
		// 	}
		// }
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
