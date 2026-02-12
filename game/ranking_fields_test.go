// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Tests for ranking logic.

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddScoreSummary(t *testing.T) {
	// Simulate the Red Team's score (winner)
	redSummary := &ScoreSummary{
		MatchPoints:        90,
		AutoPoints:         25,
		TotalTowerPoints:   30, // 2026 Climb Points
		Score:              100,
		BonusRankingPoints: 1, // e.g., Energized RP
	}

	// Simulate the Blue Team's score (loser)
	blueSummary := &ScoreSummary{
		MatchPoints:        50,
		AutoPoints:         10,
		TotalTowerPoints:   20,
		Score:              50,
		BonusRankingPoints: 0,
	}

	rankingFields := RankingFields{}

	// Test 1: Add a loss (Add a loss)
	// Assume we are the Blue Team, opponent is Red Team
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	// Expected: 0 RP (loss) + 0 Bonus = 0 RP. 1 Loss. MatchPoints=50.
	assert.Equal(t, 0, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Losses)
	assert.Equal(t, 50, rankingFields.MatchPoints)

	// Test 2: Add a win (Add a win)
	// Assume we are the Red Team, opponent is Blue Team
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	// Expected: 3 RP (win) + 1 Bonus = 4 RP. 1 Win. MatchPoints=90.
	assert.Equal(t, 4, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Wins)
	assert.Equal(t, 90, rankingFields.MatchPoints)
	assert.Equal(t, 30, rankingFields.TowerPoints) // Check if TowerPoints are recorded correctly

	// Test 3: Add a tie (Add a tie)
	rankingFields = RankingFields{}
	tieScore := &ScoreSummary{Score: 80, MatchPoints: 80}
	rankingFields.AddScoreSummary(tieScore, tieScore, false)
	// Expected: 1 RP (tie) = 1 RP. 1 Tie.
	assert.Equal(t, 1, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Ties)

	// Test 4: Disqualification (Disqualification)
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(redSummary, blueSummary, true)
	// Expected: 0 RP. 1 DQ.
	assert.Equal(t, 0, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Disqualifications)
}
