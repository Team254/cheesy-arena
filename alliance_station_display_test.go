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

func TestAllianceStationDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/alliance_station")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Alliance Station Display - Untitled Event - Cheesy Arena")
}

func TestAllianceStationDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/alliance_station/websocket?displayId=1", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "setAllianceStationDisplay")
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "setMatch")
	readWebsocketType(t, ws, "realtimeScore")

	// Change to a different screen.
	web.arena.AllianceStationDisplayScreen = "logo"
	web.arena.AllianceStationDisplayNotifier.Notify(nil)
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "setAllianceStationDisplay")

	// Inform the server what display ID this is.
	assert.Equal(t, "", web.arena.AllianceStationDisplays["1"])
	ws.Write("setAllianceStation", "R3")
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	assert.Equal(t, "R3", web.arena.AllianceStationDisplays["1"])

	// Run through a match cycle.
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
	_, ok := messages["status"]
	assert.True(t, ok)
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	web.arena.RealtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
}
