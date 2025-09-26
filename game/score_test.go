// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 6, redSummary.LeavePoints)
	assert.Equal(t, 13, redSummary.AutoPoints)
	assert.Equal(t, 12, redSummary.NumCoral)
	assert.Equal(t, 34, redSummary.CoralPoints)
	assert.Equal(t, 9, redSummary.NumAlgae)
	assert.Equal(t, 40, redSummary.AlgaePoints)
	assert.Equal(t, 14, redSummary.BargePoints)
	assert.Equal(t, 94, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 94, redSummary.Score)
	assert.Equal(t, true, redSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redSummary.CoopertitionBonus)
	assert.Equal(t, 1, redSummary.NumCoralLevels)
	assert.Equal(t, 4, redSummary.NumCoralLevelsGoal)
	assert.Equal(t, true, redSummary.AutoBonusRankingPoint)
	assert.Equal(t, false, redSummary.CoralBonusRankingPoint)
	assert.Equal(t, false, redSummary.BargeBonusRankingPoint)
	assert.Equal(t, 1, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 3, blueSummary.LeavePoints)
	assert.Equal(t, 33, blueSummary.AutoPoints)
	assert.Equal(t, 26, blueSummary.NumCoral)
	assert.Equal(t, 83, blueSummary.CoralPoints)
	assert.Equal(t, 10, blueSummary.NumAlgae)
	assert.Equal(t, 42, blueSummary.AlgaePoints)
	assert.Equal(t, 24, blueSummary.BargePoints)
	assert.Equal(t, 152, blueSummary.MatchPoints)
	assert.Equal(t, 34, blueSummary.FoulPoints)
	assert.Equal(t, 186, blueSummary.Score)
	assert.Equal(t, false, blueSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueSummary.CoopertitionBonus)
	assert.Equal(t, 1, blueSummary.NumCoralLevels)
	assert.Equal(t, 4, blueSummary.NumCoralLevelsGoal)
	assert.Equal(t, false, blueSummary.AutoBonusRankingPoint)
	assert.Equal(t, false, blueSummary.CoralBonusRankingPoint)
	assert.Equal(t, true, blueSummary.BargeBonusRankingPoint)
	assert.Equal(t, 1, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 34, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore).Score)
	blueScore.PlayoffDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore).Score)
}

func TestScoreAutoBonusRankingPoint(t *testing.T) {
	redScore := TestScore1()
	redScore.RobotsBypassed = [3]bool{false, false, false}
	redScore.LeaveStatuses = [3]bool{false, false, false}
	blueScore := TestScore2()

	// No robots left; no bonus is awarded.
	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, false, redSummary.AutoBonusRankingPoint)

	// All robots left; the bonus is awarded.
	redScore.LeaveStatuses = [3]bool{true, true, true}
	redSummary = redScore.Summarize(blueScore)
	assert.Equal(t, true, redSummary.AutoBonusRankingPoint)

	// One robot failed to leave; no bonus is awarded.
	for i := 0; i < 3; i++ {
		redScore.LeaveStatuses = [3]bool{true, true, true}
		redScore.LeaveStatuses[i] = false
		redSummary = redScore.Summarize(blueScore)
		assert.Equal(t, false, redSummary.AutoBonusRankingPoint)
	}

	// One bypassed robot failed to leave; the bonus is awarded.
	for i := 0; i < 3; i++ {
		redScore.RobotsBypassed = [3]bool{false, false, false}
		redScore.RobotsBypassed[i] = true
		redScore.LeaveStatuses = [3]bool{true, true, true}
		redScore.LeaveStatuses[i] = false
		redSummary = redScore.Summarize(blueScore)
		assert.Equal(t, true, redSummary.AutoBonusRankingPoint)
	}

	// Only one robot left but the other two were bypassed; the bonus is awarded.
	redScore.RobotsBypassed = [3]bool{false, true, true}
	redScore.LeaveStatuses = [3]bool{true, false, false}
	redSummary = redScore.Summarize(blueScore)
	assert.Equal(t, true, redSummary.AutoBonusRankingPoint)

	// No coral is scored; the bonus is not awarded.
	redScore.Reef = Reef{}
	redSummary = redScore.Summarize(blueScore)
	assert.Equal(t, false, redSummary.AutoBonusRankingPoint)
}

func TestScoreCoralBonusRankingPoint(t *testing.T) {
	// Save the original threshold value and restore it after the test.
	originalThreshold := CoralBonusPerLevelThreshold
	defer func() {
		CoralBonusPerLevelThreshold = originalThreshold
		CoralBonusCoopEnabled = true
	}()
	CoralBonusPerLevelThreshold = 3

	redScore := TestScore1()
	blueScore := TestScore2()

	redScoreSummary := redScore.Summarize(blueScore)
	blueScoreSummary := blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 2, redScoreSummary.NumCoralLevels)
	assert.Equal(t, 4, redScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, false, redScoreSummary.CoralBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 4, blueScoreSummary.NumCoralLevels)
	assert.Equal(t, 4, blueScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, true, blueScoreSummary.CoralBonusRankingPoint)

	// Activate coopertition bonus for the blue alliance.
	blueScore.ProcessorAlgae = 2
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 2, redScoreSummary.NumCoralLevels)
	assert.Equal(t, 3, redScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, false, redScoreSummary.CoralBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 4, blueScoreSummary.NumCoralLevels)
	assert.Equal(t, 3, blueScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, true, blueScoreSummary.CoralBonusRankingPoint)

	// Satisfy the Coral bonus requirement for the red alliance.
	redScore.Reef.Branches[0] = [12]bool{true, true, true, true}
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 3, redScoreSummary.NumCoralLevels)
	assert.Equal(t, 3, redScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, true, redScoreSummary.CoralBonusRankingPoint)

	// Disable the coopertition bonus.
	CoralBonusCoopEnabled = false
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 3, redScoreSummary.NumCoralLevels)
	assert.Equal(t, 4, redScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, false, redScoreSummary.CoralBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionCriteriaMet)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 4, blueScoreSummary.NumCoralLevels)
	assert.Equal(t, 4, blueScoreSummary.NumCoralLevelsGoal)
	assert.Equal(t, true, blueScoreSummary.CoralBonusRankingPoint)

	// Check that G206 disqualifies the alliance from the Coral bonus.
	blueScore.Fouls = []Foul{{FoulId: 1, RuleId: 1}}
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 0, redScoreSummary.FoulPoints)
	assert.Equal(t, false, blueScoreSummary.CoralBonusRankingPoint)
	assert.Equal(t, 0, blueScoreSummary.BonusRankingPoints)
}

func TestScoreBargeBonusRankingPoint(t *testing.T) {
	// Save the original threshold value and restore it after the test.
	originalThreshold := BargeBonusPointThreshold
	defer func() {
		BargeBonusPointThreshold = originalThreshold
	}()

	testCases := []struct {
		endgameStatuses      [3]EndgameStatus
		fouls                []Foul
		threshold            int
		expectedBonusAwarded bool
	}{
		// 0. No endgame points.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone},
			fouls:                []Foul{},
			threshold:            14,
			expectedBonusAwarded: false,
		},

		// 1. All robots parked.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameParked, EndgameParked, EndgameParked},
			fouls:                []Foul{},
			threshold:            14,
			expectedBonusAwarded: false,
		},

		// 2. Meeting the minimum threshold.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameParked, EndgameNone, EndgameDeepCage},
			fouls:                []Foul{},
			threshold:            14,
			expectedBonusAwarded: true,
		},

		// 3. Same endgame statuses not meeting a higher threshold.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameParked, EndgameNone, EndgameDeepCage},
			fouls:                []Foul{},
			threshold:            16,
			expectedBonusAwarded: false,
		},

		// 4. Meeting the new minimum threshold with a different combination.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameDeepCage, EndgameParked, EndgameParked},
			fouls:                []Foul{},
			threshold:            16,
			expectedBonusAwarded: true,
		},

		// 5. One of each endgame status with higher threshold.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameShallowCage, EndgameDeepCage, EndgameParked},
			fouls:                []Foul{},
			threshold:            21,
			expectedBonusAwarded: false,
		},

		// 6. All deep climbs.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameDeepCage, EndgameDeepCage, EndgameDeepCage},
			fouls:                []Foul{},
			threshold:            36,
			expectedBonusAwarded: true,
		},

		// 7. G206 foul disqualifies the alliance from the Barge bonus.
		{
			endgameStatuses:      [3]EndgameStatus{EndgameDeepCage, EndgameDeepCage, EndgameDeepCage},
			fouls:                []Foul{{RuleId: 1}},
			threshold:            14,
			expectedBonusAwarded: false,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				BargeBonusPointThreshold = tc.threshold
				score := Score{EndgameStatuses: tc.endgameStatuses, Fouls: tc.fouls}
				summary := score.Summarize(&Score{})
				assert.Equal(t, tc.expectedBonusAwarded, summary.BargeBonusRankingPoint)
			},
		)
	}
}

func TestScoreBargeBonusRankingPointIncludingAlgae(t *testing.T) {
	// Save the original setting and restore it after the test.
	originalIncludeAlgae := IncludeAlgaeInBargeBonus
	defer func() {
		IncludeAlgaeInBargeBonus = originalIncludeAlgae
	}()

	IncludeAlgaeInBargeBonus = false
	BargeBonusPointThreshold = 36

	score := Score{
		EndgameStatuses: [3]EndgameStatus{EndgameDeepCage, EndgameDeepCage, EndgameParked},
		BargeAlgae:      1,
		ProcessorAlgae:  1,
	}
	summary := score.Summarize(&Score{})
	assert.Equal(t, false, summary.BargeBonusRankingPoint)

	IncludeAlgaeInBargeBonus = true
	summary = score.Summarize(&Score{})
	assert.Equal(t, true, summary.BargeBonusRankingPoint)
}

func TestScoreAutoRankingPointFromFouls(t *testing.T) {
	testCases := []struct {
		ownFouls           []Foul
		opponentFouls      []Foul
		expectedCoralBonus bool
		expectedBargeBonus bool
	}{
		// 0. No fouls - no automatic ranking points.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{},
			expectedCoralBonus: false,
			expectedBargeBonus: false,
		},

		// 1. G410 foul automatically awards coral bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 14}},
			expectedCoralBonus: true,
			expectedBargeBonus: false,
		},

		// 2. G418 foul automatically awards barge bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 21}},
			expectedCoralBonus: false,
			expectedBargeBonus: true,
		},

		// 3. G428 foul automatically awards barge bonus.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 33}},
			expectedCoralBonus: false,
			expectedBargeBonus: true,
		},

		// 4. All fouls together still automatically award both bonuses.
		{
			ownFouls:           []Foul{},
			opponentFouls:      []Foul{{RuleId: 14}, {RuleId: 21}, {RuleId: 33}},
			expectedCoralBonus: true,
			expectedBargeBonus: true,
		},

		// 5. G206 makes the alliance ineligible for both bonuses.
		{
			ownFouls:           []Foul{{RuleId: 1}},
			opponentFouls:      []Foul{{RuleId: 14}, {RuleId: 21}, {RuleId: 33}},
			expectedCoralBonus: false,
			expectedBargeBonus: false,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				redScore := Score{Fouls: tc.ownFouls}
				blueScore := Score{Fouls: tc.opponentFouls}
				redSummary := redScore.Summarize(&blueScore)
				assert.Equal(t, tc.expectedCoralBonus, redSummary.CoralBonusRankingPoint)
				assert.Equal(t, tc.expectedBargeBonus, redSummary.BargeBonusRankingPoint)

				// Count expected total bonus ranking points.
				expectedBonusRankingPoints := 0
				if tc.expectedCoralBonus {
					expectedBonusRankingPoints++
				}
				if tc.expectedBargeBonus {
					expectedBonusRankingPoints++
				}
				assert.Equal(t, expectedBonusRankingPoints, redSummary.BonusRankingPoints)
			},
		)
	}
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
	score2.RobotsBypassed[0] = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.LeaveStatuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Reef.TroughFar = 5
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.BargeAlgae = 9
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.ProcessorAlgae = 3
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameParked
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsMajor = false
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
