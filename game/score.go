// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotStartLevels    [3]int
	SandstormBonuses    [3]bool
	CargoBaysPreMatch   [8]BayStatus
	CargoBays           [8]BayStatus
	RocketNearLeftBays  [3]BayStatus
	RocketNearRightBays [3]BayStatus
	RocketFarLeftBays   [3]BayStatus
	RocketFarRightBays  [3]BayStatus
	RobotEndLevels      [3]int
	Fouls               []Foul
	ElimDq              bool
}

type ScoreSummary struct {
	CargoPoints          int
	HatchPanelPoints     int
	HabClimbPoints       int
	SandstormBonusPoints int
	FoulPoints           int
	Score                int
	CompleteRocket       bool
	HabDocking           bool
}

// Represents the state of a cargo ship or rocket bay.
type BayStatus int

const (
	BayEmpty BayStatus = iota
	BayHatch
	BayHatchCargo
	BayCargo
)

var HabDockingThreshold = 15

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate sandstorm bonus points.
	for i, robotStartLevel := range score.RobotStartLevels {
		if score.SandstormBonuses[i] {
			if robotStartLevel == 1 {
				summary.SandstormBonusPoints += 3
			} else if robotStartLevel == 2 {
				summary.SandstormBonusPoints += 6
			}
		}
	}

	// Calculate cargo and hatch panel points.
	for i, bayStatus := range score.CargoBays {
		if bayStatus == BayHatchCargo {
			summary.CargoPoints += 3
			if score.CargoBaysPreMatch[i] != BayHatch {
				summary.HatchPanelPoints += 2
			}
		} else if bayStatus == BayHatch && score.CargoBaysPreMatch[i] != BayHatch {
			summary.HatchPanelPoints += 2
		}
	}
	summary.addRocketHalfPoints(score.RocketNearLeftBays)
	summary.addRocketHalfPoints(score.RocketNearRightBays)
	summary.addRocketHalfPoints(score.RocketFarLeftBays)
	summary.addRocketHalfPoints(score.RocketFarRightBays)

	// Calculate hab climb points.
	for _, level := range score.RobotEndLevels {
		switch level {
		case 1:
			summary.HabClimbPoints += 3
		case 2:
			summary.HabClimbPoints += 6
		case 3:
			summary.HabClimbPoints += 12
		}
	}

	// Calculate bonus ranking points.
	if score.isLevelComplete(0) && score.isLevelComplete(1) && score.isLevelComplete(2) {
		summary.CompleteRocket = true
	} else {
		// Check for the opponent fouls that automatically trigger the ranking point.
		for _, foul := range opponentFouls {
			if foul.IsRankingPoint {
				summary.CompleteRocket = true
				break
			}
		}
	}
	if summary.HabClimbPoints >= HabDockingThreshold {
		summary.HabDocking = true
	}

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	summary.Score = summary.CargoPoints + summary.HatchPanelPoints + summary.HabClimbPoints +
		summary.SandstormBonusPoints + summary.FoulPoints

	return summary
}

func (score *Score) Equals(other *Score) bool {
	if score.RobotStartLevels != other.RobotStartLevels ||
		score.SandstormBonuses != other.SandstormBonuses ||
		score.CargoBaysPreMatch != other.CargoBaysPreMatch ||
		score.CargoBays != other.CargoBays ||
		score.RocketNearLeftBays != other.RocketNearLeftBays ||
		score.RocketNearRightBays != other.RocketNearRightBays ||
		score.RocketFarLeftBays != other.RocketFarLeftBays ||
		score.RocketFarRightBays != other.RocketFarRightBays ||
		score.RobotEndLevels != other.RobotEndLevels ||
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

// Returns true if the score represents a valid pre-match state.
func (score *Score) IsValidPreMatch() bool {
	for i := 0; i < 3; i++ {
		// Ensure robot start level is set.
		if score.RobotStartLevels[i] == 0 || score.RobotStartLevels[i] > 3 {
			return false
		}

		// Ensure other robot fields and rocket bays are empty.
		if score.SandstormBonuses[i] || score.RobotEndLevels[i] != 0 || score.RocketNearLeftBays[i] != BayEmpty ||
			score.RocketNearRightBays[i] != BayEmpty || score.RocketFarLeftBays[i] != BayEmpty ||
			score.RocketFarRightBays[i] != BayEmpty {
			return false
		}
	}
	for i := 0; i < 8; i++ {
		if i == 3 || i == 4 {
			// Ensure cargo ship front bays are empty.
			if score.CargoBaysPreMatch[i] != BayEmpty {
				return false
			}
		} else {
			// Ensure cargo ship side bays have either a hatch or cargo but not both.
			if !(score.CargoBaysPreMatch[i] == BayHatch || score.CargoBaysPreMatch[i] == BayCargo) {
				return false
			}
		}
	}
	return score.CargoBays == score.CargoBaysPreMatch
}

// Calculates the cargo and hatch panel points for the given rocket half and adds them to the summary.
func (summary *ScoreSummary) addRocketHalfPoints(rocketHalf [3]BayStatus) {
	for _, bayStatus := range rocketHalf {
		if bayStatus == BayHatchCargo {
			summary.CargoPoints += 3
			summary.HatchPanelPoints += 2
		} else if bayStatus == BayHatch {
			summary.HatchPanelPoints += 2
		}
	}
}

// Returns true if the level is complete for at least one rocket.
func (score *Score) isLevelComplete(level int) bool {
	return score.RocketNearLeftBays[level] == BayHatchCargo && score.RocketNearRightBays[level] == BayHatchCargo ||
		score.RocketFarLeftBays[level] == BayHatchCargo && score.RocketFarRightBays[level] == BayHatchCargo
}
