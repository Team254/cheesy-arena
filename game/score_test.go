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

	redSummary := redScore.Summarize(blueScore.Fouls)
	assert.Equal(t, 30, redSummary.CargoPoints)
	assert.Equal(t, 20, redSummary.HatchPanelPoints)
	assert.Equal(t, 12, redSummary.HabClimbPoints)
	assert.Equal(t, 9, redSummary.SandstormBonusPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 71, redSummary.Score)
	assert.Equal(t, true, redSummary.CompleteRocket)
	assert.Equal(t, false, redSummary.HabDocking)

	blueSummary := blueScore.Summarize(redScore.Fouls)
	assert.Equal(t, 12, blueSummary.CargoPoints)
	assert.Equal(t, 0, blueSummary.HatchPanelPoints)
	assert.Equal(t, 15, blueSummary.HabClimbPoints)
	assert.Equal(t, 6, blueSummary.SandstormBonusPoints)
	assert.Equal(t, 55, blueSummary.FoulPoints)
	assert.Equal(t, 88, blueSummary.Score)
	assert.Equal(t, false, blueSummary.CompleteRocket)
	assert.Equal(t, true, blueSummary.HabDocking)

	// Test rocket completion boundary conditions.
	assert.Equal(t, true, redScore.Summarize(blueScore.Fouls).CompleteRocket)
	redScore.RocketFarLeftBays[1] = BayHatch
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls).CompleteRocket)
	redScore.RocketNearLeftBays[1] = BayHatchCargo
	redScore.RocketNearRightBays[1] = BayHatchCargo
	assert.Equal(t, true, redScore.Summarize(blueScore.Fouls).CompleteRocket)
	redScore.RocketNearLeftBays[2] = BayHatch
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls).CompleteRocket)
	redScore.Fouls[1].IsRankingPoint = true
	assert.Equal(t, true, redScore.Summarize(redScore.Fouls).CompleteRocket)

	// Test hab docking boundary conditions.
	assert.Equal(t, true, blueScore.Summarize(redScore.Fouls).HabDocking)
	HabDockingThreshold = 24
	assert.Equal(t, false, blueScore.Summarize(redScore.Fouls).HabDocking)
	blueScore.RobotEndLevels[0] = 3
	assert.Equal(t, true, blueScore.Summarize(redScore.Fouls).HabDocking)

	// Test elimination disqualification.
	redScore.ElimDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore.Fouls).Score)
	blueScore.ElimDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore.Fouls).Score)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2.RobotStartLevels[2] = 3
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.SandstormBonuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.CargoBaysPreMatch[7] = BayCargo
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.CargoBays[5] = BayHatchCargo
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RocketNearLeftBays[0] = BayEmpty
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RocketNearRightBays[1] = BayHatchCargo
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RocketFarLeftBays[2] = BayCargo
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RocketFarRightBays[0] = BayHatch
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RobotEndLevels[1] = 2
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleNumber = "G1000"
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsTechnical = !score2.Fouls[0].IsTechnical
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TimeInMatchSec += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.ElimDq = !score2.ElimDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
