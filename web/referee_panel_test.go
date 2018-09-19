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

	// Test foul addition.
	foulData := struct {
		Alliance       string
		TeamId         int
		Rule           string
		IsTechnical    bool
		TimeInMatchSec float64
	}{"red", 256, "G22", false, 0}
	ws.Write("addFoul", foulData)
	foulData.TeamId = 359
	foulData.IsTechnical = true
	ws.Write("addFoul", foulData)
	foulData.Alliance = "blue"
	foulData.TeamId = 1680
	ws.Write("addFoul", foulData)
	readWebsocketType(t, ws, "reload")
	readWebsocketType(t, ws, "reload")
	readWebsocketType(t, ws, "reload")
	if assert.Equal(t, 2, len(web.arena.RedRealtimeScore.CurrentScore.Fouls)) {
		assert.Equal(t, 256, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].TeamId)
		assert.Equal(t, "G22", web.arena.RedRealtimeScore.CurrentScore.Fouls[0].RuleNumber)
		assert.Equal(t, false, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].IsTechnical)
		assert.Equal(t, 0.0, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].TimeInMatchSec)
		assert.Equal(t, 359, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].TeamId)
		assert.Equal(t, "G22", web.arena.RedRealtimeScore.CurrentScore.Fouls[1].RuleNumber)
		assert.Equal(t, true, web.arena.RedRealtimeScore.CurrentScore.Fouls[1].IsTechnical)
	}
	if assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls)) {
		assert.Equal(t, 1680, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].TeamId)
		assert.Equal(t, "G22", web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].RuleNumber)
		assert.Equal(t, true, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].IsTechnical)
		assert.Equal(t, 0.0, web.arena.BlueRealtimeScore.CurrentScore.Fouls[0].TimeInMatchSec)
	}
	assert.False(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.False(t, web.arena.BlueRealtimeScore.FoulsCommitted)

	// Test foul deletion.
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	foulData.Alliance = "red"
	foulData.TeamId = 359
	foulData.TimeInMatchSec = 29 // Make it not match.
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 2, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	foulData.TimeInMatchSec = 0
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))

	// Test card setting.
	cardData := struct {
		Alliance string
		TeamId   int
		Card     string
	}{"red", 256, "yellow"}
	ws.Write("card", cardData)
	cardData.Alliance = "blue"
	cardData.TeamId = 1680
	cardData.Card = "red"
	ws.Write("card", cardData)
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
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, "fieldReset", web.arena.AllianceStationDisplayMode)
	assert.True(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.True(t, web.arena.BlueRealtimeScore.FoulsCommitted)

	// Should refresh the page when the next match is loaded.
	web.arena.MatchLoadNotifier.Notify()
	readWebsocketType(t, ws, "matchLoad")
}
