// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	ControlPanelDisabled = false

	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore.Fouls, true)
	assert.Equal(t, 10, redSummary.InitiationLinePoints)
	assert.Equal(t, 84, redSummary.AutoPowerCellPoints)
	assert.Equal(t, 94, redSummary.AutoPoints)
	assert.Equal(t, 38, redSummary.TeleopPowerCellPoints)
	assert.Equal(t, 122, redSummary.PowerCellPoints)
	assert.Equal(t, 15, redSummary.ControlPanelPoints)
	assert.Equal(t, 75, redSummary.EndgamePoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 222, redSummary.Score)
	assert.Equal(t, [3]int{0, 0, 13}, redSummary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, true, false}, redSummary.StagesActivated)
	assert.Equal(t, false, redSummary.ControlPanelRankingPoint)
	assert.Equal(t, true, redSummary.EndgameRankingPoint)

	blueSummary := blueScore.Summarize(redScore.Fouls, true)
	assert.Equal(t, 5, blueSummary.InitiationLinePoints)
	assert.Equal(t, 12, blueSummary.AutoPowerCellPoints)
	assert.Equal(t, 17, blueSummary.AutoPoints)
	assert.Equal(t, 122, blueSummary.TeleopPowerCellPoints)
	assert.Equal(t, 134, blueSummary.PowerCellPoints)
	assert.Equal(t, 35, blueSummary.ControlPanelPoints)
	assert.Equal(t, 50, blueSummary.EndgamePoints)
	assert.Equal(t, 33, blueSummary.FoulPoints)
	assert.Equal(t, 257, blueSummary.Score)
	assert.Equal(t, [3]int{0, 0, 0}, blueSummary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, true, true}, blueSummary.StagesActivated)
	assert.Equal(t, true, blueSummary.ControlPanelRankingPoint)
	assert.Equal(t, false, blueSummary.EndgameRankingPoint)

	// Test invalid foul.
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 18, blueScore.Summarize(redScore.Fouls, true).FoulPoints)

	// Test elimination disqualification.
	redScore.ElimDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls, true).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore.Fouls, true).Score)
	blueScore.ElimDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore.Fouls, true).Score)
}

func TestScoreSummaryControlPanelDisabled(t *testing.T) {
	ControlPanelDisabled = true
	CapacityWhenControlPanelDisabled = 24

	score := Score{}
	scoreSummary := score.Summarize([]Foul{}, true)
	assert.False(t, scoreSummary.ControlPanelRankingPoint)
	assert.Equal(t, 24, scoreSummary.StagePowerCellsRemaining[Stage1])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage2])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage3])

	score.AutoCellsBottom[Stage1] = 1
	score.AutoCellsOuter[Stage1] = 2
	score.AutoCellsInner[Stage1] = 3
	score.AutoCellsBottom[Stage2] = 100
	score.AutoCellsOuter[Stage2] = 100
	score.AutoCellsInner[Stage2] = 100
	score.TeleopCellsBottom[Stage1] = 4
	score.TeleopCellsOuter[Stage1] = 5
	score.TeleopCellsInner[Stage1] = 6
	score.TeleopCellsBottom[Stage2] = 100
	score.TeleopCellsOuter[Stage2] = 100
	score.TeleopCellsInner[Stage2] = 100
	score.TeleopCellsBottom[Stage3] = 100
	score.TeleopCellsOuter[Stage3] = 100
	score.TeleopCellsInner[Stage3] = 100
	score.TeleopCellsBottom[StageExtra] = 100
	score.TeleopCellsOuter[StageExtra] = 100
	score.TeleopCellsInner[StageExtra] = 100
	scoreSummary = score.Summarize([]Foul{}, true)
	assert.False(t, scoreSummary.ControlPanelRankingPoint)
	assert.Equal(t, 3, scoreSummary.StagePowerCellsRemaining[Stage1])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage2])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage3])

	score.TeleopCellsBottom[Stage1] = 10
	score.TeleopCellsOuter[Stage1] = 10
	score.TeleopCellsInner[Stage1] = 10
	scoreSummary = score.Summarize([]Foul{}, true)
	assert.True(t, scoreSummary.ControlPanelRankingPoint)
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage1])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage2])
	assert.Equal(t, 0, scoreSummary.StagePowerCellsRemaining[Stage3])
	assert.False(t, scoreSummary.StagesActivated[Stage1])
	assert.False(t, scoreSummary.StagesActivated[Stage2])
	assert.False(t, scoreSummary.StagesActivated[Stage3])
}

func TestScoreSummaryRungIsLevel(t *testing.T) {
	ControlPanelDisabled = false

	var score Score
	assert.Equal(t, 0, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 0, score.Summarize([]Foul{}, true).EndgamePoints)

	score.RungIsLevel = false
	score.EndgameStatuses = [3]EndgameStatus{EndgamePark, EndgamePark, EndgamePark}
	assert.Equal(t, 15, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 15, score.Summarize([]Foul{}, true).EndgamePoints)

	score.RungIsLevel = false
	score.EndgameStatuses = [3]EndgameStatus{EndgameHang, EndgamePark, EndgamePark}
	assert.Equal(t, 35, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 50, score.Summarize([]Foul{}, true).EndgamePoints)

	score.RungIsLevel = false
	score.EndgameStatuses = [3]EndgameStatus{EndgameHang, EndgamePark, EndgameHang}
	assert.Equal(t, 55, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 70, score.Summarize([]Foul{}, true).EndgamePoints)

	score.RungIsLevel = false
	score.EndgameStatuses = [3]EndgameStatus{EndgameHang, EndgameHang, EndgameHang}
	assert.Equal(t, 75, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 90, score.Summarize([]Foul{}, true).EndgamePoints)

	score.RungIsLevel = false
	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone}
	assert.Equal(t, 0, score.Summarize([]Foul{}, true).EndgamePoints)
	score.RungIsLevel = true
	assert.Equal(t, 0, score.Summarize([]Foul{}, true).EndgamePoints)
}

func TestScoreSummaryBoundaryConditions(t *testing.T) {
	ControlPanelDisabled = false

	// Test control panel boundary conditions.
	score := TestScore2()
	summary := score.Summarize(score.Fouls, true)
	assert.Equal(t, StageExtra, score.CellCountingStage(true))
	assert.Equal(t, [3]int{0, 0, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, true, true}, summary.StagesActivated)
	assert.Equal(t, true, summary.ControlPanelRankingPoint)
	assert.Equal(t, 224, summary.Score)

	score.TeleopCellsInner[0]--
	summary = score.Summarize(score.Fouls, true)
	assert.Equal(t, Stage1, score.CellCountingStage(true))
	assert.Equal(t, [3]int{1, 0, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{false, false, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 186, summary.Score)
	score.TeleopCellsInner[0]++

	summary = score.Summarize(score.Fouls, false)
	assert.Equal(t, Stage1, score.CellCountingStage(false))
	assert.Equal(t, [3]int{0, 0, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{false, false, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 189, summary.Score)

	score.TeleopCellsOuter[1] -= 7
	summary = score.Summarize(score.Fouls, true)
	assert.Equal(t, Stage2, score.CellCountingStage(true))
	assert.Equal(t, [3]int{0, 2, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, false, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 175, summary.Score)
	score.TeleopCellsOuter[1] += 7

	score.ControlPanelStatus = ControlPanelNone
	summary = score.Summarize(score.Fouls, true)
	assert.Equal(t, Stage2, score.CellCountingStage(true))
	assert.Equal(t, [3]int{0, 0, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, false, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 189, summary.Score)
	score.ControlPanelStatus = ControlPanelPosition

	score.TeleopCellsInner[2] -= 10
	summary = score.Summarize(score.Fouls, true)
	assert.Equal(t, Stage3, score.CellCountingStage(true))
	assert.Equal(t, [3]int{0, 0, 3}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, true, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 174, summary.Score)
	score.TeleopCellsInner[2] += 10

	score.ControlPanelStatus = ControlPanelRotation
	summary = score.Summarize(score.Fouls, true)
	assert.Equal(t, Stage3, score.CellCountingStage(true))
	assert.Equal(t, [3]int{0, 0, 0}, summary.StagePowerCellsRemaining)
	assert.Equal(t, [3]bool{true, true, false}, summary.StagesActivated)
	assert.Equal(t, false, summary.ControlPanelRankingPoint)
	assert.Equal(t, 204, summary.Score)

	// Test endgame boundary conditions.
	score = TestScore1()
	assert.Equal(t, true, score.Summarize(score.Fouls, true).EndgameRankingPoint)
	score.EndgameStatuses[0] = EndgameNone
	assert.Equal(t, false, score.Summarize(score.Fouls, true).EndgameRankingPoint)
	score.RungIsLevel = true
	assert.Equal(t, true, score.Summarize(score.Fouls, true).EndgameRankingPoint)
	score.EndgameStatuses[2] = EndgamePark
	assert.Equal(t, false, score.Summarize(score.Fouls, true).EndgameRankingPoint)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2 = TestScore1()
	score2.ExitedInitiationLine[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoCellsBottom[1] = 3
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoCellsOuter[0] = 7
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoCellsInner[1] = 8
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopCellsBottom[2] = 30
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopCellsOuter[1] = 31
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopCellsInner[0] = 32
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.ControlPanelStatus = ControlPanelNone
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameNone
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.RungIsLevel = !score2.RungIsLevel
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleId = 1
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
