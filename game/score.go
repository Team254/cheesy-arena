// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotsBypassed  [3]bool
	AutoTowerLevel1 [3]bool
	AutoFuelCount   int
	TeleopFuelCount int
	EndgameStatuses [3]EndgameStatus
	Fouls           []Foul
	PlayoffDq       bool
	HubActive       bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedFuelThreshold = 100    // Number of balls required to obtain an Energized RP
var SuperchargedFuelThreshold = 360 // Number of balls required to obtain a Supercharged RP
var TraversalPointThreshold = 50    // Number of tower points required to obtain a Traversal RP

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone   EndgameStatus = iota // No Rung (0 pts)
	EndgameLevel1                      // 1 Rung (10 pts)
	EndgameLevel2                      // 2 Rung (20 pts)
	EndgameLevel3                      // 3 Rung (30 pts)
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// --- 1. Autonomous Period Points ---
	// Fuel: 1 pt per Fuel in Active Hub
	summary.AutoFuelPoints = score.AutoFuelCount * 1

	// Tower Level 1: 10 pts per Robot
	for _, reachedL1 := range score.AutoTowerLevel1 {
		if reachedL1 {
			summary.AutoTowerPoints += 15
		}
	}

	summary.AutoPoints = summary.AutoFuelPoints + summary.AutoTowerPoints

	// --- 2. Teleop & Endgame Points ---
	// Fuel: 1 pt per Fuel in Active Hub
	summary.TeleopFuelPoints = score.TeleopFuelCount * 1

	// Endgame Tower: Level 2 (20pts), Level 3 (30pts)
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameLevel1:
			summary.EndgameTowerPoints += 10
		case EndgameLevel2:
			summary.EndgameTowerPoints += 20
		case EndgameLevel3:
			summary.EndgameTowerPoints += 30
		default:
		}
	}

	// --- 3. Penalty Points & Special Rules ---
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Handle special rule G420 (Endgame Protection)
			// Rule: If the opponent commits G420, our team gets Level 3 Climb points (30 points)
			if rule.RuleNumber == "G420" {
				summary.EndgameTowerPoints += 30
			}
		}
	}

	// Summarize Match Points
	summary.TotalFuelPoints = summary.AutoFuelPoints + summary.TeleopFuelPoints
	summary.TotalTowerPoints = summary.AutoTowerPoints + summary.EndgameTowerPoints
	summary.MatchPoints = summary.TotalFuelPoints + summary.TotalTowerPoints

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// --- 4. Ranking Points (RP) Calculation ---

	// A. Energized RP (based on Fuel count)
	totalFuel := score.AutoFuelCount + score.TeleopFuelCount
	if totalFuel >= EnergizedFuelThreshold {
		summary.EnergizedRankingPoint = true
	} else {
		summary.EnergizedRankingPoint = false
	}

	// B. Supercharged RP (based on higher Fuel count)
	if totalFuel >= SuperchargedFuelThreshold {
		summary.SuperchargedRankingPoint = true
	} else {
		summary.SuperchargedRankingPoint = false
	}

	// C. Traversal RP (based on Tower points)
	// Includes additional climb points from G420
	if summary.TotalTowerPoints >= TraversalPointThreshold {
		summary.TraversalRankingPoint = true
	}

	// Check for G206 violation (Collusion for RP).
	// If our team commits G206, all Bonus RP are revoked.
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
		score.HubActive != other.HubActive ||
		score.AutoTowerLevel1 != other.AutoTowerLevel1 ||
		score.AutoFuelCount != other.AutoFuelCount ||
		score.TeleopFuelCount != other.TeleopFuelCount ||
		score.EndgameStatuses != other.EndgameStatuses ||
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
