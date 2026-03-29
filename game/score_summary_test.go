// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	// Initialization: A draw
	redScoreSummary := &ScoreSummary{Score: 50}
	blueScoreSummary := &ScoreSummary{Score: 50}

	// 1. Same total score -> Tie
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))

	// 2. Test Tiebreaker 1: Total score (Score)
	redScoreSummary.Score = 51
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.Score = 50 // Reset

	// 3. Test Tiebreaker 2: Opponent fouls (NumOpponentMajorFouls)
	redScoreSummary.NumOpponentMajorFouls = 2
	blueScoreSummary.NumOpponentMajorFouls = 1
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.NumOpponentMajorFouls = 0 // Reset
	blueScoreSummary.NumOpponentMajorFouls = 0

	// 4. Test Tiebreaker 3: Auto points (AutoPoints)
	blueScoreSummary.AutoPoints = 20
	redScoreSummary.AutoPoints = 10
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	blueScoreSummary.AutoPoints = 10 // Reset

	// 5. Test Tiebreaker 4: Total tower points (TotalTowerPoints) - Replaces last year's Barge
	redScoreSummary.TotalTowerPoints = 30
	blueScoreSummary.TotalTowerPoints = 20
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	// If total tower points are the same -> Tie
	blueScoreSummary.TotalTowerPoints = 30
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
}
