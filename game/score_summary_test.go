// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	// 初始化：雙方平手
	redScoreSummary := &ScoreSummary{Score: 50}
	blueScoreSummary := &ScoreSummary{Score: 50}

	// 1. 總分相同 -> Tie
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))

	// 2. 測試 Tiebreaker 1: 總分 (Score)
	redScoreSummary.Score = 51
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.Score = 50 // 還原

	// 3. 測試 Tiebreaker 2: 對手犯規數 (NumOpponentMajorFouls)
	redScoreSummary.NumOpponentMajorFouls = 2
	blueScoreSummary.NumOpponentMajorFouls = 1
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.NumOpponentMajorFouls = 0 // 還原
	blueScoreSummary.NumOpponentMajorFouls = 0

	// 4. 測試 Tiebreaker 3: Auto 分數 (AutoPoints)
	blueScoreSummary.AutoPoints = 20
	redScoreSummary.AutoPoints = 10
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	blueScoreSummary.AutoPoints = 10 // 還原

	// 5. 測試 Tiebreaker 4: 爬升總分 (TotalTowerPoints) - 取代去年的 Barge
	redScoreSummary.TotalTowerPoints = 30
	blueScoreSummary.TotalTowerPoints = 20
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	// 如果連爬升分都一樣 -> Tie
	blueScoreSummary.TotalTowerPoints = 30
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
}
