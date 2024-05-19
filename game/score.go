// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	LeaveStatuses      [3]bool
	AmpSpeaker         AmpSpeaker
	EndgameStatuses    [3]EndgameStatus
	MicrophoneStatuses [3]bool
	TrapStatuses       [3]bool
	Fouls              []Foul
	PlayoffDq          bool
}

// Game-specific constants that cannot be changed by the user.
const (
	bankedAmpNoteLimit          = 2
	ensembleBonusPointThreshold = 10
	ensembleBonusRobotThreshold = 2
)

// Game-specific settings that can be changed by the user.
var MelodyBonusThresholdWithoutCoop = 18
var MelodyBonusThresholdWithCoop = 15
var AmplificationNoteLimit = 4
var AmplificationDurationSec = 10

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameParked
	EndgameStageLeft
	EndgameCenterStage
	EndgameStageRight
)

// Represents a side of the Stage field element.
type StagePosition int

const (
	StageLeft StagePosition = iota
	CenterStage
	StageRight
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, status := range score.LeaveStatuses {
		if status {
			summary.LeavePoints += 2
		}
	}
	autoNotePoints := score.AmpSpeaker.AutoNotePoints()
	summary.AutoPoints = summary.LeavePoints + autoNotePoints

	// Calculate Amp and Speaker points.
	summary.AmpPoints = score.AmpSpeaker.AmpPoints()
	summary.SpeakerPoints = score.AmpSpeaker.SpeakerPoints()

	// Calculate endgame points.
	robotsByPosition := map[StagePosition]int{StageLeft: 0, CenterStage: 0, StageRight: 0}
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameParked:
			summary.ParkPoints += 1
		case EndgameStageLeft:
			summary.OnStagePoints += 3
			robotsByPosition[StageLeft]++
		case EndgameCenterStage:
			summary.OnStagePoints += 3
			robotsByPosition[CenterStage]++
		case EndgameStageRight:
			summary.OnStagePoints += 3
			robotsByPosition[StageRight]++
		default:
		}
	}
	totalOnstageRobots := 0
	for i := 0; i < 3; i++ {
		stagePosition := StagePosition(i)
		onstageRobots := robotsByPosition[stagePosition]
		totalOnstageRobots += onstageRobots

		// Handle Harmony (multiple robots climbing on the same chain).
		if onstageRobots > 1 {
			summary.HarmonyPoints += 2 * (onstageRobots - 1)
		}

		// Handle microphones.
		if score.MicrophoneStatuses[i] && onstageRobots > 0 {
			summary.SpotlightPoints += onstageRobots
		}

		// Handle traps.
		if score.TrapStatuses[i] {
			summary.TrapPoints += 5
		}
	}
	summary.StagePoints = summary.ParkPoints + summary.OnStagePoints + summary.HarmonyPoints + summary.SpotlightPoints +
		summary.TrapPoints

	summary.MatchPoints = summary.LeavePoints + summary.AmpPoints + summary.SpeakerPoints + summary.StagePoints

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
				summary.EnsembleBonusRankingPoint = true
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	summary.NumNotes = score.AmpSpeaker.TotalNotesScored()
	summary.NumNotesGoal = MelodyBonusThresholdWithoutCoop
	if MelodyBonusThresholdWithCoop > 0 {
		// A MelodyBonusThresholdWithCoop of 0 disables the coopertition bonus.
		summary.CoopertitionCriteriaMet = score.AmpSpeaker.CoopActivated
		summary.CoopertitionBonus = summary.CoopertitionCriteriaMet && opponentScore.AmpSpeaker.CoopActivated
		if summary.CoopertitionBonus {
			summary.NumNotesGoal = MelodyBonusThresholdWithCoop
		}
	}
	if summary.NumNotes >= summary.NumNotesGoal {
		summary.MelodyBonusRankingPoint = true
	}
	if summary.StagePoints >= ensembleBonusPointThreshold && totalOnstageRobots >= ensembleBonusRobotThreshold {
		summary.EnsembleBonusRankingPoint = true
	}

	if summary.MelodyBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.EnsembleBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.LeaveStatuses != other.LeaveStatuses ||
		score.AmpSpeaker != other.AmpSpeaker ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.MicrophoneStatuses != other.MicrophoneStatuses ||
		score.TrapStatuses != other.TrapStatuses ||
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
