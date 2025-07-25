// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestReefCoralCountsAndPoints(t *testing.T) {
	testCases := []struct {
		reef                      Reef
		expectedTotalCountByLevel [4]int
		expectedAutoCount         int
		expectedAutoPoints        int
		expectedTeleopCount       int
		expectedTeleopPoints      int
	}{
		// 0. Empty Reef.
		{
			Reef{}, [4]int{0, 0, 0, 0}, 0, 0, 0, 0,
		},

		// 1. Only auto branches which have all been de-scored.
		{
			Reef{
				AutoBranches: [3][12]bool{
					{true, false, false, true, false, false, true, false, false, true, false, false},
					{false, true, false, false, true, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, true},
				},
				Branches: [3][12]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
				},
				AutoTroughNear: 1,
				AutoTroughFar:  2,
				TroughNear:     0,
				TroughFar:      0,
			},
			[4]int{0, 0, 0, 0},
			0,
			0,
			0,
			0,
		},

		// 2. Only auto branches.
		{
			Reef{
				AutoBranches: [3][12]bool{
					{true, false, false, true, false, false, true, false, false, true, false, false},
					{false, true, false, false, true, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, true},
				},
				Branches: [3][12]bool{
					{true, false, false, true, false, false, true, false, false, true, false, false},
					{false, true, false, false, true, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, true},
				},
				AutoTroughNear: 1,
				AutoTroughFar:  2,
				TroughNear:     2,
				TroughFar:      1,
			},
			[4]int{3, 4, 3, 2},
			12,
			57,
			0,
			0,
		},

		// 3. Only teleop branches.
		{
			Reef{
				AutoBranches: [3][12]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
				},
				Branches: [3][12]bool{
					{true, false, false, true, false, false, true, false, false, true, false, false},
					{false, true, false, false, true, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, true},
				},
				AutoTroughNear: 0,
				AutoTroughFar:  0,
				TroughNear:     2,
				TroughFar:      1,
			},
			[4]int{3, 4, 3, 2},
			0,
			0,
			12,
			40,
		},

		// 4. Full Reef with some auto scoring.
		{
			Reef{
				AutoBranches: [3][12]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{true, true, false, true, true, true, false, true, false, false, false, false},
				},
				Branches: [3][12]bool{
					{true, true, true, true, true, true, true, true, true, true, true, true},
					{true, true, true, true, true, true, true, true, true, true, true, true},
					{true, true, true, true, true, true, true, true, true, true, true, true},
				},
				AutoTroughNear: 1,
				AutoTroughFar:  0,
				TroughNear:     12,
				TroughFar:      12,
			},
			[4]int{24, 12, 12, 12},
			7,
			45,
			53,
			160,
		},
	}

	for i, testCase := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				for level := Level1; level < LevelCount; level++ {
					assert.Equal(
						t, testCase.expectedTotalCountByLevel[level+1], testCase.reef.CountTotalCoralByLevel(level),
					)
				}
				assert.Equal(t, testCase.expectedAutoCount, testCase.reef.AutoCoralCount())
				assert.Equal(t, testCase.expectedAutoPoints, testCase.reef.AutoCoralPoints())
				assert.Equal(t, testCase.expectedTeleopCount, testCase.reef.TeleopCoralCount())
				assert.Equal(t, testCase.expectedTeleopPoints, testCase.reef.TeleopCoralPoints())
			},
		)
	}
}

func TestReef_isAutoBonusCoralThresholdMet(t *testing.T) {
	// Save the original threshold value and restore it after the test.
	originalThreshold := AutoBonusCoralThreshold
	defer func() {
		AutoBonusCoralThreshold = originalThreshold
	}()

	testCases := []struct {
		reef                 Reef
		threshold            int
		expectedThresholdMet bool
	}{
		// 0. Empty reef.
		{
			reef:                 Reef{},
			threshold:            1,
			expectedThresholdMet: false,
		},

		// 1. Below threshold with some coral.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{true, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
				},
				AutoTroughNear: 2,
				AutoTroughFar:  1,
			},
			threshold:            5,
			expectedThresholdMet: false,
		},

		// 2. Exactly at threshold.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{true, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
				},
				AutoTroughNear: 2,
				AutoTroughFar:  1,
			},
			threshold:            4,
			expectedThresholdMet: true,
		},

		// 3. Above threshold.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{true, true, false, false, false, false, false, false, false, false, false, false},
					{true, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, true},
				},
				AutoTroughNear: 5,
				AutoTroughFar:  3,
			},
			threshold:            10,
			expectedThresholdMet: true,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				AutoBonusCoralThreshold = tc.threshold
				result := tc.reef.isAutoBonusCoralThresholdMet()
				assert.Equal(t, tc.expectedThresholdMet, result)
			},
		)
	}
}

func TestReef_countCoralBonusSatisfiedLevels(t *testing.T) {
	// Save the original threshold value and restore it after the test.
	originalThreshold := CoralBonusPerLevelThreshold
	defer func() {
		CoralBonusPerLevelThreshold = originalThreshold
	}()

	testCases := []struct {
		reef                    Reef
		threshold               int
		expectedSatisfiedLevels int
	}{
		// 0. Empty reef.
		{
			reef:                    Reef{},
			threshold:               1,
			expectedSatisfiedLevels: 0,
		},

		// 1. Two levels satisfied.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{true, true, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false},
				},
				Branches: [3][12]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false},
					{true, false, true, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, true, true, true},
				},
				AutoTroughNear: 1,
				AutoTroughFar:  0,
				TroughNear:     2,
				TroughFar:      1,
			},
			threshold:               3,
			expectedSatisfiedLevels: 2,
		},

		// 2. All levels satisfied.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{false, false, false, false, false, false, true, true, true, true, true, true},
					{true, true, false, false, false, false, false, false, false, false, false, false},
					{true, true, false, false, false, false, false, false, false, false, false, false},
				},
				Branches: [3][12]bool{
					{false, false, false, false, false, true, true, true, true, true, true, true},
					{true, true, true, true, false, true, false, false, true, true, false, false},
					{true, true, true, true, true, false, false, false, false, false, true, true},
				},
				AutoTroughNear: 3,
				AutoTroughFar:  2,
				TroughNear:     2,
				TroughFar:      5,
			},
			threshold:               7,
			expectedSatisfiedLevels: 4,
		},

		// 3. No levels satisfied with higher threshold and same scoring as above.
		{
			reef: Reef{
				AutoBranches: [3][12]bool{
					{false, false, false, false, false, true, true, true, true, true, true, true},
					{true, true, false, false, false, false, false, false, false, false, false, false},
					{true, true, false, false, false, false, false, false, false, false, false, false},
				},
				Branches: [3][12]bool{
					{false, false, false, false, false, false, true, true, true, true, true, true},
					{true, true, true, true, false, false, false, false, true, true, false, false},
					{true, true, true, true, true, false, false, false, false, false, true, true},
				},
				AutoTroughNear: 3,
				AutoTroughFar:  2,
				TroughNear:     2,
				TroughFar:      5,
			},
			threshold:               8,
			expectedSatisfiedLevels: 0,
		},
	}

	for i, tc := range testCases {
		t.Run(
			strconv.Itoa(i),
			func(t *testing.T) {
				CoralBonusPerLevelThreshold = tc.threshold
				result := tc.reef.countCoralBonusSatisfiedLevels()
				assert.Equal(t, tc.expectedSatisfiedLevels, result)
			},
		)
	}
}
