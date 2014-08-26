// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAudienceDisplay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	recorder := getHttpResponse("/displays/audience")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Audience Display - Untitled Event - Cheesy Arena")
}

func TestAudienceDisplayWebsocket(t *testing.T) {
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
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/audience/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn}

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
}

func TestPitDisplay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	recorder := getHttpResponse("/displays/pit")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Pit Display - Untitled Event - Cheesy Arena")
}

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
	mainArena.Setup()

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/announcer/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn}

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

	// Load another match to reset the results.
	mainArena.ResetMatch()
	mainArena.LoadTestMatch()
	readWebsocketType(t, redWs, "score")
	readWebsocketType(t, blueWs, "score")
	assert.Equal(t, *NewRealtimeScore(), *mainArena.redRealtimeScore)
	assert.Equal(t, *NewRealtimeScore(), *mainArena.blueRealtimeScore)
}

func TestRefereeDisplay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	recorder := getHttpResponse("/displays/referee")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Referee Display - Untitled Event - Cheesy Arena")
}

func TestRefereeDisplayWebsocket(t *testing.T) {
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
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/referee/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn}

	// Test foul addition.
	foulData := struct {
		Alliance       string
		TeamId         int
		Rule           string
		TimeInMatchSec float64
		IsTechnical    bool
	}{"red", 256, "G22", 0, false}
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
	if assert.Equal(t, 2, len(mainArena.redRealtimeScore.Fouls)) {
		assert.Equal(t, 256, mainArena.redRealtimeScore.Fouls[0].TeamId)
		assert.Equal(t, "G22", mainArena.redRealtimeScore.Fouls[0].Rule)
		assert.Equal(t, 0, mainArena.redRealtimeScore.Fouls[0].TimeInMatchSec)
		assert.Equal(t, false, mainArena.redRealtimeScore.Fouls[0].IsTechnical)
		assert.Equal(t, 359, mainArena.redRealtimeScore.Fouls[1].TeamId)
		assert.Equal(t, "G22", mainArena.redRealtimeScore.Fouls[1].Rule)
		assert.Equal(t, true, mainArena.redRealtimeScore.Fouls[1].IsTechnical)
	}
	if assert.Equal(t, 1, len(mainArena.blueRealtimeScore.Fouls)) {
		assert.Equal(t, 1680, mainArena.blueRealtimeScore.Fouls[0].TeamId)
		assert.Equal(t, "G22", mainArena.blueRealtimeScore.Fouls[0].Rule)
		assert.Equal(t, 0, mainArena.blueRealtimeScore.Fouls[0].TimeInMatchSec)
		assert.Equal(t, true, mainArena.blueRealtimeScore.Fouls[0].IsTechnical)
	}
	assert.False(t, mainArena.redRealtimeScore.FoulsCommitted)
	assert.False(t, mainArena.blueRealtimeScore.FoulsCommitted)

	// Test foul deletion.
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 0, len(mainArena.blueRealtimeScore.Fouls))
	foulData.Alliance = "red"
	foulData.TeamId = 359
	foulData.TimeInMatchSec = 29 // Make it not match.
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 2, len(mainArena.redRealtimeScore.Fouls))
	foulData.TimeInMatchSec = 0
	ws.Write("deleteFoul", foulData)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, 1, len(mainArena.redRealtimeScore.Fouls))

	// Test match committing.
	mainArena.MatchState = POST_MATCH
	ws.Write("commitMatch", foulData)
	readWebsocketType(t, ws, "reload")
	assert.True(t, mainArena.redRealtimeScore.FoulsCommitted)
	assert.True(t, mainArena.blueRealtimeScore.FoulsCommitted)

	// Should refresh the page when the next match is loaded.
	mainArena.matchLoadTeamsNotifier.Notify(nil)
	readWebsocketType(t, ws, "reload")
}
