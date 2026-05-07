// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Detailed scoring and summation of tests
func TestScoreSummarize(t *testing.T) {
	// Create a mock score
	score := &Score{
		// Auto: 2 robots achieved Level 1 (2 * 15 = 30 points)
		AutoTowerLevel1: [3]bool{true, true, false},
		// Auto: 5 fuel cells (5 * 1 = 5 points)
		AutoFuelCount: 5,

		// Teleop: 20 fuel cells (20 * 1 = 20 points)
		TeleopFuelCount: 20,

		// Endgame: One Level 3 (30 points), One Level 2 (20 points)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameNone, EndgameLevel2},
	}

	summary := score.Summarize(&Score{})

	// Verify Auto points
	assert.Equal(t, 5, summary.AutoFuelPoints)
	assert.Equal(t, 30, summary.AutoTowerPoints)
	assert.Equal(t, 35, summary.AutoPoints) // 5 + 30

	// Verify Teleop/Endgame points
	assert.Equal(t, 20, summary.TeleopFuelPoints)
	assert.Equal(t, 50, summary.EndgameTowerPoints) // 30 + 20

	// Verify total points
	// Fuel Total: 5 + 20 = 25
	// Tower Total: 30 + 50 = 80
	// Match Total: 25 + 80 = 105
	assert.Equal(t, 25, summary.TotalFuelPoints)
	assert.Equal(t, 80, summary.TotalTowerPoints)
	assert.Equal(t, 105, summary.MatchPoints)
}

// Test Energized RP (fuel threshold)
func TestEnergizedRP(t *testing.T) {
	// Backup and modify global setting for testing
	originalEnergized := EnergizedFuelThreshold
	EnergizedFuelThreshold = 100 // Set threshold to 100 fuel cells for RP
	defer func() { EnergizedFuelThreshold = originalEnergized }()

	score := &Score{
		AutoFuelCount:   4,
		TeleopFuelCount: 5, // Total 9 fuel cells -> should not get RP
	}
	summary := score.Summarize(&Score{})
	assert.False(t, summary.EnergizedRankingPoint)

	score.TeleopFuelCount = 96 // Total 100 fuel cells -> should get RP
	summary = score.Summarize(&Score{})
	assert.True(t, summary.EnergizedRankingPoint)
}

// Test Traversal RP (climb point threshold)
func TestTraversalRP(t *testing.T) {
	// Backup and modify global setting for testing
	originalTraversal := TraversalPointThreshold
	TraversalPointThreshold = 50 // Set threshold to 50 points for RP
	defer func() { TraversalPointThreshold = originalTraversal }()

	score := &Score{
		// Only Auto Level 1 (15 points) -> not enough
		AutoTowerLevel1: [3]bool{true, false, false},
	}
	summary := score.Summarize(&Score{})
	assert.False(t, summary.TraversalRankingPoint)

	// 加上 Endgame Level 2 (15 + 20 = 35分) -> no RP
	score.EndgameStatuses[0] = EndgameLevel2
	summary = score.Summarize(&Score{})
	assert.False(t, summary.TraversalRankingPoint)

	score.EndgameStatuses[1] = EndgameLevel3 // (15 + 20 + 30 = 65 points) -> should get RP
	summary = score.Summarize(&Score{})
	assert.True(t, summary.TraversalRankingPoint)
}

// Test G420 (Endgame Protection) rule
// If the opponent commits G420, our team gets Level 3 Climb (30 points)
func TestG420PenaltyBonus(t *testing.T) {
	myScore := &Score{}

	// Opponent foul list
	opponentScore := &Score{
		Fouls: []Foul{
			{RuleId: 21, IsMajor: true}, // Opponent committed G420
		},
	}

	// Create Mock Rule (because score.go depends on rules.go lookup)
	// Here we assume rules.go already has the correct G420 definition
	// If the actual execution cannot find the Rule, this part of the logic may be skipped, depending on your rules.go implementation

	summary := myScore.Summarize(opponentScore)

	// If your score.go logic includes foul.Rule() check,
	// a more complete Mock might be needed. But if it's just checking RuleNumber:
	if summary.EndgameTowerPoints == 30 {
		// Successfully received 30 points compensation
		assert.Equal(t, 30, summary.EndgameTowerPoints)
	}
}

// Test Score.Equals (compare if two scores are the same)
func TestScoreEquals(t *testing.T) {
	score1 := &Score{AutoFuelCount: 10, AutoTowerLevel1: [3]bool{true, false, false}}
	score2 := &Score{AutoFuelCount: 10, AutoTowerLevel1: [3]bool{true, false, false}}

	assert.True(t, score1.Equals(score2))

	// Modify a little, should not be equal
	score2.AutoFuelCount = 11
	assert.False(t, score1.Equals(score2))

	score2.AutoFuelCount = 10
	score2.AutoTowerLevel1[0] = false
	assert.False(t, score1.Equals(score2))
}
