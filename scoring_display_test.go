// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
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
	redWs := &Websocket{redConn}
	blueConn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := &Websocket{blueConn}

	// Should receive a score update right after connection.
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, blueWs, "score")
	readWebsocketType(t, blueWs, "matchTime")

	// Send a match worth of scoring commands in.
	redWs.Write("robotSet", nil)
	blueWs.Write("containerSet", nil)
	redWs.Write("stackedToteSet", nil)
	redWs.Write("robotSet", nil)
	redWs.Write("toteSet", nil)
	blueWs.Write("stackedToteSet", nil)
	redWs.Write("commit", nil)
	blueWs.Write("commit", nil)
	redWs.Write("uncommitAuto", nil)
	redWs.Write("robotSet", nil)
	redWs.Write("commit", nil)
	for i := 0; i < 8; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 3; i++ {
		readWebsocketType(t, blueWs, "score")
	}
	assert.True(t, mainArena.redRealtimeScore.CurrentScore.AutoRobotSet)
	assert.False(t, mainArena.redRealtimeScore.CurrentScore.AutoContainerSet)
	assert.True(t, mainArena.redRealtimeScore.CurrentScore.AutoToteSet)
	assert.False(t, mainArena.redRealtimeScore.CurrentScore.AutoStackedToteSet)
	assert.False(t, mainArena.blueRealtimeScore.CurrentScore.AutoRobotSet)
	assert.True(t, mainArena.blueRealtimeScore.CurrentScore.AutoContainerSet)
	assert.False(t, mainArena.blueRealtimeScore.CurrentScore.AutoToteSet)
	assert.True(t, mainArena.blueRealtimeScore.CurrentScore.AutoStackedToteSet)

	stacks := []Stack{Stack{6, true, true}, Stack{1, false, false}, Stack{2, true, false}, Stack{}}
	blueWs.Write("commit", stacks)
	redWs.Write("commit", stacks)
	stacks[0].Litter = false
	blueWs.Write("commit", stacks)
	redWs.Write("toteSet", nil)
	blueWs.Write("stackedToteSet", nil)
	for i := 0; i < 2; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 3; i++ {
		readWebsocketType(t, blueWs, "score")
	}
	assert.Equal(t, stacks, mainArena.blueRealtimeScore.CurrentScore.Stacks)
	stacks[0].Litter = true
	assert.Equal(t, stacks, mainArena.redRealtimeScore.CurrentScore.Stacks)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	mainArena.MatchState = POST_MATCH
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "dialog") // Should be an error message about co-op not matching.
	readWebsocketType(t, blueWs, "dialog")
	redWs.Write("stackedToteSet", nil)
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")

	// Load another match to reset the results.
	mainArena.ResetMatch()
	mainArena.LoadTestMatch()
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")
	assert.Equal(t, *NewRealtimeScore(), *mainArena.redRealtimeScore)
	assert.Equal(t, *NewRealtimeScore(), *mainArena.blueRealtimeScore)
}
