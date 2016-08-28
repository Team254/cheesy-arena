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
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	recorder := getHttpResponse("/displays/announcer")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Announcer Display - Untitled Event - Cheesy Arena")
}

func TestAnnouncerDisplayWebsocket(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	server, wsUrl := startTestServer()
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
	_, ok := messages["setAudienceDisplay"]
	assert.True(t, ok)
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	mainArena.realtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
	mainArena.scorePostedNotifier.Notify(nil)
	readWebsocketType(t, ws, "setFinalScore")

	// Test triggering the final score screen.
	ws.Write("setAudienceDisplay", "score")
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	assert.Equal(t, "score", mainArena.audienceDisplayScreen)
}
