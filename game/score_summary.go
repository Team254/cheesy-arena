// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	LeavePoints               int
	AutoPoints                int
	AmpPoints                 int
	SpeakerPoints             int
	StagePoints               int
	MatchPoints               int
	FoulPoints                int
	Score                     int
	CoopertitionCriteriaMet   bool
	CoopertitionBonus         bool
	NumNotes                  int
	NumNotesGoal              int
	MelodyBonusRankingPoint   bool
	EnsembleBonusRankingPoint bool
	BonusRankingPoints        int
	NumOpponentTechFouls      int

	// Fields only needed for TBA.
	ParkPoints      int
	OnStagePoints   int
	HarmonyPoints   int
	SpotlightPoints int
	TrapPoints      int
}

type MatchStatus int

const (
	MatchScheduled MatchStatus = iota
	MatchHidden
	RedWonMatch
	BlueWonMatch
	TieMatch
)

func (t MatchStatus) Get() MatchStatus {
	return t
}

// Determines the winner of the match given the score summaries for both alliances.
func DetermineMatchStatus(redScoreSummary, blueScoreSummary *ScoreSummary, applyPlayoffTiebreakers bool) MatchStatus {
	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status
	}

	if applyPlayoffTiebreakers {
		// Check scoring breakdowns to resolve playoff ties.
		if status := comparePoints(
			redScoreSummary.NumOpponentTechFouls, blueScoreSummary.NumOpponentTechFouls,
		); status != TieMatch {
			return status
		}
		if status := comparePoints(redScoreSummary.AutoPoints, blueScoreSummary.AutoPoints); status != TieMatch {
			return status
		}
		if status := comparePoints(redScoreSummary.StagePoints, blueScoreSummary.StagePoints); status != TieMatch {
			return status
		}
	}

	return TieMatch
}

// Helper method to compare the red and blue alliance point totals and return the appropriate MatchStatus.
func comparePoints(redPoints, bluePoints int) MatchStatus {
	if redPoints > bluePoints {
		return RedWonMatch
	}
	if redPoints < bluePoints {
		return BlueWonMatch
	}
	return TieMatch
}
