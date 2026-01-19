// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package field

import (
	"os"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestTeamSign_GenerateInMatchRearText(t *testing.T) {
	arena := setupTestArena(t)
	defer os.Remove("test.db")

	// 載入 2026 模擬分數
	arena.RedRealtimeScore.CurrentScore = *game.TestScore1()
	arena.BlueRealtimeScore.CurrentScore = *game.TestScore2()

	// 設定為資格賽 (顯示 RP 進度/球數)
	arena.CurrentMatch = &model.Match{Type: model.Qualification}

	// 預期分數計算:
	// Red Score: 95 (Auto:25 + Teleop:20 + End:50)
	// Blue Score: 130 (Auto:20 + Teleop:40 + End:70)
	// Red Fuel: 25
	// Blue Fuel: 50

	// 測試紅隊視角: "倒數時間 R分數-B分數 R球數"
	// R095-B130 25
	assert.Equal(t, "01:23 R095-B130 25", generateInMatchTeamRearText(arena, true, "01:23"))

	// 測試藍隊視角: "倒數時間 B分數-R分數 B球數"
	// B130-R095 50
	assert.Equal(t, "01:23 B130-R095 50", generateInMatchTeamRearText(arena, false, "01:23"))
}

func TestTeamSign_GenerateInMatchTimerRearText(t *testing.T) {
	arena := setupTestArena(t)
	defer os.Remove("test.db")

	arena.RedRealtimeScore.CurrentScore = *game.TestScore1()
	arena.BlueRealtimeScore.CurrentScore = *game.TestScore2()

	// 測試計時器背板顯示 (Timer Rear Text)
	// 格式: A-[AutoFuel] T-[TeleopFuel] Tot-[TotalFuel]

	// 紅隊: Auto=5, Teleop=20, Total=25
	assert.Equal(t, "A-05 T-20 Tot-25", generateInMatchTimerRearText(arena, true))

	// 藍隊: Auto=10, Teleop=40, Total=50
	assert.Equal(t, "A-10 T-40 Tot-50", generateInMatchTimerRearText(arena, false))
}
