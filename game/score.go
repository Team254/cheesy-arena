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
	ForceCubes             int
	ForcePlayed            bool
	LevitateCubes          int
	LevitatePlayed         bool
	BoostCubes             int
	BoostPlayed            bool
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
	autoRuns := score.AutoRuns
	if autoRuns > 3 {
		autoRuns = 3
	}
	summary.AutoRunPoints = 5 * autoRuns
	summary.AutoPoints = summary.AutoRunPoints + score.AutoOwnershipPoints

	// Calculate teleop score.
	summary.OwnershipPoints = score.AutoOwnershipPoints + score.TeleopOwnershipPoints
	forceCubes := score.ForceCubes
	if forceCubes > 3 {
		forceCubes = 3
	}
	levitateCubes := score.LevitateCubes
	if levitateCubes > 3 {
		levitateCubes = 3
	}
	boostCubes := score.BoostCubes
	if boostCubes > 3 {
		boostCubes = 3
	}
	summary.VaultPoints = 5 * (forceCubes + levitateCubes + boostCubes)
	climbs := score.Climbs
	if climbs > 3 {
		climbs = 3
	}
	if score.LevitatePlayed && score.Climbs < 3 {
		climbs++
	}
	parks := score.Parks
	if parks+climbs > 3 {
		parks = 3 - climbs
	}
	summary.ParkClimbPoints = 5*parks + 30*climbs

	// Calculate bonuses.
	if autoRuns == 3 && score.AutoEndSwitchOwnership {
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
		score.TeleopOwnershipPoints != other.TeleopOwnershipPoints || score.ForceCubes != other.ForceCubes ||
		score.ForcePlayed != other.ForcePlayed || score.LevitateCubes != other.LevitateCubes ||
		score.LevitatePlayed != other.LevitatePlayed || score.BoostCubes != other.BoostCubes ||
		score.BoostPlayed != other.BoostPlayed || score.Parks != other.Parks || score.Climbs != other.Climbs ||
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
