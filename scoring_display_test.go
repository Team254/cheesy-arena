// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestScoringDisplay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	recorder := getHttpResponse("/displays/scoring/invalidalliance")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid alliance")
	recorder = getHttpResponse("/displays/scoring/red")
	assert.Equal(t, 200, recorder.Code)
	recorder = getHttpResponse("/displays/scoring/blue")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring - Untitled Event - Cheesy Arena")
}

func TestScoringDisplayWebsocket(t *testing.T) {
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
	_, _, err = websocket.DefaultDialer.Dial(wsUrl+"/displays/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/scoring/red/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := &Websocket{redConn, new(sync.Mutex)}
	blueConn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := &Websocket{blueConn, new(sync.Mutex)}

	// Should receive a score update right after connection.
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, blueWs, "score")
	readWebsocketType(t, blueWs, "matchTime")

	// Send a match worth of scoring commands in.
	redWs.Write("mobility", nil)
	blueWs.Write("mobility", nil)
	blueWs.Write("mobility", nil)
	blueWs.Write("mobility", nil)
	blueWs.Write("mobility", nil)
	blueWs.Write("undoMobility", nil)
	redWs.Write("commit", nil)
	blueWs.Write("commit", nil)
	redWs.Write("uncommitAuto", nil)
	redWs.Write("commit", nil)
	for i := 0; i < 4; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 6; i++ {
		readWebsocketType(t, blueWs, "score")
	}

	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoMobility)
	assert.Equal(t, 2, mainArena.blueRealtimeScore.CurrentScore.AutoMobility)

	redWs.Write("mobility", nil)
	for i := 0; i < 1; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 0; i++ {
		readWebsocketType(t, blueWs, "score")
	}

	// Make sure auto scores haven't changed in teleop.
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoMobility)
	assert.Equal(t, 2, mainArena.blueRealtimeScore.CurrentScore.AutoMobility)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	mainArena.MatchState = postMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")

	// Load another match to reset the results.
	mainArena.ResetMatch()
	mainArena.LoadTestMatch()
	readWebsocketType(t, redWs, "reload")
	readWebsocketType(t, blueWs, "reload")
	assert.Equal(t, NewRealtimeScore(), mainArena.redRealtimeScore)
	assert.Equal(t, NewRealtimeScore(), mainArena.blueRealtimeScore)
}
