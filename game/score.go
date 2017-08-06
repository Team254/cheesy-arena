// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoMobility int
	AutoRotors   int
	AutoFuelLow  int
	AutoFuelHigh int
	Rotors       int
	FuelLow      int
	FuelHigh     int
	Takeoffs     int
	Fouls        []Foul
	ElimDq       bool
}

type ScoreSummary struct {
	AutoMobilityPoints  int
	AutoPoints          int
	RotorPoints         int
	TakeoffPoints       int
	PressurePoints      int
	BonusPoints         int
	FoulPoints          int
	Score               int
	PressureGoalReached bool
	RotorGoalReached    bool
}

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul, matchType string) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous score.
	summary.AutoMobilityPoints = 5 * score.AutoMobility
	summary.AutoPoints = summary.AutoMobilityPoints + 60*score.AutoRotors + score.AutoFuelHigh +
		score.AutoFuelLow/3

	// Calculate teleop score.
	summary.RotorPoints = 60*score.AutoRotors + 40*score.Rotors
	summary.TakeoffPoints = 50 * score.Takeoffs
	summary.PressurePoints = (9*score.AutoFuelHigh + 3*score.AutoFuelLow + 3*score.FuelHigh + score.FuelLow) / 9

	// Calculate bonuses.
	if summary.PressurePoints >= 40 {
		summary.PressureGoalReached = true
		if matchType == "elimination" {
			summary.BonusPoints += 20
		}
	}
	if score.AutoRotors+score.Rotors == 4 {
		summary.RotorGoalReached = true
		if matchType == "elimination" {
			summary.BonusPoints += 100
		}
	}

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	summary.Score = summary.AutoMobilityPoints + summary.RotorPoints + summary.TakeoffPoints + summary.PressurePoints +
		summary.BonusPoints + summary.FoulPoints

	return summary
}
