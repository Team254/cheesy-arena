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
	assert.Equal(t, 4, redSummary.TaxiPoints)
	assert.Equal(t, 7, redSummary.AutoCargoCount)
	assert.Equal(t, 26, redSummary.AutoCargoPoints)
	assert.Equal(t, 17, redSummary.CargoCount)
	assert.Equal(t, 44, redSummary.CargoPoints)
	assert.Equal(t, 19, redSummary.HangarPoints)
	assert.Equal(t, 67, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 67, redSummary.Score)
	assert.Equal(t, true, redSummary.QuintetAchieved)
	assert.Equal(t, false, redSummary.CargoBonusRankingPoint)
	assert.Equal(t, true, redSummary.HangarBonusRankingPoint)

	blueSummary := blueScore.Summarize(redScore.Fouls)
	assert.Equal(t, 2, blueSummary.TaxiPoints)
	assert.Equal(t, 4, blueSummary.AutoCargoCount)
	assert.Equal(t, 14, blueSummary.AutoCargoPoints)
	assert.Equal(t, 25, blueSummary.CargoCount)
	assert.Equal(t, 45, blueSummary.CargoPoints)
	assert.Equal(t, 14, blueSummary.HangarPoints)
	assert.Equal(t, 61, blueSummary.MatchPoints)
	assert.Equal(t, 20, blueSummary.FoulPoints)
	assert.Equal(t, 81, blueSummary.Score)
	assert.Equal(t, false, blueSummary.QuintetAchieved)
	assert.Equal(t, true, blueSummary.CargoBonusRankingPoint)
	assert.Equal(t, false, blueSummary.HangarBonusRankingPoint)

	// Test invalid foul.
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 12, blueScore.Summarize(redScore.Fouls).FoulPoints)

	// Test elimination disqualification.
	redScore.ElimDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore.Fouls).Score)
	blueScore.ElimDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore.Fouls).Score)
}

func TestScoreCargoBonusRankingPoint(t *testing.T) {
	var score Score

	score.AutoCargoLower[0] = 2
	summary := score.Summarize([]Foul{})
	assert.Equal(t, 3, summary.AutoCargoRemaining)
	assert.Equal(t, 18, summary.TeleopCargoRemaining)
	assert.Equal(t, false, summary.QuintetAchieved)
	assert.Equal(t, false, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 17
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 0, summary.AutoCargoRemaining)
	assert.Equal(t, 1, summary.TeleopCargoRemaining)
	assert.Equal(t, true, summary.QuintetAchieved)
	assert.Equal(t, false, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 18
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 0, summary.AutoCargoRemaining)
	assert.Equal(t, 0, summary.TeleopCargoRemaining)
	assert.Equal(t, true, summary.QuintetAchieved)
	assert.Equal(t, true, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 5
	score.TeleopCargoLower[0] = 12
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 0, summary.AutoCargoRemaining)
	assert.Equal(t, 1, summary.TeleopCargoRemaining)
	assert.Equal(t, true, summary.QuintetAchieved)
	assert.Equal(t, false, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 5
	score.TeleopCargoLower[0] = 13
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 0, summary.AutoCargoRemaining)
	assert.Equal(t, 0, summary.TeleopCargoRemaining)
	assert.Equal(t, true, summary.QuintetAchieved)
	assert.Equal(t, true, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 3
	score.TeleopCargoLower[0] = 6
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 2, summary.AutoCargoRemaining)
	assert.Equal(t, 11, summary.TeleopCargoRemaining)
	assert.Equal(t, false, summary.QuintetAchieved)
	assert.Equal(t, false, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 4
	score.TeleopCargoLower[0] = 15
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 1, summary.AutoCargoRemaining)
	assert.Equal(t, 1, summary.TeleopCargoRemaining)
	assert.Equal(t, false, summary.QuintetAchieved)
	assert.Equal(t, false, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 4
	score.TeleopCargoLower[0] = 16
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 1, summary.AutoCargoRemaining)
	assert.Equal(t, 0, summary.TeleopCargoRemaining)
	assert.Equal(t, false, summary.QuintetAchieved)
	assert.Equal(t, true, summary.CargoBonusRankingPoint)

	score.AutoCargoLower[0] = 0
	score.TeleopCargoLower[0] = 20
	summary = score.Summarize([]Foul{})
	assert.Equal(t, 5, summary.AutoCargoRemaining)
	assert.Equal(t, 0, summary.TeleopCargoRemaining)
	assert.Equal(t, false, summary.QuintetAchieved)
	assert.Equal(t, true, summary.CargoBonusRankingPoint)
}

func TestScoreHangarBonusRankingPoint(t *testing.T) {
	var score Score

	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameLow, EndgameLow, EndgameLow}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameLow, EndgameLow, EndgameMid}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameMid, EndgameLow, EndgameMid}
	assert.Equal(t, true, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameMid, EndgameLow, EndgameNone}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameHigh, EndgameLow, EndgameNone}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameHigh, EndgameLow, EndgameLow}
	assert.Equal(t, true, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameHigh, EndgameMid, EndgameNone}
	assert.Equal(t, true, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameHigh, EndgameNone, EndgameNone}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameTraversal}
	assert.Equal(t, false, score.Summarize([]Foul{}).HangarBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameLow, EndgameTraversal}
	assert.Equal(t, true, score.Summarize([]Foul{}).HangarBonusRankingPoint)
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
	score2.TaxiStatuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoCargoLower[1] = 3
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoCargoUpper[0] = 7
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopCargoLower[2] = 30
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopCargoUpper[1] = 31
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[0] = EndgameNone
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
