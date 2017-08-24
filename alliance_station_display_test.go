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
	setupTest(t)

	recorder := getHttpResponse("/displays/alliance_station")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Alliance Station Display - Untitled Event - Cheesy Arena")
}

func TestAllianceStationDisplayWebsocket(t *testing.T) {
	setupTest(t)

	server, wsUrl := startTestServer()
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
	mainArena.allianceStationDisplayScreen = "logo"
	mainArena.allianceStationDisplayNotifier.Notify(nil)
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "setAllianceStationDisplay")

	// Inform the server what display ID this is.
	assert.Equal(t, "", mainArena.allianceStationDisplays["1"])
	ws.Write("setAllianceStation", "R3")
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	assert.Equal(t, "R3", mainArena.allianceStationDisplays["1"])

	// Run through a match cycle.
	mainArena.matchLoadTeamsNotifier.Notify(nil)
	readWebsocketType(t, ws, "setMatch")
	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].Bypass = true
	mainArena.StartMatch()
	mainArena.Update()
	messages := readWebsocketMultiple(t, ws, 2)
	_, ok := messages["status"]
	assert.True(t, ok)
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	mainArena.realtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
}
