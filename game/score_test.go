// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	MelodyBonusThresholdWithoutCoop = 18
	MelodyBonusThresholdWithCoop = 15
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 4, redSummary.LeavePoints)
	assert.Equal(t, 36, redSummary.AutoPoints)
	assert.Equal(t, 6, redSummary.AmpPoints)
	assert.Equal(t, 57, redSummary.SpeakerPoints)
	assert.Equal(t, 14, redSummary.StagePoints)
	assert.Equal(t, 81, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 81, redSummary.Score)
	assert.Equal(t, true, redSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redSummary.CoopertitionBonus)
	assert.Equal(t, 17, redSummary.NumNotes)
	assert.Equal(t, 18, redSummary.NumNotesGoal)
	assert.Equal(t, false, redSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, redSummary.EnsembleBonusRankingPoint)
	assert.Equal(t, 0, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentTechFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 2, blueSummary.LeavePoints)
	assert.Equal(t, 42, blueSummary.AutoPoints)
	assert.Equal(t, 51, blueSummary.AmpPoints)
	assert.Equal(t, 161, blueSummary.SpeakerPoints)
	assert.Equal(t, 13, blueSummary.StagePoints)
	assert.Equal(t, 227, blueSummary.MatchPoints) // 187
	assert.Equal(t, 29, blueSummary.FoulPoints)
	assert.Equal(t, 256, blueSummary.Score)
	assert.Equal(t, false, blueSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueSummary.CoopertitionBonus)
	assert.Equal(t, 85, blueSummary.NumNotes)
	assert.Equal(t, 18, blueSummary.NumNotesGoal)
	assert.Equal(t, true, blueSummary.MelodyBonusRankingPoint)
	assert.Equal(t, true, blueSummary.EnsembleBonusRankingPoint)
	assert.Equal(t, 2, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentTechFouls)

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

func TestScoreMelodyBonusRankingPoint(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redScoreSummary := redScore.Summarize(blueScore)
	blueScoreSummary := blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 17, redScoreSummary.NumNotes)
	assert.Equal(t, 18, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 85, blueScoreSummary.NumNotes)
	assert.Equal(t, 18, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Reduce blue notes to 18 and verify that the bonus is still awarded.
	blueScore.AmpSpeaker.TeleopAmpNotes = 2
	blueScore.AmpSpeaker.TeleopAmplifiedSpeakerNotes = 5
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 17, redScoreSummary.NumNotes)
	assert.Equal(t, 18, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 18, blueScoreSummary.NumNotes)
	assert.Equal(t, 18, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Increase non-coopertition threshold above the blue note count.
	MelodyBonusThresholdWithoutCoop = 19
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 17, redScoreSummary.NumNotes)
	assert.Equal(t, 19, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 18, blueScoreSummary.NumNotes)
	assert.Equal(t, 19, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, false, blueScoreSummary.MelodyBonusRankingPoint)

	// Reduce red notes to the non-coopertition threshold.
	MelodyBonusThresholdWithCoop = 16
	redScore.AmpSpeaker.TeleopAmpNotes = 3
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 16, redScoreSummary.NumNotes)
	assert.Equal(t, 19, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 18, blueScoreSummary.NumNotes)
	assert.Equal(t, 19, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, false, blueScoreSummary.MelodyBonusRankingPoint)

	// Make blue fulfill the coopertition bonus requirement.
	blueScore.AmpSpeaker.CoopActivated = true
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 16, redScoreSummary.NumNotes)
	assert.Equal(t, 16, redScoreSummary.NumNotesGoal)
	assert.Equal(t, true, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 18, blueScoreSummary.NumNotes)
	assert.Equal(t, 16, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Disable the coopertition bonus.
	MelodyBonusThresholdWithCoop = 0
	blueScore.AmpSpeaker.AutoSpeakerNotes = 9
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 16, redScoreSummary.NumNotes)
	assert.Equal(t, 19, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 19, blueScoreSummary.NumNotes)
	assert.Equal(t, 19, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)
}

func TestScoreEnsembleBonusRankingPoint(t *testing.T) {
	var score Score

	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone}
	score.MicrophoneStatuses = [3]bool{false, false, false}
	score.TrapStatuses = [3]bool{false, false, false}
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameStageLeft, EndgameCenterStage, EndgameStageRight}
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	// Try various combinations of Harmony.
	score.EndgameStatuses = [3]EndgameStatus{EndgameStageLeft, EndgameCenterStage, EndgameStageLeft}
	assert.Equal(t, 11, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameCenterStage, EndgameCenterStage, EndgameStageLeft}
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameCenterStage, EndgameCenterStage, EndgameStageLeft}
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameStageRight, EndgameStageRight, EndgameCenterStage}
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameStageRight, EndgameStageRight, EndgameStageRight}
	assert.Equal(t, 13, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	// Try various combinations with microphones.
	score.EndgameStatuses = [3]EndgameStatus{EndgameStageLeft, EndgameCenterStage, EndgameStageRight}
	score.MicrophoneStatuses = [3]bool{true, false, false}
	assert.Equal(t, 10, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.MicrophoneStatuses = [3]bool{true, true, true}
	assert.Equal(t, 12, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameStageRight, EndgameStageRight}
	score.MicrophoneStatuses = [3]bool{false, false, true}
	assert.Equal(t, 10, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameParked, EndgameStageRight, EndgameCenterStage}
	score.MicrophoneStatuses = [3]bool{false, true, false}
	assert.Equal(t, 8, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	// Try various combinations with traps.
	score.EndgameStatuses = [3]EndgameStatus{EndgameStageLeft, EndgameCenterStage, EndgameParked}
	score.MicrophoneStatuses = [3]bool{false, false, false}
	score.TrapStatuses = [3]bool{false, false, true}
	assert.Equal(t, 12, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameParked, EndgameCenterStage, EndgameParked}
	score.TrapStatuses = [3]bool{true, true, true}
	assert.Equal(t, 20, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
	score.EndgameStatuses = [3]EndgameStatus{EndgameParked, EndgameParked, EndgameParked}
	assert.Equal(t, 18, score.Summarize(&Score{}).StagePoints)
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
}

func TestScoreFreeEnsembleBonusRankingPointFromFoul(t *testing.T) {
	var score1, score2 Score
	foul := Foul{IsTechnical: true, RuleId: 29}

	assert.Equal(t, true, foul.Rule().IsTechnical)
	assert.Equal(t, true, foul.Rule().IsRankingPoint)
	score2.Fouls = []Foul{foul}

	summary := score1.Summarize(&score2)
	assert.Equal(t, 5, summary.Score)
	assert.Equal(t, true, summary.EnsembleBonusRankingPoint)
	assert.Equal(t, 1, summary.BonusRankingPoints)

	summary = score2.Summarize(&score1)
	assert.Equal(t, 0, summary.Score)
	assert.Equal(t, false, summary.EnsembleBonusRankingPoint)
	assert.Equal(t, 0, summary.BonusRankingPoints)
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
	score2.LeaveStatuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AmpSpeaker.AutoAmpNotes = 5
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameParked
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.MicrophoneStatuses[0] = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TrapStatuses[0] = false
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
