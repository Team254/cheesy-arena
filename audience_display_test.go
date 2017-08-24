// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAudienceDisplay(t *testing.T) {
	setupTest(t)

	recorder := getHttpResponse("/displays/audience")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Audience Display - Untitled Event - Cheesy Arena")
}

func TestAudienceDisplayWebsocket(t *testing.T) {
	setupTest(t)

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/audience/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "setAudienceDisplay")
	readWebsocketType(t, ws, "setMatch")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "setFinalScore")
	readWebsocketType(t, ws, "allianceSelection")

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
	messages := readWebsocketMultiple(t, ws, 3)
	screen, ok := messages["setAudienceDisplay"]
	if assert.True(t, ok) {
		assert.Equal(t, "match", screen)
	}
	sound, ok := messages["playSound"]
	if assert.True(t, ok) {
		assert.Equal(t, "match-start", sound)
	}
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	mainArena.realtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
	mainArena.scorePostedNotifier.Notify(nil)
	readWebsocketType(t, ws, "setFinalScore")

	// Test other overlays.
	mainArena.allianceSelectionNotifier.Notify(nil)
	readWebsocketType(t, ws, "allianceSelection")
	mainArena.lowerThirdNotifier.Notify(nil)
	readWebsocketType(t, ws, "lowerThird")
}
