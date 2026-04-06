// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 18, redSummary.AutoFuelPoints)
	assert.Equal(t, 15, redSummary.AutoTowerPoints)
	assert.Equal(t, 70, redSummary.TeleopFuelPoints)
	assert.Equal(t, 30, redSummary.TeleopTowerPoints)
	assert.Equal(t, 88, redSummary.NumFuel)
	assert.Equal(t, 100, redSummary.NumFuelGoal)
	assert.Equal(t, 133, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 133, redSummary.Score)
	assert.Equal(t, false, redSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, false, redSummary.SuperchargedBonusRankingPoint)
	assert.Equal(t, false, redSummary.TraversalBonusRankingPoint)
	assert.Equal(t, 0, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentMajorFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 35, blueSummary.AutoFuelPoints)
	assert.Equal(t, 15, blueSummary.AutoTowerPoints)
	assert.Equal(t, 79, blueSummary.TeleopFuelPoints)
	assert.Equal(t, 60, blueSummary.TeleopTowerPoints)
	assert.Equal(t, 114, blueSummary.NumFuel)
	assert.Equal(t, 360, blueSummary.NumFuelGoal)
	assert.Equal(t, 189, blueSummary.MatchPoints)
	assert.Equal(t, 85, blueSummary.FoulPoints)
	assert.Equal(t, 274, blueSummary.Score)
	assert.Equal(t, true, blueSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, false, blueSummary.SuperchargedBonusRankingPoint)
	assert.Equal(t, true, blueSummary.TraversalBonusRankingPoint)
	assert.Equal(t, 2, blueSummary.BonusRankingPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 85, blueScore.Summarize(redScore).FoulPoints)

	// Test that G206 does not add foul points.
	redScore.Fouls = append(redScore.Fouls, Foul{FoulId: 8, RuleId: 1})
	blueSummary = blueScore.Summarize(redScore)
	assert.Equal(t, 85, blueSummary.FoulPoints)
	assert.Equal(t, 5, blueSummary.NumOpponentMajorFouls)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore).Score)
	blueScore.PlayoffDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore).Score)
}

func TestScoreEnergizedBonusRankingPoint(t *testing.T) {
	originalThreshold := EnergizedBonusThreshold
	originalSuperchargedThreshold := SuperchargedBonusThreshold
	defer func() {
		EnergizedBonusThreshold = originalThreshold
		SuperchargedBonusThreshold = originalSuperchargedThreshold
	}()
	EnergizedBonusThreshold = 91
	SuperchargedBonusThreshold = 351

	redScore := TestScore1()
	redSummary := redScore.Summarize(&Score{})
	assert.Equal(t, false, redSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, 91, redSummary.NumFuelGoal)

	redScore.Hub.ShiftCounts[ShiftEndgame] += 2
	redSummary = redScore.Summarize(&Score{})
	assert.Equal(t, false, redSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, 91, redSummary.NumFuelGoal)

	// Meeting the threshold awards the ranking point and advances the displayed goal.
	redScore.Hub.ShiftCounts[ShiftEndgame] += 1
	redSummary = redScore.Summarize(&Score{})
	assert.Equal(t, true, redSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, 351, redSummary.NumFuelGoal)

	// Fuel scored while the Hub is inactive for that alliance does not count.
	redScore = TestScore1()
	redScore.Hub.ShiftCounts[Shift2] += 100
	redSummary = redScore.Summarize(&Score{})
	assert.Equal(t, false, redSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, 88, redSummary.NumFuel)
}

func TestScoreSuperchargedBonusRankingPoint(t *testing.T) {
	originalEnergizedThreshold := EnergizedBonusThreshold
	originalSuperchargedThreshold := SuperchargedBonusThreshold
	defer func() {
		EnergizedBonusThreshold = originalEnergizedThreshold
		SuperchargedBonusThreshold = originalSuperchargedThreshold
	}()
	EnergizedBonusThreshold = 113
	SuperchargedBonusThreshold = 361

	blueScore := TestScore2()
	blueScore.Hub.ShiftCounts[ShiftEndgame] += 245
	blueSummary := blueScore.Summarize(&Score{})
	assert.Equal(t, true, blueSummary.EnergizedBonusRankingPoint)
	assert.Equal(t, false, blueSummary.SuperchargedBonusRankingPoint)
	assert.Equal(t, 359, blueSummary.NumFuel)
	assert.Equal(t, 361, blueSummary.NumFuelGoal)

	blueScore.Hub.ShiftCounts[ShiftEndgame] += 2
	blueSummary = blueScore.Summarize(&Score{})
	assert.Equal(t, true, blueSummary.SuperchargedBonusRankingPoint)
	assert.Equal(t, 361, blueSummary.NumFuel)
	assert.Equal(t, 361, blueSummary.NumFuelGoal)

	// Fuel scored in inactive shifts still does not count toward the threshold.
	blueScore = TestScore2()
	blueScore.Hub.ShiftCounts[Shift1] += 500
	blueSummary = blueScore.Summarize(&Score{})
	assert.Equal(t, false, blueSummary.SuperchargedBonusRankingPoint)
	assert.Equal(t, 114, blueSummary.NumFuel)
}

func TestScoreTraversalBonusRankingPoint(t *testing.T) {
	originalThreshold := TraversalBonusThreshold
	defer func() {
		TraversalBonusThreshold = originalThreshold
	}()

	testCases := []struct {
		autoTowerStatuses    [3]TowerStatus
		endgameTowerStatuses [3]TowerStatus
		fouls                []Foul
		threshold            int
		expectedBonusAwarded bool
	}{
		// 0. No tower points.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerNone, TowerNone, TowerNone},
			endgameTowerStatuses: [3]TowerStatus{TowerNone, TowerNone, TowerNone},
			fouls:                []Foul{},
			threshold:            52,
			expectedBonusAwarded: false,
		},

		// 1. Only one robot counts and at Level 1 only.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerLevel1, TowerLevel3, TowerLevel2},
			endgameTowerStatuses: [3]TowerStatus{TowerNone, TowerNone, TowerNone},
			fouls:                []Foul{},
			threshold:            15,
			expectedBonusAwarded: true,
		},

		// 2. Meeting the threshold with auto and teleop tower points.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerLevel1, TowerNone, TowerNone},
			endgameTowerStatuses: [3]TowerStatus{TowerLevel2, TowerLevel1, TowerNone},
			fouls:                []Foul{},
			threshold:            45,
			expectedBonusAwarded: true,
		},

		// 3. The same tower statuses do not meet a higher threshold.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerNone, TowerLevel3, TowerNone},
			endgameTowerStatuses: [3]TowerStatus{TowerLevel2, TowerLevel1, TowerNone},
			fouls:                []Foul{},
			threshold:            46,
			expectedBonusAwarded: false,
		},

		// 4. All Level 3 climbs easily clear a large threshold.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerNone, TowerNone, TowerNone},
			endgameTowerStatuses: [3]TowerStatus{TowerLevel3, TowerLevel3, TowerLevel3},
			fouls:                []Foul{},
			threshold:            90,
			expectedBonusAwarded: true,
		},

		// 5. G206 makes the alliance ineligible for the traversal bonus.
		{
			autoTowerStatuses:    [3]TowerStatus{TowerNone, TowerNone, TowerLevel2},
			endgameTowerStatuses: [3]TowerStatus{TowerLevel3, TowerLevel3, TowerLevel3},
			fouls:                []Foul{{RuleId: 1}},
			threshold:            52,
			expectedBonusAwarded: false,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				TraversalBonusThreshold = tc.threshold
				score := Score{
					AutoTowerStatuses:    tc.autoTowerStatuses,
					EndgameTowerStatuses: tc.endgameTowerStatuses,
					Fouls:                tc.fouls,
				}
				summary := score.Summarize(&Score{})
				assert.Equal(t, tc.expectedBonusAwarded, summary.TraversalBonusRankingPoint)
			},
		)
	}
}

func TestScoreFuelBonusRankingPointActiveShifts(t *testing.T) {
	originalThreshold := EnergizedBonusThreshold
	defer func() {
		EnergizedBonusThreshold = originalThreshold
	}()
	EnergizedBonusThreshold = 149

	score := Score{
		Hub: Hub{
			WonAuto:     false,
			ShiftCounts: [ShiftCount]int{0, 0, 0, 71, 0, 79, 0},
		},
	}
	summary := score.Summarize(&Score{})
	assert.Equal(t, false, summary.EnergizedBonusRankingPoint)
	assert.Equal(t, 0, summary.NumFuel)

	score.Hub.WonAuto = true
	summary = score.Summarize(&Score{})
	assert.Equal(t, true, summary.EnergizedBonusRankingPoint)
	assert.Equal(t, 150, summary.NumFuel)
}

func TestScoreBonusRankingPointDisqualificationFromFouls(t *testing.T) {
	originalEnergizedThreshold := EnergizedBonusThreshold
	originalSuperchargedThreshold := SuperchargedBonusThreshold
	originalTraversalThreshold := TraversalBonusThreshold
	defer func() {
		EnergizedBonusThreshold = originalEnergizedThreshold
		SuperchargedBonusThreshold = originalSuperchargedThreshold
		TraversalBonusThreshold = originalTraversalThreshold
	}()
	EnergizedBonusThreshold = 89
	SuperchargedBonusThreshold = 299
	TraversalBonusThreshold = 44

	testCases := []struct {
		score                    Score
		expectedEnergizedBonus   bool
		expectedSupercharged     bool
		expectedTraversalBonus   bool
		expectedBonusRankingPoin int
	}{
		// 0. All bonus ranking points are awarded.
		{
			score: Score{
				Hub: Hub{
					WonAuto:     false,
					ShiftCounts: [ShiftCount]int{60, 60, 80, 0, 80, 0, 19},
				},
				AutoTowerStatuses:    [3]TowerStatus{TowerLevel1, TowerNone, TowerNone},
				EndgameTowerStatuses: [3]TowerStatus{TowerLevel3, TowerLevel3, TowerNone},
			},
			expectedEnergizedBonus:   true,
			expectedSupercharged:     true,
			expectedTraversalBonus:   true,
			expectedBonusRankingPoin: 3,
		},

		// 1. G206 removes all bonus ranking points.
		{
			score: Score{
				Hub: Hub{
					WonAuto:     false,
					ShiftCounts: [ShiftCount]int{60, 60, 80, 0, 80, 0, 19},
				},
				AutoTowerStatuses:    [3]TowerStatus{TowerLevel1, TowerNone, TowerNone},
				EndgameTowerStatuses: [3]TowerStatus{TowerLevel3, TowerLevel3, TowerNone},
				Fouls:                []Foul{{RuleId: 1}},
			},
			expectedEnergizedBonus:   false,
			expectedSupercharged:     false,
			expectedTraversalBonus:   false,
			expectedBonusRankingPoin: 0,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				summary := tc.score.Summarize(&Score{})
				assert.Equal(t, tc.expectedEnergizedBonus, summary.EnergizedBonusRankingPoint)
				assert.Equal(t, tc.expectedSupercharged, summary.SuperchargedBonusRankingPoint)
				assert.Equal(t, tc.expectedTraversalBonus, summary.TraversalBonusRankingPoint)
				assert.Equal(t, tc.expectedBonusRankingPoin, summary.BonusRankingPoints)
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
	score2.AutoTowerStatuses[0] = TowerLevel1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Hub.WonAuto = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Hub.ShiftCounts[Shift3]++
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameTowerStatuses[1] = TowerLevel3
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
	score2.Fouls[0].TeamId++
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
