// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoRuns               int
	AutoOwnershipPoints    int
	AutoEndSwitchOwnership bool
	TeleopOwnershipPoints  int
	VaultCubes             int
	Levitate               bool
	Climbs                 int
	Parks                  int
	Fouls                  []Foul
	ElimDq                 bool
}

type ScoreSummary struct {
	AutoRunPoints   int
	AutoPoints      int
	OwnershipPoints int
	VaultPoints     int
	ParkClimbPoints int
	FoulPoints      int
	Score           int
	AutoQuest       bool
	FaceTheBoss     bool
}

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous score.
	summary.AutoRunPoints = 5 * score.AutoRuns
	summary.AutoPoints = summary.AutoRunPoints + score.AutoOwnershipPoints

	// Calculate teleop score.
	summary.OwnershipPoints = score.AutoOwnershipPoints + score.TeleopOwnershipPoints
	summary.VaultPoints = 5 * score.VaultCubes
	climbs := score.Climbs
	if climbs > 3 {
		climbs = 3
	}
	if score.Levitate && score.Climbs < 3 {
		climbs++
	}
	parks := score.Parks
	if parks+climbs > 3 {
		parks = 3 - climbs
	}
	summary.ParkClimbPoints = 5*parks + 30*climbs

	// Calculate bonuses.
	if score.AutoRuns == 3 && score.AutoEndSwitchOwnership {
		summary.AutoQuest = true
	}
	if climbs == 3 {
		summary.FaceTheBoss = true
	}

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	summary.Score = summary.AutoRunPoints + summary.OwnershipPoints + summary.VaultPoints + summary.ParkClimbPoints +
		summary.FoulPoints

	return summary
}

func (score *Score) Equals(other *Score) bool {
	if score.AutoRuns != other.AutoRuns || score.AutoEndSwitchOwnership != other.AutoEndSwitchOwnership ||
		score.AutoOwnershipPoints != other.AutoOwnershipPoints ||
		score.TeleopOwnershipPoints != other.TeleopOwnershipPoints || score.VaultCubes != other.VaultCubes ||
		score.Levitate != other.Levitate || score.Parks != other.Parks || score.Climbs != other.Climbs ||
		score.ElimDq != other.ElimDq || len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
