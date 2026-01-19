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
	// 模擬紅隊分數 (贏家)
	redSummary := &ScoreSummary{
		MatchPoints:        90,
		AutoPoints:         25,
		TotalTowerPoints:   30, // 2026 爬升分
		Score:              100,
		BonusRankingPoints: 1, // 例如拿到 Energized RP
	}

	// 模擬藍隊分數 (輸家)
	blueSummary := &ScoreSummary{
		MatchPoints:        50,
		AutoPoints:         10,
		TotalTowerPoints:   20,
		Score:              50,
		BonusRankingPoints: 0,
	}

	rankingFields := RankingFields{}

	// 測試 1: 加入一場敗場 (Add a loss)
	// 假設我們是藍隊，對手是紅隊
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	// 預期: 0 RP(輸) + 0 Bonus = 0 RP. 1 Loss. MatchPoints=50.
	assert.Equal(t, 0, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Losses)
	assert.Equal(t, 50, rankingFields.MatchPoints)

	// 測試 2: 加入一場勝場 (Add a win)
	// 假設我們是紅隊，對手是藍隊
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	// 預期: 3 RP(贏) + 1 Bonus = 4 RP. 1 Win. MatchPoints=90.
	assert.Equal(t, 4, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Wins)
	assert.Equal(t, 90, rankingFields.MatchPoints)
	assert.Equal(t, 30, rankingFields.TowerPoints) // 檢查 TowerPoints 是否正確記錄

	// 測試 3: 加入一場平手 (Add a tie)
	rankingFields = RankingFields{}
	tieScore := &ScoreSummary{Score: 80, MatchPoints: 80}
	rankingFields.AddScoreSummary(tieScore, tieScore, false)
	// 預期: 1 RP(平) = 1 RP. 1 Tie.
	assert.Equal(t, 1, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Ties)

	// 測試 4: 犯規被取消資格 (Disqualification)
	rankingFields = RankingFields{}
	rankingFields.AddScoreSummary(redSummary, blueSummary, true)
	// 預期: 0 RP. 1 DQ.
	assert.Equal(t, 0, rankingFields.RankingPoints)
	assert.Equal(t, 1, rankingFields.Disqualifications)
}
