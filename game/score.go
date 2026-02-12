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
}

// Game-specific settings that can be changed via the settings.
var EnergizedFuelThreshold = 42    // 獲得 Energized RP 需要的球數
var SuperchargedFuelThreshold = 55 // 獲得 Supercharged RP 需要的球數
var TraversalPointThreshold = 60   // 獲得 Traversal RP 需要的爬升總分

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone   EndgameStatus = iota
	EndgameLevel2               // Low Rung (20 pts)
	EndgameLevel3               // Mid Rung (30 pts)
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
			summary.AutoTowerPoints += 10
		}
	}

	summary.AutoPoints = summary.AutoFuelPoints + summary.AutoTowerPoints

	// --- 2. Teleop & Endgame Points ---
	// Fuel: 1 pt per Fuel in Active Hub
	summary.TeleopFuelPoints = score.TeleopFuelCount * 1

	// Endgame Tower: Level 2 (20pts), Level 3 (30pts)
	for _, status := range score.EndgameStatuses {
		switch status {
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
			// 處理特殊規則 G420 (Endgame Protection)
			// 規則: 對手犯規，我方獲得 Level 3 Climb 分數 (30分)
			if rule.RuleNumber == "G420" {
				summary.EndgameTowerPoints += 30
			}
		}
	}

	// 彙總 Match Points
	summary.TotalFuelPoints = summary.AutoFuelPoints + summary.TeleopFuelPoints
	summary.TotalTowerPoints = summary.AutoTowerPoints + summary.EndgameTowerPoints
	summary.MatchPoints = summary.TotalFuelPoints + summary.TotalTowerPoints

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// --- 4. Ranking Points (RP) Calculation ---

	// A. Energized RP (基於 Fuel 數量)
	totalFuel := score.AutoFuelCount + score.TeleopFuelCount
	if totalFuel >= EnergizedFuelThreshold {
		summary.EnergizedRankingPoint = true
	}

	// B. Supercharged RP (基於更高的 Fuel 數量)
	if totalFuel >= SuperchargedFuelThreshold {
		summary.SuperchargedRankingPoint = true
	}

	// C. Traversal RP (基於 Tower 總分)
	// 包含從 G420 獲得的額外爬升分數
	if summary.TotalTowerPoints >= TraversalPointThreshold {
		summary.TraversalRankingPoint = true
	}

	// Check for G206 violation (Collusion for RP).
	// 如果自己犯規 G206，取消所有 Bonus RP
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
