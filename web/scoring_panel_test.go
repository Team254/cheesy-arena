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
	assert.Equal(t, [3]bool{false, false, false}, web.arena.RedRealtimeScore.CurrentScore.MobilityStatuses)
	scoringData := struct {
		TeamPosition int
		GridRow      int
		GridNode     int
		NodeState    game.NodeState
	}{}
	web.arena.MatchState = field.AutoPeriod
	scoringData.TeamPosition = 1
	redWs.Write("mobilityStatus", scoringData)
	scoringData.TeamPosition = 3
	redWs.Write("mobilityStatus", scoringData)
	scoringData.TeamPosition = 2
	redWs.Write("autoDockStatus", scoringData)
	redWs.Write("autoChargeStationLevel", scoringData)
	scoringData.GridRow = 2
	scoringData.GridNode = 7
	scoringData.NodeState = game.ConeThenCube
	redWs.Write("gridNode", scoringData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [3]bool{true, false, true}, web.arena.RedRealtimeScore.CurrentScore.MobilityStatuses)
	assert.Equal(t, [3]bool{false, true, false}, web.arena.RedRealtimeScore.CurrentScore.AutoDockStatuses)
	assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.AutoChargeStationLevel)
	assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.Grid.AutoScoring[2][7])
	assert.Equal(t, game.ConeThenCube, web.arena.RedRealtimeScore.CurrentScore.Grid.Nodes[2][7])

	// Send some teleoperated period scoring commands.
	web.arena.MatchState = field.TeleopPeriod
	scoringData.GridRow = 0
	scoringData.GridNode = 1
	scoringData.NodeState = game.TwoCubes
	blueWs.Write("gridNode", scoringData)
	scoringData.GridRow = 2
	blueWs.Write("gridAutoScoring", scoringData)
	scoringData.TeamPosition = 2
	blueWs.Write("endgameStatus", scoringData)
	scoringData.TeamPosition = 3
	blueWs.Write("endgameStatus", scoringData)
	blueWs.Write("endgameStatus", scoringData)
	blueWs.Write("endgameChargeStationLevel", scoringData)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, false, web.arena.BlueRealtimeScore.CurrentScore.Grid.AutoScoring[0][1])
	assert.Equal(t, game.TwoCubes, web.arena.BlueRealtimeScore.CurrentScore.Grid.Nodes[0][1])
	assert.Equal(t, true, web.arena.BlueRealtimeScore.CurrentScore.Grid.AutoScoring[2][1])
	assert.Equal(
		t,
		[3]game.EndgameStatus{game.EndgameNone, game.EndgameParked, game.EndgameDocked},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameStatuses,
	)
	assert.Equal(t, true, web.arena.BlueRealtimeScore.CurrentScore.EndgameChargeStationLevel)

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
