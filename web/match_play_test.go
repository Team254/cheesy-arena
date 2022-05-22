// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestMatchPlay(t *testing.T) {
	web := setupTestWeb(t)

	match1 := model.Match{Type: "practice", DisplayName: "1", Status: model.RedWonMatch}
	match2 := model.Match{Type: "practice", DisplayName: "2"}
	match3 := model.Match{Type: "qualification", DisplayName: "1", Status: model.BlueWonMatch}
	match4 := model.Match{Type: "elimination", DisplayName: "SF1-1", Status: model.TieMatch}
	match5 := model.Match{Type: "elimination", DisplayName: "SF1-2"}
	web.arena.Database.CreateMatch(&match1)
	web.arena.Database.CreateMatch(&match2)
	web.arena.Database.CreateMatch(&match3)
	web.arena.Database.CreateMatch(&match4)
	web.arena.Database.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := web.getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "P1")
	assert.Contains(t, recorder.Body.String(), "P2")
	assert.Contains(t, recorder.Body.String(), "Q1")
	assert.Contains(t, recorder.Body.String(), "SF1-1")
	assert.Contains(t, recorder.Body.String(), "SF1-2")
}

func TestMatchPlayLoad(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateTeam(&model.Team{Id: 101})
	web.arena.Database.CreateTeam(&model.Team{Id: 102})
	web.arena.Database.CreateTeam(&model.Team{Id: 103})
	web.arena.Database.CreateTeam(&model.Team{Id: 104})
	web.arena.Database.CreateTeam(&model.Team{Id: 105})
	web.arena.Database.CreateTeam(&model.Team{Id: 106})
	match := model.Match{Type: "elimination", DisplayName: "QF4-3", Status: model.RedWonMatch, Red1: 101,
		Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	web.arena.Database.CreateMatch(&match)
	recorder := web.getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101")
	assert.NotContains(t, recorder.Body.String(), "102")
	assert.NotContains(t, recorder.Body.String(), "103")
	assert.NotContains(t, recorder.Body.String(), "104")
	assert.NotContains(t, recorder.Body.String(), "105")
	assert.NotContains(t, recorder.Body.String(), "106")

	// Load the match and check for the team numbers again.
	recorder = web.getHttpResponse(fmt.Sprintf("/match_play/%d/load", match.Id))
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "101")
	assert.Contains(t, recorder.Body.String(), "102")
	assert.Contains(t, recorder.Body.String(), "103")
	assert.Contains(t, recorder.Body.String(), "104")
	assert.Contains(t, recorder.Body.String(), "105")
	assert.Contains(t, recorder.Body.String(), "106")

	// Load a test match.
	recorder = web.getHttpResponse("/match_play/0/load")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "101")
	assert.NotContains(t, recorder.Body.String(), "102")
	assert.NotContains(t, recorder.Body.String(), "103")
	assert.NotContains(t, recorder.Body.String(), "104")
	assert.NotContains(t, recorder.Body.String(), "105")
	assert.NotContains(t, recorder.Body.String(), "106")
}

func TestMatchPlayShowResult(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/match_play/1/show_result")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
	match := model.Match{Type: "qualification", DisplayName: "1", Status: model.TieMatch}
	web.arena.Database.CreateMatch(&match)
	recorder = web.getHttpResponse(fmt.Sprintf("/match_play/%d/show_result", match.Id))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No result found")
	web.arena.Database.CreateMatchResult(model.BuildTestMatchResult(match.Id, 1))
	recorder = web.getHttpResponse(fmt.Sprintf("/match_play/%d/show_result", match.Id))
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, match.Id, web.arena.SavedMatch.Id)
	assert.Equal(t, match.Id, web.arena.SavedMatchResult.MatchId)
}

func TestMatchPlayErrors(t *testing.T) {
	web := setupTestWeb(t)

	// Load an invalid match.
	recorder := web.getHttpResponse("/match_play/1114/load")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid match")
}

func TestCommitMatch(t *testing.T) {
	web := setupTestWeb(t)

	// Committing test match should do nothing.
	match := &model.Match{Id: 0, Type: "test", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	err := web.commitMatchScore(match, &model.MatchResult{MatchId: match.Id}, true)
	assert.Nil(t, err)
	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)

	// Committing the same match more than once should create a second match result record.
	match.Type = "qualification"
	assert.Nil(t, web.arena.Database.CreateMatch(match))
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueScore = &game.Score{TaxiStatuses: [3]bool{true, false, false}}
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, model.BlueWonMatch, match.Status)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.RedScore = &game.Score{TaxiStatuses: [3]bool{true, false, true}}
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, model.RedWonMatch, match.Status)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, model.TieMatch, match.Status)

	// Verify TBA publishing by checking the log for the expected failure messages.
	web.arena.TbaClient.BaseUrl = "fakeUrl"
	web.arena.EventSettings.TbaPublishingEnabled = true
	var writer bytes.Buffer
	log.SetOutput(&writer)
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100) // Allow some time for the asynchronous publishing to happen.
	assert.Contains(t, writer.String(), "Failed to publish matches")
	assert.Contains(t, writer.String(), "Failed to publish rankings")
}

func TestCommitEliminationTie(t *testing.T) {
	web := setupTestWeb(t)

	match := &model.Match{Id: 0, Type: "qualification", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
	web.arena.Database.CreateMatch(match)
	matchResult := &model.MatchResult{
		MatchId: match.Id,
		RedScore: &game.Score{
			TeleopCargoUpper: [4]int{1, 2, 0, 3},
			Fouls:            []game.Foul{{RuleId: 1}, {RuleId: 2}, {RuleId: 4}}},
		BlueScore: &game.Score{},
	}
	err := web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, model.TieMatch, match.Status)
	match.Type = "elimination"
	web.arena.Database.UpdateMatch(match)
	web.commitMatchScore(match, matchResult, true)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, model.TieMatch, match.Status) // No elimination tiebreakers.
}

func TestCommitCards(t *testing.T) {
	web := setupTestWeb(t)

	// Check that a yellow card sticks with a team.
	team := &model.Team{Id: 5}
	web.arena.Database.CreateTeam(team)
	match := &model.Match{Id: 0, Type: "qualification", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
	assert.Nil(t, web.arena.Database.CreateMatch(match))
	matchResult := model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueCards = map[string]string{"5": "yellow"}
	err := web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	team, _ = web.arena.Database.GetTeamById(5)
	assert.True(t, team.YellowCard)

	// Check that editing a match result removes a yellow card from a team.
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	team, _ = web.arena.Database.GetTeamById(5)
	assert.False(t, team.YellowCard)

	// Check that a red card causes a yellow card to stick with a team.
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueCards = map[string]string{"5": "red"}
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	team, _ = web.arena.Database.GetTeamById(5)
	assert.True(t, team.YellowCard)

	// Check that a red card in eliminations zeroes out the score.
	tournament.CreateTestAlliances(web.arena.Database, 2)
	match.Type = "elimination"
	match.ElimRedAlliance = 1
	match.ElimBlueAlliance = 2
	web.arena.Database.UpdateMatch(match)
	matchResult = model.BuildTestMatchResult(match.Id, 0)
	matchResult.MatchType = match.Type
	matchResult.RedCards = map[string]string{"1": "red"}
	assert.Nil(t, web.commitMatchScore(match, matchResult, true))
	assert.Equal(t, 0, matchResult.RedScoreSummary().Score)
	assert.NotEqual(t, 0, matchResult.BlueScoreSummary().Score)
}

func TestMatchPlayWebsocketCommands(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "scoringStatus")
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	readWebsocketType(t, ws, "eventStatus")

	// Test that a server-side error is communicated to the client.
	ws.Write("nonexistenttype", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Invalid message type")

	// Test match setup commands.
	ws.Write("substituteTeam", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("substituteTeam", map[string]interface{}{"team": 254, "position": "B5"})
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("substituteTeam", map[string]interface{}{"team": 254, "position": "B1"})
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, 254, web.arena.CurrentMatch.Blue1)
	ws.Write("substituteTeam", map[string]interface{}{"team": 0, "position": "B1"})
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, 0, web.arena.CurrentMatch.Blue1)
	ws.Write("toggleBypass", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Failed to parse")
	ws.Write("toggleBypass", "R4")
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("toggleBypass", "R3")
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, true, web.arena.AllianceStations["R3"].Bypass)
	ws.Write("toggleBypass", "R3")
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, false, web.arena.AllianceStations["R3"].Bypass)

	// Go through match flow.
	ws.Write("abortMatch", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot abort match")
	ws.Write("startMatch", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot start match")
	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	ws.Write("startMatch", nil)
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "eventStatus")
	assert.Equal(t, field.StartMatch, web.arena.MatchState)
	ws.Write("commitResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("discardResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Cannot reset match")
	ws.Write("abortMatch", nil)
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	assert.Equal(t, field.PostMatch, web.arena.MatchState)
	web.arena.RedRealtimeScore.CurrentScore.TeleopCargoUpper = [4]int{1, 1, 1, 4}
	web.arena.BlueRealtimeScore.CurrentScore.TaxiStatuses = [3]bool{true, false, true}
	ws.Write("commitResults", nil)
	readWebsocketMultiple(t, ws, 3) // reload, realtimeScore, setAllianceStationDisplay
	assert.Equal(t, [4]int{1, 1, 1, 4}, web.arena.SavedMatchResult.RedScore.TeleopCargoUpper)
	assert.Equal(t, [3]bool{true, false, true}, web.arena.SavedMatchResult.BlueScore.TaxiStatuses)
	assert.Equal(t, field.PreMatch, web.arena.MatchState)
	ws.Write("discardResults", nil)
	readWebsocketMultiple(t, ws, 3) // reload, realtimeScore, setAllianceStationDisplay
	assert.Equal(t, field.PreMatch, web.arena.MatchState)

	// Test changing the displays.
	ws.Write("setAudienceDisplay", "logo")
	readWebsocketType(t, ws, "audienceDisplayMode")
	assert.Equal(t, "logo", web.arena.AudienceDisplayMode)
	ws.Write("setAllianceStationDisplay", "logo")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	assert.Equal(t, "logo", web.arena.AllianceStationDisplayMode)
}

func TestMatchPlayWebsocketNotifications(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateTeam(&model.Team{Id: 254})

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "scoringStatus")
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	readWebsocketType(t, ws, "eventStatus")

	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, web.arena.StartMatch())
	web.arena.Update()
	messages := readWebsocketMultiple(t, ws, 5)
	_, ok := messages["matchTime"]
	assert.True(t, ok)
	_, ok = messages["audienceDisplayMode"]
	assert.True(t, ok)
	_, ok = messages["allianceStationDisplayMode"]
	assert.True(t, ok)
	_, ok = messages["eventStatus"]
	assert.True(t, ok)
	web.arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	web.arena.Update()
	messages = readWebsocketMultiple(t, ws, 2)
	statusReceived, matchTime := getStatusMatchTime(t, messages)
	assert.Equal(t, true, statusReceived)
	assert.Equal(t, field.AutoPeriod, matchTime.MatchState)
	assert.Equal(t, 3, matchTime.MatchTimeSec)
	web.arena.ScoringStatusNotifier.Notify()
	readWebsocketType(t, ws, "scoringStatus")

	// Should get a tick notification when an integer second threshold is crossed.
	web.arena.MatchStartTime = time.Now().Add(-time.Second - 10*time.Millisecond) // Crossed
	web.arena.Update()
	err = mapstructure.Decode(readWebsocketType(t, ws, "matchTime"), &matchTime)
	assert.Nil(t, err)
	assert.Equal(t, field.AutoPeriod, matchTime.MatchState)
	assert.Equal(t, 1, matchTime.MatchTimeSec)
	web.arena.MatchStartTime = time.Now().Add(-2*time.Second + 10*time.Millisecond) // Not crossed yet
	web.arena.Update()
	web.arena.MatchStartTime = time.Now().Add(-2*time.Second - 10*time.Millisecond) // Crossed
	web.arena.Update()
	err = mapstructure.Decode(readWebsocketType(t, ws, "matchTime"), &matchTime)
	assert.Nil(t, err)
	assert.Equal(t, field.AutoPeriod, matchTime.MatchState)
	assert.Equal(t, 2, matchTime.MatchTimeSec)

	// Check across a match state boundary.
	web.arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec) * time.Second)
	web.arena.Update()
	statusReceived, matchTime = readWebsocketStatusMatchTime(t, ws)
	assert.Equal(t, true, statusReceived)
	assert.Equal(t, field.PausePeriod, matchTime.MatchState)
	assert.Equal(t, game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec, matchTime.MatchTimeSec)
}

// Handles the status and matchTime messages arriving in either order.
func readWebsocketStatusMatchTime(t *testing.T, ws *websocket.Websocket) (bool, field.MatchTimeMessage) {
	return getStatusMatchTime(t, readWebsocketMultiple(t, ws, 2))
}

func getStatusMatchTime(t *testing.T, messages map[string]interface{}) (bool, field.MatchTimeMessage) {
	_, statusReceived := messages["arenaStatus"]
	message, ok := messages["matchTime"]
	var matchTime field.MatchTimeMessage
	if assert.True(t, ok) {
		err := mapstructure.Decode(message, &matchTime)
		assert.Nil(t, err)
	}
	return statusReceived, matchTime
}
