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
	redWs.Write("defenseCrossed", "2")
	blueWs.Write("autoDefenseReached", nil)
	redWs.Write("highGoal", nil)
	redWs.Write("highGoal", nil)
	redWs.Write("lowGoal", nil)
	redWs.Write("defenseCrossed", "5")
	blueWs.Write("defenseCrossed", "1")
	redWs.Write("undoHighGoal", nil)
	redWs.Write("commit", nil)
	blueWs.Write("autoDefenseReached", nil)
	blueWs.Write("commit", nil)
	redWs.Write("uncommitAuto", nil)
	redWs.Write("autoDefenseReached", nil)
	redWs.Write("defenseCrossed", "2")
	redWs.Write("commit", nil)
	for i := 0; i < 11; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 4; i++ {
		readWebsocketType(t, blueWs, "score")
	}

	assert.Equal(t, [5]int{0, 2, 0, 0, 1}, mainArena.redRealtimeScore.CurrentScore.AutoDefensesCrossed)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoDefensesReached)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoHighGoals)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoLowGoals)
	assert.Equal(t, [5]int{1, 0, 0, 0, 0}, mainArena.blueRealtimeScore.CurrentScore.AutoDefensesCrossed)
	assert.Equal(t, 2, mainArena.blueRealtimeScore.CurrentScore.AutoDefensesReached)

	redWs.Write("defenseCrossed", "2")
	blueWs.Write("autoDefenseReached", nil)
	redWs.Write("highGoal", nil)
	redWs.Write("highGoal", nil)
	redWs.Write("lowGoal", nil)
	redWs.Write("defenseCrossed", "5")
	blueWs.Write("defenseCrossed", "3")
	blueWs.Write("challenge", nil)
	blueWs.Write("scale", nil)
	blueWs.Write("undoChallenge", nil)
	redWs.Write("challenge", nil)
	redWs.Write("defenseCrossed", "3")
	redWs.Write("undoHighGoal", nil)
	for i := 0; i < 8; i++ {
		readWebsocketType(t, redWs, "score")
	}
	for i := 0; i < 5; i++ {
		readWebsocketType(t, blueWs, "score")
	}

	// Make sure auto scores haven't changed in teleop.
	assert.Equal(t, [5]int{0, 2, 0, 0, 1}, mainArena.redRealtimeScore.CurrentScore.AutoDefensesCrossed)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoDefensesReached)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoHighGoals)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.AutoLowGoals)
	assert.Equal(t, [5]int{1, 0, 0, 0, 0}, mainArena.blueRealtimeScore.CurrentScore.AutoDefensesCrossed)
	assert.Equal(t, 2, mainArena.blueRealtimeScore.CurrentScore.AutoDefensesReached)

	assert.Equal(t, [5]int{0, 0, 1, 0, 1}, mainArena.redRealtimeScore.CurrentScore.DefensesCrossed)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.HighGoals)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.LowGoals)
	assert.Equal(t, 1, mainArena.redRealtimeScore.CurrentScore.Challenges)
	assert.Equal(t, [5]int{0, 0, 1, 0, 0}, mainArena.blueRealtimeScore.CurrentScore.DefensesCrossed)
	assert.Equal(t, 0, mainArena.blueRealtimeScore.CurrentScore.Challenges)
	assert.Equal(t, 1, mainArena.blueRealtimeScore.CurrentScore.Scales)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	mainArena.MatchState = POST_MATCH
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
