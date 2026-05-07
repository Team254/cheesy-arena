// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package web

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	// 測試基本的頁面存取 (確保 URL 路由正確)
	recorder := web.getHttpResponse("/panels/scoring/invalidposition")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid position")

	recorder = web.getHttpResponse("/panels/scoring/red_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/red_far")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_far")
	assert.Equal(t, 200, recorder.Code)

	// 確保標題正確 (表示 Template Render 成功)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()

	// 1. 測試 WebSocket 連線建立
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err) // 無效位置應報錯

	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_near/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)

	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))

	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue_near/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)

	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))

	// 2. 接收初始狀態更新 (Handshake)
	readWebsocketType(t, redWs, "resetLocalState")
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "resetLocalState")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// --- 2026 測試開始: Auto Period ---
	web.arena.MatchState = field.AutoPeriod

	// 3. 測試 Auto Tower (Level 1)
	// 用來取代原本的 Leave/Reef 測試
	autoTowerData := struct {
		RobotIndex int
		Adjustment int
	}{}
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.AutoTowerLevel1)

	// 設定 Robot 1 完成 Auto Tower
	autoTowerData.RobotIndex = 0
	autoTowerData.Adjustment = 1 // True
	redWs.Write("auto_tower", autoTowerData)

	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, [3]bool{true, false, false}, web.arena.RedRealtimeScore.CurrentScore.AutoTowerLevel1)

	// 設定 Robot 2 完成 Auto Tower
	autoTowerData.RobotIndex = 1
	redWs.Write("auto_tower", autoTowerData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, [3]bool{true, true, false}, web.arena.RedRealtimeScore.CurrentScore.AutoTowerLevel1)

	// 4. 測試 Fuel (Auto)
	fuelData := struct {
		Adjustment int
		Autonomous bool
	}{}
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.AutoFuelCount)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.AutoFuelCount)

	// 紅隊 Auto 進球 +1
	fuelData.Adjustment = 1
	fuelData.Autonomous = true
	redWs.Write("fuel", fuelData)
	redWs.Write("fuel", fuelData) // 總共 +2

	for i := 0; i < 2; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.AutoFuelCount)

	// 藍隊 Auto 進球 +5 (一次性)
	fuelData.Adjustment = 5
	blueWs.Write("fuel", fuelData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, 5, web.arena.BlueRealtimeScore.CurrentScore.AutoFuelCount)

	// --- 2026 測試開始: Teleop Period ---
	web.arena.MatchState = field.TeleopPeriod

	// 5. 測試 Fuel (Teleop)
	fuelData.Autonomous = false
	fuelData.Adjustment = 1

	// 紅隊 Teleop 進球
	redWs.Write("fuel", fuelData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")

	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.AutoFuelCount) // Auto 應保持不變
	assert.Equal(t, 1, web.arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)

	// 紅隊 Teleop 扣分 (修正)
	fuelData.Adjustment = -1
	redWs.Write("fuel", fuelData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)

	// 6. 測試 Climb (Endgame)
	climbData := struct {
		RobotIndex int
		Level      int
	}{}
	assert.Equal(t, game.EndgameNone, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[0])

	// Robot 1 爬升 Level 3
	climbData.RobotIndex = 0
	climbData.Level = 3
	redWs.Write("climb", climbData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.EndgameLevel3, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[0])

	// Robot 2 爬升 Level 2
	climbData.RobotIndex = 1
	climbData.Level = 2
	redWs.Write("climb", climbData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.EndgameLevel2, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[1])

	// Robot 1 修正為 None
	climbData.RobotIndex = 0
	climbData.Level = 0
	redWs.Write("climb", climbData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.EndgameNone, web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses[0])

	// 7. 測試無效指令與 Commit 邏輯 (保持原有架構)
	redWs.Write("invalid", nil) // 應被忽略

	// 測試：比賽未結束不能 Commit
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))

	// 測試：比賽結束後 Commit
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // 等待處理
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))

	// 8. 測試 Reset (載入新比賽)
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")

	// 驗證分數歸零
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
}
