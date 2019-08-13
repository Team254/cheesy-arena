// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/scoring/invalidalliance")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid alliance")
	recorder = web.getHttpResponse("/panels/scoring/red")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))
	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// Send a some pre-match scoring commands.
	redWs.Write("1", nil)
	blueWs.Write("2", nil)
	blueWs.Write("2", nil)
	blueWs.Write("2", nil)
	blueWs.Write("2", nil)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 1, web.arena.RedRealtimeScore.CurrentScore.RobotStartLevels[0])
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.RobotStartLevels[1])
	redWs.Write("e", nil)
	redWs.Write("i", nil)
	redWs.Write("i", nil)
	redWs.Write("v", nil)
	redWs.Write("q", nil)
	redWs.Write(",", nil)
	for i := 0; i < 3; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [8]game.BayStatus{1, 0, 0, 0, 0, 0, 0, 3},
		web.arena.RedRealtimeScore.CurrentScore.CargoBaysPreMatch)
	assert.Equal(t, [8]game.BayStatus{1, 0, 0, 0, 0, 0, 0, 3}, web.arena.RedRealtimeScore.CurrentScore.CargoBays)
	assert.Equal(t, [3]game.BayStatus{0, 0, 0}, web.arena.RedRealtimeScore.CurrentScore.RocketNearLeftBays)
	assert.Equal(t, [3]game.BayStatus{0, 0, 0}, web.arena.RedRealtimeScore.CurrentScore.RocketNearRightBays)
	assert.Equal(t, [3]game.BayStatus{0, 0, 0}, web.arena.RedRealtimeScore.CurrentScore.RocketFarLeftBays)
	assert.Equal(t, [3]game.BayStatus{0, 0, 0}, web.arena.RedRealtimeScore.CurrentScore.RocketFarRightBays)

	// Send some in-match scoring commands.
	web.arena.MatchState = field.AutoPeriod
	redWs.Write("e", nil)
	redWs.Write("i", nil)
	redWs.Write("k", nil)
	redWs.Write("4", nil)
	blueWs.Write("9", nil)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [8]game.BayStatus{1, 0, 0, 0, 0, 0, 0, 3},
		web.arena.RedRealtimeScore.CurrentScore.CargoBaysPreMatch)
	assert.Equal(t, [8]game.BayStatus{2, 0, 0, 0, 0, 0, 0, 2}, web.arena.RedRealtimeScore.CurrentScore.CargoBays)
	assert.Equal(t, [3]game.BayStatus{0, 1, 0}, web.arena.RedRealtimeScore.CurrentScore.RocketFarRightBays)
	assert.True(t, web.arena.RedRealtimeScore.CurrentScore.SandstormBonuses[0])
	assert.Equal(t, 1, web.arena.BlueRealtimeScore.CurrentScore.RobotEndLevels[2])

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
}
