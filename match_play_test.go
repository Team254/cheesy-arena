// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"testing"
	"time"
)

func TestMatchPlay(t *testing.T) {
	setupTest(t)

	match1 := model.Match{Type: "practice", DisplayName: "1", Status: "complete", Winner: "R"}
	match2 := model.Match{Type: "practice", DisplayName: "2"}
	match3 := model.Match{Type: "qualification", DisplayName: "1", Status: "complete", Winner: "B"}
	match4 := model.Match{Type: "elimination", DisplayName: "SF1-1", Status: "complete", Winner: "T"}
	match5 := model.Match{Type: "elimination", DisplayName: "SF1-2"}
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
	setupTest(t)

	db.CreateTeam(&model.Team{Id: 101})
	db.CreateTeam(&model.Team{Id: 102})
	db.CreateTeam(&model.Team{Id: 103})
	db.CreateTeam(&model.Team{Id: 104})
	db.CreateTeam(&model.Team{Id: 105})
	db.CreateTeam(&model.Team{Id: 106})
	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: "complete", Winner: "R", Red1: 101,
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

func TestMatchPlayShowResult(t *testing.T) {
	setupTest(t)

	recorder := getHttpResponse("/match_play/1/show_result")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
	match := model.Match{Type: "qualification", DisplayName: "1", Status: "complete"}
	db.CreateMatch(&match)
	recorder = getHttpResponse(fmt.Sprintf("/match_play/%d/show_result", match.Id))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No result found")
	db.CreateMatchResult(&model.MatchResult{MatchId: match.Id})
	recorder = getHttpResponse(fmt.Sprintf("/match_play/%d/show_result", match.Id))
	assert.Equal(t, 302, recorder.Code)
	assert.Equal(t, match.Id, mainArena.savedMatch.Id)
	assert.Equal(t, match.Id, mainArena.savedMatchResult.MatchId)
}

func TestMatchPlayErrors(t *testing.T) {
	setupTest(t)

	// Load an invalid match.
	recorder := getHttpResponse("/match_play/1114/load")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
}

func TestCommitMatch(t *testing.T) {
	setupTest(t)

	// Committing test match should do nothing.
	match := &model.Match{Id: 0, Type: "test", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	err := CommitMatchScore(match, &model.MatchResult{MatchId: match.Id}, false)
	assert.Nil(t, err)
	matchResult, err := db.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)

	// Committing the same match more than once should create a second match result record.
	match.Id = 1
	match.Type = "qualification"
	db.CreateMatch(match)
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueScore = &game.Score{AutoMobility: 2}
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "B", match.Winner)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.RedScore = &game.Score{AutoMobility: 1}
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "R", match.Winner)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	assert.Equal(t, 3, matchResult.PlayNumber)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "T", match.Winner)

	// Verify TBA and STEMtv publishing by checking the log for the expected failure messages.
	tbaClient.BaseUrl = "fakeUrl"
	stemTvClient.BaseUrl = "fakeUrl"
	eventSettings.TbaPublishingEnabled = true
	eventSettings.StemTvPublishingEnabled = true
	var writer bytes.Buffer
	log.SetOutput(&writer)
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10) // Allow some time for the asynchronous publishing to happen.
	assert.Contains(t, writer.String(), "Failed to publish matches")
	assert.Contains(t, writer.String(), "Failed to publish rankings")
	assert.Contains(t, writer.String(), "Failed to publish match video split to STEMtv")
}

func TestCommitEliminationTie(t *testing.T) {
	setupTest(t)

	match := &model.Match{Id: 0, Type: "qualification", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
	db.CreateMatch(match)
	matchResult := &model.MatchResult{MatchId: match.Id, RedScore: &game.Score{FuelHigh: 15, Fouls: []game.Foul{{}}},
		BlueScore: &game.Score{}}
	err := CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "T", match.Winner)
	match.Type = "elimination"
	db.SaveMatch(match)
	CommitMatchScore(match, matchResult, false)
	match, _ = db.GetMatchById(1)
	assert.Equal(t, "T", match.Winner) // No elimination tiebreakers.
}

func TestCommitCards(t *testing.T) {
	setupTest(t)

	// Check that a yellow card sticks with a team.
	team := &model.Team{Id: 5}
	db.CreateTeam(team)
	match := &model.Match{Id: 0, Type: "qualification", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
	db.CreateMatch(match)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueCards = map[string]string{"5": "yellow"}
	err := CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	team, _ = db.GetTeamById(5)
	assert.True(t, team.YellowCard)

	// Check that editing a match result removes a yellow card from a team.
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	team, _ = db.GetTeamById(5)
	assert.False(t, team.YellowCard)

	// Check that a red card causes a yellow card to stick with a team.
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueCards = map[string]string{"5": "red"}
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	team, _ = db.GetTeamById(5)
	assert.True(t, team.YellowCard)

	// Check that a red card in eliminations zeroes out the score.
	createTestAlliances(db, 2)
	match.Type = "elimination"
	db.SaveMatch(match)
	matchResult = model.BuildTestMatchResult(match.Id, 10)
	matchResult.MatchType = match.Type
	matchResult.RedCards = map[string]string{"1": "red"}
	err = CommitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)
	assert.Equal(t, 533, matchResult.BlueScoreSummary().Score)
}

func TestMatchPlayWebsocketCommands(t *testing.T) {
	setupTest(t)

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "status")
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "setAudienceDisplay")
	readWebsocketType(t, ws, "scoringStatus")
	readWebsocketType(t, ws, "setAllianceStationDisplay")

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
	assert.Equal(t, startMatch, mainArena.MatchState)
	ws.Write("commitResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("discardResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("abortMatch", nil)
	readWebsocketType(t, ws, "status")
	readWebsocketType(t, ws, "setAudienceDisplay")
	assert.Equal(t, postMatch, mainArena.MatchState)
	mainArena.redRealtimeScore.CurrentScore.AutoMobility = 1
	mainArena.blueRealtimeScore.CurrentScore.AutoFuelLow = 2
	ws.Write("commitResults", nil)
	readWebsocketMultiple(t, ws, 3) // reload, realtimeScore, setAllianceStationDisplay
	assert.Equal(t, 1, mainArena.savedMatchResult.RedScore.AutoMobility)
	assert.Equal(t, 2, mainArena.savedMatchResult.BlueScore.AutoFuelLow)
	assert.Equal(t, preMatch, mainArena.MatchState)
	ws.Write("discardResults", nil)
	readWebsocketMultiple(t, ws, 3) // reload, realtimeScore, setAllianceStationDisplay
	assert.Equal(t, preMatch, mainArena.MatchState)

	// Test changing the displays.
	ws.Write("setAudienceDisplay", "logo")
	readWebsocketType(t, ws, "setAudienceDisplay")
	assert.Equal(t, "logo", mainArena.audienceDisplayScreen)
	ws.Write("setAllianceStationDisplay", "logo")
	readWebsocketType(t, ws, "setAllianceStationDisplay")
	assert.Equal(t, "logo", mainArena.allianceStationDisplayScreen)
}

func TestMatchPlayWebsocketNotifications(t *testing.T) {
	setupTest(t)

	db.CreateTeam(&model.Team{Id: 254})

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "status")
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "setAudienceDisplay")
	readWebsocketType(t, ws, "scoringStatus")

	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].Bypass = true
	mainArena.StartMatch()
	mainArena.Update()
	messages := readWebsocketMultiple(t, ws, 4)
	statusReceived, matchTime := getStatusMatchTime(t, messages)
	assert.Equal(t, true, statusReceived)
	assert.Equal(t, 2, matchTime.MatchState)
	assert.Equal(t, 0, matchTime.MatchTimeSec)
	_, ok := messages["setAudienceDisplay"]
	assert.True(t, ok)
	_, ok = messages["setAllianceStationDisplay"]
	assert.True(t, ok)
	mainArena.scoringStatusNotifier.Notify(nil)
	readWebsocketType(t, ws, "scoringStatus")

	// Should get a tick notification when an integer second threshold is crossed.
	mainArena.matchStartTime = time.Now().Add(-time.Second + 10*time.Millisecond) // Not crossed yet
	mainArena.Update()
	mainArena.matchStartTime = time.Now().Add(-time.Second - 10*time.Millisecond) // Crossed
	mainArena.Update()
	mainArena.matchStartTime = time.Now().Add(-2*time.Second + 10*time.Millisecond) // Not crossed yet
	mainArena.Update()
	mainArena.matchStartTime = time.Now().Add(-2*time.Second - 10*time.Millisecond) // Crossed
	mainArena.Update()
	err = mapstructure.Decode(readWebsocketType(t, ws, "matchTime"), &matchTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchTime.MatchState)
	assert.Equal(t, 1, matchTime.MatchTimeSec)
	err = mapstructure.Decode(readWebsocketType(t, ws, "matchTime"), &matchTime)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchTime.MatchState)
	assert.Equal(t, 2, matchTime.MatchTimeSec)

	// Check across a match state boundary.
	mainArena.matchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.AutoDurationSec) * time.Second)
	mainArena.Update()
	statusReceived, matchTime = readWebsocketStatusMatchTime(t, ws)
	assert.Equal(t, true, statusReceived)
	assert.Equal(t, 3, matchTime.MatchState)
	assert.Equal(t, game.MatchTiming.AutoDurationSec, matchTime.MatchTimeSec)
}

// Handles the status and matchTime messages arriving in either order.
func readWebsocketStatusMatchTime(t *testing.T, ws *Websocket) (bool, MatchTimeMessage) {
	return getStatusMatchTime(t, readWebsocketMultiple(t, ws, 2))
}

func getStatusMatchTime(t *testing.T, messages map[string]interface{}) (bool, MatchTimeMessage) {
	_, statusReceived := messages["status"]
	message, ok := messages["matchTime"]
	var matchTime MatchTimeMessage
	if assert.True(t, ok) {
		err := mapstructure.Decode(message, &matchTime)
		assert.Nil(t, err)
	}
	return statusReceived, matchTime
}
