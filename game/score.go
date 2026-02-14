// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotsBypassed      [3]bool
	ActiveFuel          int              // FUEL scored while hub was active
	InactiveFuel        int              // FUEL scored while hub was inactive (does NOT count for RPs)
	AutoFuel            int              // FUEL scored during autonomous
	AutoClimbStatuses   [3]EndgameStatus // Climb status at end of auto (Level 1 only)
	TeleopClimbStatuses [3]EndgameStatus // Climb status at end of teleop (Levels 1-3)
	Fouls               []Foul
	PlayoffDq           bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedRPThreshold = 100    // Minimum FUEL scored in HUB for ENERGIZED RP
var SuperchargedRPThreshold = 360 // Minimum FUEL scored in HUB for SUPERCHARGED RP
var TraversalRPThreshold = 50     // Minimum TOWER points for TRAVERSAL RP

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameLevel1
	EndgameLevel2
	EndgameLevel3
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	summary.AutoFuelPoints = score.AutoFuel * 1 // 1 point per auto FUEL

	// Auto climb points (Level 1 only = 15 points)
	for _, status := range score.AutoClimbStatuses {
		if status == EndgameLevel1 {
			summary.AutoClimbPoints += 15
		}
	}

	summary.AutoPoints = summary.AutoFuelPoints + summary.AutoClimbPoints

	// Calculate teleop FUEL points (only active FUEL counts for match points).
	summary.ActiveFuel = score.ActiveFuel
	summary.InactiveFuel = score.InactiveFuel
	summary.TotalFuel = score.AutoFuel + score.ActiveFuel + score.InactiveFuel
	summary.ActiveFuelPoints = score.ActiveFuel * 1 // 1 point per active FUEL

	// Calculate teleop climb points (TOWER points).
	for _, status := range score.TeleopClimbStatuses {
		switch status {
		case EndgameLevel1:
			summary.TeleopClimbPoints += 10
		case EndgameLevel2:
			summary.TeleopClimbPoints += 20
		case EndgameLevel3:
			summary.TeleopClimbPoints += 30
		default:
		}
	}

	summary.MatchPoints = summary.AutoFuelPoints + summary.AutoClimbPoints +
		summary.ActiveFuelPoints + summary.TeleopClimbPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				switch rule.RuleNumber {
				case "G206":
					// G206 violations handled below
				}
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	// Only auto and active FUEL count towards ENERGIZED and SUPERCHARGED RPs (inactive FUEL does NOT count).
	fuelForRankingPoints := score.AutoFuel + score.ActiveFuel

	// ENERGIZED RP: FUEL scored in HUB at or above threshold (auto + active only).
	if fuelForRankingPoints >= EnergizedRPThreshold {
		summary.EnergizedRankingPoint = true
	}

	// SUPERCHARGED RP: FUEL scored in HUB at or above threshold (auto + active only).
	if fuelForRankingPoints >= SuperchargedRPThreshold {
		summary.SuperchargedRankingPoint = true
	}

	// TRAVERSAL RP: TOWER points scored during match at or above threshold.
	totalTowerPoints := summary.AutoClimbPoints + summary.TeleopClimbPoints
	if totalTowerPoints >= TraversalRPThreshold {
		summary.TraversalRankingPoint = true
	}

	// Check for G206 violation (collusion to influence ranking points).
	for _, foul := range score.Fouls {
		if foul.Rule() != nil && foul.Rule().RuleNumber == "G206" {
			summary.EnergizedRankingPoint = false
			summary.SuperchargedRankingPoint = false
			summary.TraversalRankingPoint = false
			break
		}
	}

	// Add up the bonus ranking points.
	if summary.EnergizedRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.SuperchargedRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.TraversalRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.RobotsBypassed != other.RobotsBypassed ||
		score.ActiveFuel != other.ActiveFuel ||
		score.InactiveFuel != other.InactiveFuel ||
		score.AutoFuel != other.AutoFuel ||
		score.AutoClimbStatuses != other.AutoClimbStatuses ||
		score.TeleopClimbStatuses != other.TeleopClimbStatuses ||
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
