// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoRuns                 int
	AutoSwitchOwnershipSec   float64
	AutoScaleOwnershipSec    float64
	AutoEndSwitchOwnership   bool
	TeleopScaleOwnershipSec  float64
	TeleopScaleBoostSec      float64
	TeleopSwitchOwnershipSec float64
	TeleopSwitchBoostSec     float64
	ForceCubes               int
	ForceCubesPlayed         int
	LevitateCubes            int
	LevitatePlayed           bool
	BoostCubes               int
	BoostCubesPlayed         int
	Climbs                   int
	Parks                    int
	Fouls                    []Foul
	ElimDq                   bool
}

type ScoreSummary struct {
	AutoRunPoints         int
	AutoOwnershipPoints   int
	AutoPoints            int
	TeleopOwnershipPoints int
	OwnershipPoints       int
	VaultPoints           int
	ParkClimbPoints       int
	FoulPoints            int
	Score                 int
	AutoQuest             bool
	FaceTheBoss           bool
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
	summary.AutoOwnershipPoints = int(2 * (score.AutoScaleOwnershipSec + score.AutoSwitchOwnershipSec))
	summary.AutoPoints = summary.AutoRunPoints + summary.AutoOwnershipPoints

	// Calculate teleop score.
	summary.TeleopOwnershipPoints = int(score.TeleopScaleOwnershipSec + score.TeleopScaleBoostSec +
		score.TeleopSwitchOwnershipSec + score.TeleopSwitchBoostSec)
	summary.OwnershipPoints = summary.AutoOwnershipPoints + summary.TeleopOwnershipPoints
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
		score.AutoScaleOwnershipSec != other.AutoScaleOwnershipSec ||
		score.AutoSwitchOwnershipSec != other.AutoSwitchOwnershipSec ||
		score.TeleopScaleOwnershipSec != other.TeleopScaleOwnershipSec ||
		score.TeleopScaleBoostSec != other.TeleopScaleBoostSec ||
		score.TeleopSwitchOwnershipSec != other.TeleopSwitchOwnershipSec ||
		score.TeleopSwitchBoostSec != other.TeleopSwitchBoostSec ||
		score.ForceCubes != other.ForceCubes ||
		score.ForceCubesPlayed != other.ForceCubesPlayed || score.LevitateCubes != other.LevitateCubes ||
		score.LevitatePlayed != other.LevitatePlayed || score.BoostCubes != other.BoostCubes ||
		score.BoostCubesPlayed != other.BoostCubesPlayed || score.Parks != other.Parks ||
		score.Climbs != other.Climbs || score.ElimDq != other.ElimDq || len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
