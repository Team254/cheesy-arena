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
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
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

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "resetLocalState")
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "resetLocalState")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// Send some endgame scoring commands.
	endgameData := struct {
		TeamPosition       int
		EndgameTowerStatus int
	}{}
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerNone, game.TowerNone, game.TowerNone},
		web.arena.RedRealtimeScore.CurrentScore.EndgameTowerStatuses,
	)
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerNone, game.TowerNone, game.TowerNone},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameTowerStatuses,
	)
	endgameData.TeamPosition = 1
	endgameData.EndgameTowerStatus = 2
	redWs.Write("endgame", endgameData)
	endgameData.TeamPosition = 2
	endgameData.EndgameTowerStatus = 3
	blueWs.Write("endgame", endgameData)
	endgameData.TeamPosition = 3
	endgameData.EndgameTowerStatus = 1
	blueWs.Write("endgame", endgameData)
	endgameData.TeamPosition = 3
	endgameData.EndgameTowerStatus = 1
	redWs.Write("endgame", endgameData)
	endgameData.TeamPosition = 3
	endgameData.EndgameTowerStatus = 3
	redWs.Write("endgame", endgameData)
	endgameData.TeamPosition = 2
	endgameData.EndgameTowerStatus = 0
	redWs.Write("endgame", endgameData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerLevel2, game.TowerNone, game.TowerLevel3},
		web.arena.RedRealtimeScore.CurrentScore.EndgameTowerStatuses,
	)
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerNone, game.TowerLevel3, game.TowerLevel1},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameTowerStatuses,
	)

	// Add a couple of fouls.
	foulData := struct {
		Alliance string
		IsMajor  bool
	}{Alliance: "red", IsMajor: true}
	redWs.Write("addFoul", foulData)
	foulData = struct {
		Alliance string
		IsMajor  bool
	}{Alliance: "blue", IsMajor: false}
	blueWs.Write("addFoul", foulData)
	for i := 0; i < 2; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].IsMajor)
	assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, false, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].IsMajor)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)
	endgameData.TeamPosition = 1
	endgameData.EndgameTowerStatus = 4
	blueWs.Write("endgame", endgameData)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
}
