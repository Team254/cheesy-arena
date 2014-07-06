// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchPlay(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	match1 := Match{Type: "practice", DisplayName: "1", Status: "complete", Winner: "R"}
	match2 := Match{Type: "practice", DisplayName: "2"}
	match3 := Match{Type: "qualification", DisplayName: "1", Status: "complete", Winner: "B"}
	match4 := Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete", Winner: "T"}
	match5 := Match{Type: "elimination", DisplayName: "SF1-2"}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	db.CreateMatch(&match3)
	db.CreateMatch(&match4)
	db.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "P1")
	assert.Contains(t, recorder.Body.String(), "P2")
	assert.Contains(t, recorder.Body.String(), "Q1")
	assert.Contains(t, recorder.Body.String(), "SF1-1")
	assert.Contains(t, recorder.Body.String(), "SF1-2")
}

func TestMatchPlayLoad(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	db.CreateTeam(&Team{Id: 101})
	db.CreateTeam(&Team{Id: 102})
	db.CreateTeam(&Team{Id: 103})
	db.CreateTeam(&Team{Id: 104})
	db.CreateTeam(&Team{Id: 105})
	db.CreateTeam(&Team{Id: 106})
	match := Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 101,
		Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	recorder := getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101")
	assert.NotContains(t, recorder.Body.String(), "102")
	assert.NotContains(t, recorder.Body.String(), "103")
	assert.NotContains(t, recorder.Body.String(), "104")
	assert.NotContains(t, recorder.Body.String(), "105")
	assert.NotContains(t, recorder.Body.String(), "106")

	// Load the match and check for the team numbers again.
	recorder = getHttpResponse(fmt.Sprintf("/match_play/%d/load", match.Id))
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "101")
	assert.Contains(t, recorder.Body.String(), "102")
	assert.Contains(t, recorder.Body.String(), "103")
	assert.Contains(t, recorder.Body.String(), "104")
	assert.Contains(t, recorder.Body.String(), "105")
	assert.Contains(t, recorder.Body.String(), "106")

	// Load a test match.
	recorder = getHttpResponse("/match_play/0/load")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101")
	assert.NotContains(t, recorder.Body.String(), "102")
	assert.NotContains(t, recorder.Body.String(), "103")
	assert.NotContains(t, recorder.Body.String(), "104")
	assert.NotContains(t, recorder.Body.String(), "105")
	assert.NotContains(t, recorder.Body.String(), "106")
}

func TestMatchPlayErrors(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Load an invalid match.
	recorder := getHttpResponse("/match_play/1114/load")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
}

func TestCommitMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	// Committing test match should do nothing.
	match := &Match{Id: 0, Type: "test", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	err = CommitMatchScore(match, &MatchResult{MatchId: match.Id})
	assert.Nil(t, err)
	matchResult, err := db.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)

	// Committing the same match more than once should create a second match result record.
	match.Id = 1
	match.Type = "qualification"
	db.CreateMatch(match)
	matchResult = &MatchResult{MatchId: match.Id, BlueScore: Score{AutoHigh: 1}}
	err = CommitMatchScore(match, matchResult)
	assert.Nil(t, err)
	assert.Equal(t, 1, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "B", match.Winner)
	matchResult = &MatchResult{MatchId: match.Id, RedScore: Score{AutoHigh: 1}}
	err = CommitMatchScore(match, matchResult)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "R", match.Winner)
	matchResult = &MatchResult{MatchId: match.Id}
	err = CommitMatchScore(match, matchResult)
	assert.Nil(t, err)
	assert.Equal(t, 3, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "T", match.Winner)
}

func TestMatchPlayWebsocket(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	db.CreateTeam(&Team{Id: 254})
	mainArena.Setup()

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket?test", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn}

	// Should get a status update right after connection.
	messageType, _, err := ws.Read()
	assert.Nil(t, err)
	assert.Equal(t, "status", messageType)

	// Test that a server-side error is communicated to the client.
	ws.Write("nonexistenttype", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Invalid message type")

	// Test match setup commands.
	ws.Write("substituteTeam", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("substituteTeam", map[string]interface{}{"team": 254, "position": "B5"})
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("substituteTeam", map[string]interface{}{"team": 254, "position": "B1"})
	readWebsocketType(t, ws, "status")
	assert.Equal(t, 254, mainArena.currentMatch.Blue1)
	ws.Write("substituteTeam", map[string]interface{}{"team": 0, "position": "B1"})
	readWebsocketType(t, ws, "status")
	assert.Equal(t, 0, mainArena.currentMatch.Blue1)
	ws.Write("toggleBypass", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Failed to parse")
	ws.Write("toggleBypass", "R4")
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("toggleBypass", "R3")
	readWebsocketType(t, ws, "status")
	assert.Equal(t, true, mainArena.AllianceStations["R3"].Bypass)
	ws.Write("toggleBypass", "R3")
	readWebsocketType(t, ws, "status")
	assert.Equal(t, false, mainArena.AllianceStations["R3"].Bypass)

	// Go through match flow.
	ws.Write("abortMatch", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot abort match")
	ws.Write("startMatch", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot start match")
	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].Bypass = true
	ws.Write("startMatch", nil)
	readWebsocketType(t, ws, "status")
	assert.Equal(t, START_MATCH, mainArena.MatchState)
	ws.Write("commitResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("discardResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("abortMatch", nil)
	readWebsocketType(t, ws, "status")
	assert.Equal(t, POST_MATCH, mainArena.MatchState)
	ws.Write("commitResults", nil)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, PRE_MATCH, mainArena.MatchState)
	ws.Write("discardResults", nil)
	readWebsocketType(t, ws, "reload")
	assert.Equal(t, PRE_MATCH, mainArena.MatchState)
}

func readWebsocketError(t *testing.T, ws *Websocket) string {
	messageType, data, err := ws.Read()
	if assert.Nil(t, err) && assert.Equal(t, "error", messageType) {
		return data.(string)
	}
	return "error"
}

func readWebsocketType(t *testing.T, ws *Websocket, expectedMessageType string) {
	messageType, _, err := ws.Read()
	if assert.Nil(t, err) {
		assert.Equal(t, expectedMessageType, messageType)
	}
}
