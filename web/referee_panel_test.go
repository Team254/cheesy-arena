// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRefereePanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/referee")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Referee Panel - Untitled Event - Cheesy Arena")
}

func TestRefereePanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/referee/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "scoringStatus")

	// Test foul addition.
	addFoulData := struct {
		Alliance    string
		IsTechnical bool
	}{"red", true}
	ws.Write("addFoul", addFoulData)
	addFoulData.IsTechnical = false
	ws.Write("addFoul", addFoulData)
	addFoulData.Alliance = "blue"
	ws.Write("addFoul", addFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "realtimeScore")
	if assert.Equal(t, 2, len(web.arena.RedRealtimeScore.CurrentScore.Fouls)) {
		assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].IsTechnical)
		assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].TeamId)
		assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].RuleId)
		assert.Equal(t, false, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].IsTechnical)
		assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].TeamId)
		assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].RuleId)
	}
	if assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls)) {
		assert.Equal(t, false, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].IsTechnical)
		assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].TeamId)
		assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].RuleId)
	}
	assert.False(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.False(t, web.arena.BlueRealtimeScore.FoulsCommitted)

	// Test foul mutation.
	modifyFoulData := struct {
		Alliance string
		Index    int
		TeamId   int
		RuleId   int
	}{}
	modifyFoulData.Alliance = "red"
	modifyFoulData.Index = 1
	ws.Write("toggleFoulType", modifyFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].IsTechnical)
	modifyFoulData.Index = 0
	modifyFoulData.TeamId = 256
	ws.Write("updateFoulTeam", modifyFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	assert.Equal(t, 256, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].TeamId)
	modifyFoulData.Alliance = "blue"
	modifyFoulData.RuleId = 3
	ws.Write("updateFoulRule", modifyFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	assert.Equal(t, 3, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].RuleId)

	// Test foul deletion.
	modifyFoulData.Alliance = "blue"
	modifyFoulData.Index = 0
	ws.Write("deleteFoul", modifyFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	modifyFoulData.Alliance = "red"
	modifyFoulData.Index = -1 // Invalid index.
	ws.Write("deleteFoul", modifyFoulData)
	assert.Equal(t, 2, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	modifyFoulData.Alliance = "red"
	modifyFoulData.Index = 2 // Invalid index.
	ws.Write("deleteFoul", modifyFoulData)
	assert.Equal(t, 2, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	modifyFoulData.Index = 1
	ws.Write("deleteFoul", modifyFoulData)
	readWebsocketType(t, ws, "realtimeScore")
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))

	// Test card setting.
	cardData := struct {
		Alliance string
		TeamId   int
		Card     string
	}{"red", 256, "yellow"}
	ws.Write("card", cardData)
	readWebsocketType(t, ws, "realtimeScore")
	cardData.Alliance = "blue"
	cardData.TeamId = 1680
	cardData.Card = "red"
	ws.Write("card", cardData)
	readWebsocketType(t, ws, "realtimeScore")
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	if assert.Equal(t, 1, len(web.arena.RedRealtimeScore.Cards)) {
		assert.Equal(t, "yellow", web.arena.RedRealtimeScore.Cards["256"])
	}
	if assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.Cards)) {
		assert.Equal(t, "red", web.arena.BlueRealtimeScore.Cards["1680"])
	}

	// Test field reset and match committing.
	web.arena.MatchState = field.PostMatch
	ws.Write("signalReset", nil)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, "fieldReset", web.arena.AllianceStationDisplayMode)
	assert.False(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.False(t, web.arena.BlueRealtimeScore.FoulsCommitted)
	web.arena.AllianceStationDisplayMode = "logo"
	ws.Write("commitMatch", nil)
	readWebsocketType(t, ws, "scoringStatus")
	assert.Equal(t, "fieldReset", web.arena.AllianceStationDisplayMode)
	assert.True(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.True(t, web.arena.BlueRealtimeScore.FoulsCommitted)

	// Should refresh the page when the next match is loaded.
	web.arena.MatchLoadNotifier.Notify()
	readWebsocketType(t, ws, "matchLoad")
}
