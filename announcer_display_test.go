// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestAnnouncerDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/announcer")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Announcer Display - Untitled Event - Cheesy Arena")
}

func TestAnnouncerDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/announcer/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "setMatch")
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")

	web.arena.MatchLoadTeamsNotifier.Notify(nil)
	readWebsocketType(t, ws, "setMatch")
	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	web.arena.StartMatch()
	web.arena.Update()
	messages := readWebsocketMultiple(t, ws, 2)
	_, ok := messages["setAudienceDisplay"]
	assert.True(t, ok)
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	web.arena.RealtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
	web.arena.ScorePostedNotifier.Notify(nil)
	readWebsocketType(t, ws, "setFinalScore")

	// Test triggering the final score screen.
	ws.Write("setAudienceDisplay", "score")
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	assert.Equal(t, "score", web.arena.AudienceDisplayScreen)
}
