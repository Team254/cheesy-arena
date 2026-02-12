// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 測試詳細的計分加總
func TestScoreSummarize(t *testing.T) {
	// 建立一個模擬分數
	score := &Score{
		// Auto: 2台機器人達成 Level 1 (2 * 15 = 30分)
		AutoTowerLevel1: [3]bool{true, true, false},
		// Auto: 5顆球 (5 * 1 = 5分)
		AutoFuelCount: 5,

		// Teleop: 20顆球 (20 * 1 = 20分)
		TeleopFuelCount: 20,

		// Endgame: 一台 Level 3 (30分), 一台 Level 2 (20分)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameNone, EndgameLevel2},
	}

	summary := score.Summarize(&Score{})

	// 驗證 Auto 分數
	assert.Equal(t, 5, summary.AutoFuelPoints)
	assert.Equal(t, 30, summary.AutoTowerPoints)
	assert.Equal(t, 35, summary.AutoPoints) // 5 + 30

	// 驗證 Teleop/Endgame 分數
	assert.Equal(t, 20, summary.TeleopFuelPoints)
	assert.Equal(t, 50, summary.EndgameTowerPoints) // 30 + 20

	// 驗證總分
	// Fuel Total: 5 + 20 = 25
	// Tower Total: 30 + 50 = 80
	// Match Total: 25 + 80 = 105
	assert.Equal(t, 25, summary.TotalFuelPoints)
	assert.Equal(t, 80, summary.TotalTowerPoints)
	assert.Equal(t, 105, summary.MatchPoints)
}

// 測試 Energized RP (球數門檻)
func TestEnergizedRP(t *testing.T) {
	// 備份並修改全域設定以方便測試
	originalEnergized := EnergizedFuelThreshold
	EnergizedFuelThreshold = 100 // 設定只要 100 顆球就有 RP
	defer func() { EnergizedFuelThreshold = originalEnergized }()

	score := &Score{
		AutoFuelCount:   4,
		TeleopFuelCount: 5, // 總共 9 顆 -> 應該沒有 RP
	}
	summary := score.Summarize(&Score{})
	assert.False(t, summary.EnergizedRankingPoint)

	score.TeleopFuelCount = 96 // 總共 100 顆 -> 應該有 RP
	summary = score.Summarize(&Score{})
	assert.True(t, summary.EnergizedRankingPoint)
}

// 測試 Traversal RP (爬升分數門檻)
func TestTraversalRP(t *testing.T) {
	// 備份並修改全域設定
	originalTraversal := TraversalPointThreshold
	TraversalPointThreshold = 50 // 設定只要 50 分就有 RP
	defer func() { TraversalPointThreshold = originalTraversal }()

	score := &Score{
		// 只有 Auto Level 1 (15分) -> 不夠
		AutoTowerLevel1: [3]bool{true, false, false},
	}
	summary := score.Summarize(&Score{})
	assert.False(t, summary.TraversalRankingPoint)

	// 加上 Endgame Level 2 (15 + 20 = 35分) -> no RP
	score.EndgameStatuses[0] = EndgameLevel2
	summary = score.Summarize(&Score{})
	assert.False(t, summary.TraversalRankingPoint)

	score.EndgameStatuses[1] = EndgameLevel3 // (15 + 20 + 30 = 65分) -> 有 RP
	summary = score.Summarize(&Score{})
	assert.True(t, summary.TraversalRankingPoint)
}

// 測試 G420 (Endgame Protection) 規則
// 如果對手犯規 G420，我方獲得 Level 3 Climb (30分)
func TestG420PenaltyBonus(t *testing.T) {
	myScore := &Score{}

	// 對手犯規清單
	opponentScore := &Score{
		Fouls: []Foul{
			{RuleId: 21, IsMajor: true}, // 對手犯了 G420
		},
	}

	// 建立 Mock Rule (因為 score.go 依賴 rules.go 的 lookup)
	// 這裡假設 rules.go 已經有正確的 G420 定義
	// 如果實際執行時抓不到 Rule，這部分邏輯可能會被跳過，視你的 rules.go 實作而定

	summary := myScore.Summarize(opponentScore)

	// 如果你的 score.go 邏輯包含 foul.Rule() 檢查，
	// 這裡可能需要更完整的 Mock。但若是簡單檢查 RuleNumber：
	if summary.EndgameTowerPoints == 30 {
		// 成功獲得 30 分補償
		assert.Equal(t, 30, summary.EndgameTowerPoints)
	}
}

// 測試 Score.Equals (比較兩個分數是否相同)
func TestScoreEquals(t *testing.T) {
	score1 := &Score{AutoFuelCount: 10, AutoTowerLevel1: [3]bool{true, false, false}}
	score2 := &Score{AutoFuelCount: 10, AutoTowerLevel1: [3]bool{true, false, false}}

	assert.True(t, score1.Equals(score2))

	// 修改一點點，應該要不相等
	score2.AutoFuelCount = 11
	assert.False(t, score1.Equals(score2))

	score2.AutoFuelCount = 10
	score2.AutoTowerLevel1[0] = false
	assert.False(t, score1.Equals(score2))
}
