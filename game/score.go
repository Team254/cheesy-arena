// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	MobilityStatuses          [3]bool
	Grid                      Grid
	AutoDockStatuses          [3]bool
	AutoChargeStationLevel    bool
	EndgameStatuses           [3]EndgameStatus
	EndgameChargeStationLevel bool
	Fouls                     []Foul
	PlayoffDq                 bool
}

var SustainabilityBonusLinkThresholdWithoutCoop = 7
var SustainabilityBonusLinkThresholdWithCoop = 6
var ActivationBonusPointThreshold = 26

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameParked
	EndgameDocked
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, mobility := range score.MobilityStatuses {
		if mobility {
			summary.MobilityPoints += 3
		}
	}
	autoGridPoints := score.Grid.AutoGamePiecePoints()
	autoChargeStationPoints := 0
	for i := 0; i < 3; i++ {
		if score.AutoDockStatuses[i] {
			autoChargeStationPoints += 8
			if score.AutoChargeStationLevel {
				autoChargeStationPoints += 4
			}
			break
		}
	}
	summary.AutoPoints = summary.MobilityPoints + autoGridPoints + autoChargeStationPoints

	// Calculate teleoperated period points.
	teleopGridPoints := score.Grid.TeleopGamePiecePoints() + score.Grid.LinkPoints() + score.Grid.SuperchargedPoints()
	teleopChargeStationPoints := 0
	for i := 0; i < 3; i++ {
		switch score.EndgameStatuses[i] {
		case EndgameParked:
			summary.ParkPoints += 2
		case EndgameDocked:
			teleopChargeStationPoints += 6
			if score.EndgameChargeStationLevel {
				teleopChargeStationPoints += 4
			}
		}
	}

	summary.GridPoints = autoGridPoints + teleopGridPoints
	summary.ChargeStationPoints = autoChargeStationPoints + teleopChargeStationPoints
	summary.EndgamePoints = teleopChargeStationPoints + summary.ParkPoints
	summary.MatchPoints = summary.MobilityPoints + summary.GridPoints + summary.ChargeStationPoints + summary.ParkPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of tech fouls since it is used to break ties in playoffs.
		if foul.IsTechnical {
			summary.NumOpponentTechFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				summary.SustainabilityBonusRankingPoint = true
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	summary.CoopertitionBonus = score.Grid.IsCoopertitionThresholdAchieved() &&
		opponentScore.Grid.IsCoopertitionThresholdAchieved()
	summary.NumLinks = len(score.Grid.Links())
	summary.NumLinksGoal = SustainabilityBonusLinkThresholdWithoutCoop
	// A SustainabilityBonusLinkThresholdWithCoop of 0 disables the coopertition bonus.
	if SustainabilityBonusLinkThresholdWithCoop > 0 && summary.CoopertitionBonus {
		summary.NumLinksGoal = SustainabilityBonusLinkThresholdWithCoop
	}
	if summary.NumLinks >= summary.NumLinksGoal {
		summary.SustainabilityBonusRankingPoint = true
	}
	summary.ActivationBonusRankingPoint = summary.ChargeStationPoints >= ActivationBonusPointThreshold

	if summary.SustainabilityBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.ActivationBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.MobilityStatuses != other.MobilityStatuses ||
		score.Grid != other.Grid ||
		score.AutoDockStatuses != other.AutoDockStatuses ||
		score.AutoChargeStationLevel != other.AutoChargeStationLevel ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.EndgameChargeStationLevel != other.EndgameChargeStationLevel ||
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
