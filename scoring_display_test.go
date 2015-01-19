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

	// Should a score update right after connection.
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")

	// Send a match worth of scoring commands in.
	// TODO(pat): Update for 2015.
	/*
		redWs.Write("preload", "3")
		blueWs.Write("preload", "3")
		redWs.Write("mobility", nil)
		blueWs.Write("mobility", nil)
		blueWs.Write("mobility", nil)
		blueWs.Write("mobility", nil)
		blueWs.Write("scoredHighHot", nil)
		blueWs.Write("scoredHigh", nil)
		blueWs.Write("scoredLowHot", nil)
		blueWs.Write("scoredLow", nil)
		blueWs.Write("undo", nil)
		redWs.Write("commit", nil)
		blueWs.Write("commit", nil)
		redWs.Write("deadBall", nil)
		redWs.Write("commit", nil)
		redWs.Write("scoredLow", nil)
		redWs.Write("commit", nil)
		redWs.Write("scoredHigh", nil)
		redWs.Write("commit", nil)
		redWs.Write("assist", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("scoredLow", nil)
		blueWs.Write("scoredHigh", nil)
		blueWs.Write("commit", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("assist", nil)
		blueWs.Write("truss", nil)
		blueWs.Write("catch", nil)
		blueWs.Write("undo", nil)
		blueWs.Write("scoredLow", nil)
		blueWs.Write("commit", nil)
		mainArena.MatchState = POST_MATCH
		redWs.Write("commitMatch", nil)
		for i := 0; i < 11; i++ {
			readWebsocketType(t, redWs, "score")
		}
		for i := 0; i < 24; i++ {
			readWebsocketType(t, blueWs, "score")
		}

		assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoMobilityBonuses)
		assert.Equal(t, 0, mainArena.redRealtimeScore.CurrentScore.AutoHighHot)
		assert.Equal(t, 0, mainArena.redRealtimeScore.CurrentScore.AutoHigh)
		assert.Equal(t, 0, mainArena.redRealtimeScore.CurrentScore.AutoLowHot)
		assert.Equal(t, 0, mainArena.redRealtimeScore.CurrentScore.AutoLow)
		assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoClearHigh)
		assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoClearLow)
		assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoClearDead)
		if assert.Equal(t, 1, len(mainArena.redRealtimeScore.CurrentScore.Cycles)) {
			assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.Cycles[0].Assists)
			assert.False(t, mainArena.redRealtimeScore.CurrentScore.Cycles[0].Truss)
			assert.False(t, mainArena.redRealtimeScore.CurrentScore.Cycles[0].Catch)
			assert.False(t, mainArena.redRealtimeScore.CurrentScore.Cycles[0].ScoredHigh)
			assert.False(t, mainArena.redRealtimeScore.CurrentScore.Cycles[0].ScoredLow)
			assert.False(t, mainArena.redRealtimeScore.CurrentScore.Cycles[0].DeadBall)
		}
		assert.True(t, mainArena.redRealtimeScore.AutoCommitted)
		assert.True(t, mainArena.redRealtimeScore.TeleopCommitted)

		assert.Equal(t, 3, mainArena.blueRealtimeScore.CurrentScore.AutoMobilityBonuses)
		assert.Equal(t, 1, mainArena.blueRealtimeScore.CurrentScore.AutoHighHot)
		assert.Equal(t, 1, mainArena.blueRealtimeScore.CurrentScore.AutoHigh)
		assert.Equal(t, 1, mainArena.blueRealtimeScore.CurrentScore.AutoLowHot)
		assert.Equal(t, 0, mainArena.blueRealtimeScore.CurrentScore.AutoLow)
		assert.Equal(t, 0, mainArena.blueRealtimeScore.CurrentScore.AutoClearHigh)
		assert.Equal(t, 0, mainArena.blueRealtimeScore.CurrentScore.AutoClearLow)
		assert.Equal(t, 0, mainArena.blueRealtimeScore.CurrentScore.AutoClearDead)
		if assert.Equal(t, 2, len(mainArena.blueRealtimeScore.CurrentScore.Cycles)) {
			assert.Equal(t, 3, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].Assists)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].Truss)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].Catch)
			assert.True(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].ScoredHigh)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].ScoredLow)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[0].DeadBall)
			assert.Equal(t, 2, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].Assists)
			assert.True(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].Truss)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].Catch)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].ScoredHigh)
			assert.True(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].ScoredLow)
			assert.False(t, mainArena.blueRealtimeScore.CurrentScore.Cycles[1].DeadBall)
		}
		assert.True(t, mainArena.blueRealtimeScore.AutoCommitted)
		assert.False(t, mainArena.blueRealtimeScore.TeleopCommitted)
	*/

	// Load another match to reset the results.
	mainArena.ResetMatch()
	mainArena.LoadTestMatch()
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")
	assert.Equal(t, *NewRealtimeScore(), *mainArena.redRealtimeScore)
	assert.Equal(t, *NewRealtimeScore(), *mainArena.blueRealtimeScore)
}
