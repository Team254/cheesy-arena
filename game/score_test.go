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

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 6, redSummary.MobilityPoints)
	assert.Equal(t, 36, redSummary.AutoPoints)
	assert.Equal(t, 52, redSummary.GridPoints)
	assert.Equal(t, 18, redSummary.ChargeStationPoints)
	assert.Equal(t, 2, redSummary.ParkPoints)
	assert.Equal(t, 12, redSummary.EndgamePoints)
	assert.Equal(t, 78, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 78, redSummary.Score)
	assert.Equal(t, false, redSummary.CoopertitionBonus)
	assert.Equal(t, 0, redSummary.NumLinks)
	assert.Equal(t, 6, redSummary.NumLinksGoal)
	assert.Equal(t, false, redSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, false, redSummary.ActivationBonusRankingPoint)
	assert.Equal(t, 0, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentTechFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 3, blueSummary.MobilityPoints)
	assert.Equal(t, 43, blueSummary.AutoPoints)
	assert.Equal(t, 154, blueSummary.GridPoints)
	assert.Equal(t, 30, blueSummary.ChargeStationPoints)
	assert.Equal(t, 0, blueSummary.ParkPoints)
	assert.Equal(t, 18, blueSummary.EndgamePoints)
	assert.Equal(t, 187, blueSummary.MatchPoints)
	assert.Equal(t, 29, blueSummary.FoulPoints)
	assert.Equal(t, 216, blueSummary.Score)
	assert.Equal(t, false, blueSummary.CoopertitionBonus)
	assert.Equal(t, 9, blueSummary.NumLinks)
	assert.Equal(t, 6, blueSummary.NumLinksGoal)
	assert.Equal(t, true, blueSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, true, blueSummary.ActivationBonusRankingPoint)
	assert.Equal(t, 2, blueSummary.BonusRankingPoints)
	assert.Equal(t, 2, blueSummary.NumOpponentTechFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 29, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore).Score)
	blueScore.PlayoffDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore).Score)
}

func TestScoreSustainabilityBonusRankingPoint(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redScoreSummary := redScore.Summarize(blueScore)
	blueScoreSummary := blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumLinks)
	assert.Equal(t, 6, redScoreSummary.NumLinksGoal)
	assert.Equal(t, false, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 9, blueScoreSummary.NumLinks)
	assert.Equal(t, 6, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, true, blueScoreSummary.SustainabilityBonusRankingPoint)

	// Reduce blue links to 8 and verify that the bonus is still awarded.
	blueScore.Grid.Nodes[rowBottom][0] = Empty
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumLinks)
	assert.Equal(t, 6, redScoreSummary.NumLinksGoal)
	assert.Equal(t, false, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 8, blueScoreSummary.NumLinks)
	assert.Equal(t, 6, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, true, blueScoreSummary.SustainabilityBonusRankingPoint)

	// Increase non-coopertition threshold to 9.
	SustainabilityBonusLinkThresholdWithoutCoop = 9
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumLinks)
	assert.Equal(t, 9, redScoreSummary.NumLinksGoal)
	assert.Equal(t, false, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 8, blueScoreSummary.NumLinks)
	assert.Equal(t, 9, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, false, blueScoreSummary.SustainabilityBonusRankingPoint)

	// Reduce blue links to 6 and verify that the sustainability bonus is not awarded.
	blueScore.Grid.Nodes[rowMiddle][0] = Empty
	blueScore.Grid.Nodes[rowTop][0] = Empty
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumLinks)
	assert.Equal(t, 9, redScoreSummary.NumLinksGoal)
	assert.Equal(t, false, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumLinks)
	assert.Equal(t, 9, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, false, blueScoreSummary.SustainabilityBonusRankingPoint)

	// Make red fulfill the coopertition bonus requirement.
	redScore.Grid.Nodes[rowBottom][4] = Cone
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumLinks)
	assert.Equal(t, 5, redScoreSummary.NumLinksGoal)
	assert.Equal(t, false, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumLinks)
	assert.Equal(t, 5, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, true, blueScoreSummary.SustainabilityBonusRankingPoint)

	// Reduce coopertition threshold to 1 and make red fulfill the sustainability bonus requirement.
	SustainabilityBonusLinkThresholdWithCoop = 1
	redScore.Grid.Nodes[rowBottom][5] = Cube
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 1, redScoreSummary.NumLinks)
	assert.Equal(t, 1, redScoreSummary.NumLinksGoal)
	assert.Equal(t, true, redScoreSummary.SustainabilityBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumLinks)
	assert.Equal(t, 1, blueScoreSummary.NumLinksGoal)
	assert.Equal(t, true, blueScoreSummary.SustainabilityBonusRankingPoint)
}

func TestScoreActivationBonusRankingPoint(t *testing.T) {
	var score Score

	score.AutoDockStatuses = [3]bool{true, false, false}
	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone}
	assert.Equal(t, false, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	score.AutoDockStatuses = [3]bool{true, false, false}
	score.EndgameStatuses = [3]EndgameStatus{EndgameDocked, EndgameNone, EndgameDocked}
	assert.Equal(t, false, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	score.AutoChargeStationLevel = false
	score.EndgameChargeStationLevel = true
	assert.Equal(t, true, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	score.AutoChargeStationLevel = true
	score.EndgameChargeStationLevel = false
	assert.Equal(t, false, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	ActivationBonusPointThreshold = 30
	score.AutoChargeStationLevel = true
	score.EndgameChargeStationLevel = true
	assert.Equal(t, true, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	score.AutoChargeStationLevel = false
	score.EndgameChargeStationLevel = true
	assert.Equal(t, false, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	ActivationBonusPointThreshold = 42
	score.AutoDockStatuses = [3]bool{true, true, true}
	score.EndgameStatuses = [3]EndgameStatus{EndgameDocked, EndgameDocked, EndgameDocked}
	score.AutoChargeStationLevel = true
	score.EndgameChargeStationLevel = true
	assert.Equal(t, true, score.Summarize(&Score{}).ActivationBonusRankingPoint)

	ActivationBonusPointThreshold = 43
	assert.Equal(t, false, score.Summarize(&Score{}).ActivationBonusRankingPoint)
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
	score2.MobilityStatuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Grid.Nodes[rowTop][8] = ConeThenCube
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoDockStatuses[2] = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoChargeStationLevel = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameParked
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameChargeStationLevel = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsTechnical = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleId = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.PlayoffDq = !score2.PlayoffDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
