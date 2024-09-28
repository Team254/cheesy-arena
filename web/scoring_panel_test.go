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

	// Send some autonomous period scoring commands.
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)
	scoringData := struct {
		TeamPosition int
		StageIndex   int
	}{}
	web.arena.MatchState = field.AutoPeriod
	scoringData.TeamPosition = 1
	redWs.Write("leave", scoringData)
	scoringData.TeamPosition = 3
	redWs.Write("leave", scoringData)
	for i := 0; i < 2; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [3]bool{true, false, true}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)
	redWs.Write("leave", scoringData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, [3]bool{true, false, false}, web.arena.RedRealtimeScore.CurrentScore.LeaveStatuses)

	// Send some teleoperated period scoring commands.
	web.arena.MatchState = field.TeleopPeriod
	scoringData.TeamPosition = 1
	scoringData.StageIndex = 0
	blueWs.Write("onStage", scoringData)
	scoringData.TeamPosition = 2
	scoringData.StageIndex = 1
	blueWs.Write("onStage", scoringData)
	scoringData.TeamPosition = 3
	scoringData.StageIndex = 2
	redWs.Write("onStage", scoringData)
	redWs.Write("microphone", scoringData)
	scoringData.StageIndex = 0
	redWs.Write("trap", scoringData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameStageLeft, game.EndgameCenterStage, game.EndgameNone},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameStatuses,
	)
	assert.Equal(t, [3]bool{false, false, false}, web.arena.BlueRealtimeScore.CurrentScore.MicrophoneStatuses)
	assert.Equal(t, [3]bool{false, false, false}, web.arena.BlueRealtimeScore.CurrentScore.TrapStatuses)
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameNone, game.EndgameNone, game.EndgameStageRight},
		web.arena.RedRealtimeScore.CurrentScore.EndgameStatuses,
	)
	assert.Equal(t, [3]bool{false, false, true}, web.arena.RedRealtimeScore.CurrentScore.MicrophoneStatuses)
	assert.Equal(t, [3]bool{true, false, false}, web.arena.RedRealtimeScore.CurrentScore.TrapStatuses)
	scoringData.StageIndex = 1
	redWs.Write("trap", scoringData)
	scoringData.StageIndex = 0
	redWs.Write("trap", scoringData)
	scoringData.StageIndex = 2
	redWs.Write("microphone", scoringData)
	scoringData.TeamPosition = 1
	blueWs.Write("park", scoringData)
	scoringData.TeamPosition = 2
	scoringData.StageIndex = 1
	blueWs.Write("onStage", scoringData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameParked, game.EndgameNone, game.EndgameNone},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameStatuses,
	)
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.MicrophoneStatuses)
	assert.Equal(t, [3]bool{false, true, false}, web.arena.RedRealtimeScore.CurrentScore.TrapStatuses)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)
	scoringData.TeamPosition = 0
	redWs.Write("leave", scoringData)
	scoringData.TeamPosition = 4
	redWs.Write("onStage", scoringData)
	scoringData.TeamPosition = 1
	scoringData.StageIndex = 3
	blueWs.Write("onStage", scoringData)

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
