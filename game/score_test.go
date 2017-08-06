// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore.Fouls, "qualification")
	assert.Equal(t, 0, redSummary.AutoMobilityPoints)
	assert.Equal(t, 80, redSummary.AutoPoints)
	assert.Equal(t, 100, redSummary.RotorPoints)
	assert.Equal(t, 50, redSummary.TakeoffPoints)
	assert.Equal(t, 40, redSummary.PressurePoints)
	assert.Equal(t, 0, redSummary.BonusPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 190, redSummary.Score)
	assert.Equal(t, true, redSummary.PressureGoalReached)
	assert.Equal(t, false, redSummary.RotorGoalReached)

	blueSummary := blueScore.Summarize(redScore.Fouls, "qualification")
	assert.Equal(t, 10, blueSummary.AutoMobilityPoints)
	assert.Equal(t, 133, blueSummary.AutoPoints)
	assert.Equal(t, 200, blueSummary.RotorPoints)
	assert.Equal(t, 150, blueSummary.TakeoffPoints)
	assert.Equal(t, 18, blueSummary.PressurePoints)
	assert.Equal(t, 0, blueSummary.BonusPoints)
	assert.Equal(t, 55, blueSummary.FoulPoints)
	assert.Equal(t, 433, blueSummary.Score)
	assert.Equal(t, false, blueSummary.PressureGoalReached)
	assert.Equal(t, true, blueSummary.RotorGoalReached)

	// Test pressure boundary conditions.
	redScore.AutoFuelHigh = 19
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls, "qualification").PressureGoalReached)
	redScore.FuelLow = 18
	assert.Equal(t, true, redScore.Summarize(blueScore.Fouls, "qualification").PressureGoalReached)
	redScore.AutoFuelLow = 1
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls, "qualification").PressureGoalReached)
	redScore.FuelHigh = 56
	assert.Equal(t, true, redScore.Summarize(blueScore.Fouls, "qualification").PressureGoalReached)

	// Test rotor boundary conditions.
	blueScore.AutoRotors = 1
	assert.Equal(t, false, blueScore.Summarize(blueScore.Fouls, "qualification").RotorGoalReached)
	blueScore.Rotors = 3
	assert.Equal(t, true, blueScore.Summarize(blueScore.Fouls, "qualification").RotorGoalReached)

	// Test elimination bonus.
	redSummary = redScore.Summarize(blueScore.Fouls, "elimination")
	blueSummary = blueScore.Summarize(redScore.Fouls, "elimination")
	assert.Equal(t, 20, redSummary.BonusPoints)
	assert.Equal(t, 210, redSummary.Score)
	assert.Equal(t, 100, blueSummary.BonusPoints)
	assert.Equal(t, 513, blueSummary.Score)
	redScore.Rotors = 3
	redSummary = redScore.Summarize(blueScore.Fouls, "elimination")
	assert.Equal(t, 120, redSummary.BonusPoints)
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls, "qualification").BonusPoints)
	assert.Equal(t, 0, blueScore.Summarize(blueScore.Fouls, "qualification").BonusPoints)

	// Test elimination disqualification.
	redScore.ElimDq = true
	blueScore.ElimDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls, "elimination").Score)
	assert.Equal(t, 0, blueScore.Summarize(redScore.Fouls, "elimination").Score)
}
