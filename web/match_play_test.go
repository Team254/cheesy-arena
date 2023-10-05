// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"bytes"
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

	// Check that some text near the bottom of the page is present.
	recorder := web.getHttpResponse("/match_play")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Are you sure you want to discard the results for this match?")
}

func TestMatchPlayMatchList(t *testing.T) {
	web := setupTestWeb(t)

	match1 := model.Match{Type: model.Practice, ShortName: "P1", Status: game.RedWonMatch}
	match2 := model.Match{Type: model.Practice, ShortName: "P2"}
	match3 := model.Match{Type: model.Qualification, ShortName: "Q1", Status: game.BlueWonMatch}
	match4 := model.Match{Type: model.Playoff, ShortName: "SF1-1", Status: game.TieMatch}
	match5 := model.Match{Type: model.Playoff, ShortName: "SF1-2"}
	web.arena.Database.CreateMatch(&match1)
	web.arena.Database.CreateMatch(&match2)
	web.arena.Database.CreateMatch(&match3)
	web.arena.Database.CreateMatch(&match4)
	web.arena.Database.CreateMatch(&match5)

	// Check that all matches are listed on the page.
	recorder := web.getHttpResponse("/match_play/match_load")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "P1")
	assert.Contains(t, recorder.Body.String(), "P2")
	assert.Contains(t, recorder.Body.String(), "Q1")
	assert.Contains(t, recorder.Body.String(), "SF1-1")
	assert.Contains(t, recorder.Body.String(), "SF1-2")
}

func TestCommitMatch(t *testing.T) {
	web := setupTestWeb(t)

	// Committing test match should update the stored saved match but not persist anything.
	match := &model.Match{Id: 0, Type: model.Test, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	matchResult := &model.MatchResult{MatchId: match.Id, RedScore: &game.Score{}, BlueScore: &game.Score{}}
	matchResult.BlueScore.MobilityStatuses[2] = true
	err := web.commitMatchScore(match, matchResult, false)
	assert.Nil(t, err)
	matchResult, err = web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)
	assert.Equal(t, match, web.arena.SavedMatch)
	assert.Equal(t, game.BlueWonMatch, web.arena.SavedMatch.Status)

	// Committing the same match more than once should create a second match result record.
	match.Type = model.Qualification
	assert.Nil(t, web.arena.Database.CreateMatch(match))
	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.BlueScore = &game.Score{MobilityStatuses: [3]bool{true, false, false}}
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.BlueWonMatch, match.Status)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	matchResult.RedScore = &game.Score{MobilityStatuses: [3]bool{true, false, true}}
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 2, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.RedWonMatch, match.Status)

	matchResult = model.NewMatchResult()
	matchResult.MatchId = match.Id
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, matchResult.PlayNumber)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.TieMatch, match.Status)

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

func TestCommitTiebreak(t *testing.T) {
	web := setupTestWeb(t)

	match := &model.Match{
		Type:                model.Qualification,
		TypeOrder:           1,
		Red1:                1,
		Red2:                2,
		Red3:                3,
		Blue1:               4,
		Blue2:               5,
		Blue3:               6,
		UseTiebreakCriteria: false,
	}
	web.arena.Database.CreateMatch(match)
	matchResult := &model.MatchResult{
		MatchId: match.Id,
		// These should all be fields that aren't part of the tiebreaker.
		RedScore: &game.Score{
			Grid:  game.Grid{Nodes: [3][9]game.NodeState{{game.Cube}, {game.Cone}}},
			Fouls: []game.Foul{{RuleId: 1}, {RuleId: 2}},
		},
		BlueScore: &game.Score{
			Fouls: []game.Foul{{RuleId: 1}},
		},
	}

	// Sanity check that the test scores are equal; they will need to be updated accordingly for each new game.
	assert.Equal(
		t,
		matchResult.RedScore.Summarize(matchResult.BlueScore).Score,
		matchResult.BlueScore.Summarize(matchResult.RedScore).Score,
	)

	err := web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.TieMatch, match.Status)

	// The match should still be tied since the tiebreaker criteria for a perfect tie are fulfilled.
	match.UseTiebreakCriteria = true
	web.arena.Database.UpdateMatch(match)
	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.TieMatch, match.Status)

	// Change the score to still be equal nominally but trigger the tiebreaker criteria.
	matchResult.BlueScore.AutoDockStatuses = [3]bool{true, false, false}
	matchResult.BlueScore.AutoChargeStationLevel = true
	matchResult.BlueScore.Fouls = []game.Foul{{IsTechnical: false}, {IsTechnical: true}}

	// Sanity check that the test scores are equal; they will need to be updated accordingly for each new game.
	assert.Equal(
		t,
		matchResult.RedScore.Summarize(matchResult.BlueScore).Score,
		matchResult.BlueScore.Summarize(matchResult.RedScore).Score,
	)

	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.RedWonMatch, match.Status)

	// Swap red and blue and verify that the tie is broken in the other direction.
	matchResult.RedScore, matchResult.BlueScore = matchResult.BlueScore, matchResult.RedScore

	// Sanity check that the test scores are equal; they will need to be updated accordingly for each new game.
	assert.Equal(
		t,
		matchResult.RedScore.Summarize(matchResult.BlueScore).Score,
		matchResult.BlueScore.Summarize(matchResult.RedScore).Score,
	)

	err = web.commitMatchScore(match, matchResult, true)
	assert.Nil(t, err)
	match, _ = web.arena.Database.GetMatchById(1)
	assert.Equal(t, game.BlueWonMatch, match.Status)
}

func TestCommitCards(t *testing.T) {
	web := setupTestWeb(t)

	// Check that a yellow card sticks with a team.
	team := &model.Team{Id: 5}
	web.arena.Database.CreateTeam(team)
	match := &model.Match{Id: 0, Type: model.Qualification, Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
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

	// Check that a red card in playoffs zeroes out the score.
	tournament.CreateTestAlliances(web.arena.Database, 2)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())
	match.Type = model.Playoff
	match.PlayoffRedAlliance = 1
	match.PlayoffBlueAlliance = 2
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
	web.arena.Database.CreateTeam(&model.Team{Id: 254})

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "scorePosted")
	readWebsocketType(t, ws, "scoringStatus")

	// Test that a server-side error is communicated to the client.
	ws.Write("nonexistenttype", nil)
	assert.Contains(t, readWebsocketError(t, ws), "Invalid message type")

	// Test match setup commands.
	ws.Write("substituteTeams", map[string]int{"Red1": 0, "Red2": 0, "Red3": 0, "Blue1": 1, "Blue2": 0, "Blue3": 0})
	assert.Equal(t, readWebsocketError(t, ws), "Team 1 is not present at the event.")
	ws.Write("substituteTeams", map[string]int{"Red1": 0, "Red2": 0, "Red3": 0, "Blue1": 254, "Blue2": 0, "Blue3": 0})
	readWebsocketType(t, ws, "matchLoad")
	assert.Equal(t, 254, web.arena.CurrentMatch.Blue1)
	ws.Write("substituteTeams", map[string]int{"Red1": 0, "Red2": 0, "Red3": 0, "Blue1": 0, "Blue2": 0, "Blue3": 0})
	readWebsocketType(t, ws, "matchLoad")
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
	assert.Contains(t, readWebsocketError(t, ws), "cannot abort match")
	ws.Write("startMatch", nil)
	assert.Contains(t, readWebsocketError(t, ws), "cannot start match")
	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	ws.Write("startMatch", nil)
	readWebsocketType(t, ws, "eventStatus")
	assert.Equal(t, field.StartMatch, web.arena.MatchState)
	ws.Write("commitResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "cannot commit match while it is in progress")
	ws.Write("discardResults", nil)
	assert.Contains(t, readWebsocketError(t, ws), "cannot reset match while it is in progress")
	ws.Write("abortMatch", nil)
	readWebsocketType(t, ws, "audienceDisplayMode")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	assert.Equal(t, field.PostMatch, web.arena.MatchState)
	web.arena.RedRealtimeScore.CurrentScore.AutoDockStatuses = [3]bool{false, true, true}
	web.arena.BlueRealtimeScore.CurrentScore.MobilityStatuses = [3]bool{true, false, true}
	ws.Write("commitResults", nil)
	readWebsocketMultiple(t, ws, 5) // scorePosted, matchLoad, realtimeScore, allianceStationDisplayMode, scoringStatus
	assert.Equal(t, [3]bool{false, true, true}, web.arena.SavedMatchResult.RedScore.AutoDockStatuses)
	assert.Equal(t, [3]bool{true, false, true}, web.arena.SavedMatchResult.BlueScore.MobilityStatuses)
	assert.Equal(t, field.PreMatch, web.arena.MatchState)
	ws.Write("discardResults", nil)
	readWebsocketMultiple(t, ws, 4) // matchLoad, realtimeScore, allianceStationDisplayMode, scoringStatus
	assert.Equal(t, field.PreMatch, web.arena.MatchState)

	// Test changing the displays.
	ws.Write("setAudienceDisplay", "logo")
	readWebsocketType(t, ws, "audienceDisplayMode")
	assert.Equal(t, "logo", web.arena.AudienceDisplayMode)
	ws.Write("setAllianceStationDisplay", "logo")
	readWebsocketType(t, ws, "allianceStationDisplayMode")
	assert.Equal(t, "logo", web.arena.AllianceStationDisplayMode)
}

func TestMatchPlayWebsocketLoadMatch(t *testing.T) {
	web := setupTestWeb(t)
	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.CreatePlayoffTournament()

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketMultiple(t, ws, 10)

	web.arena.Database.CreateTeam(&model.Team{Id: 101})
	web.arena.Database.CreateTeam(&model.Team{Id: 102})
	web.arena.Database.CreateTeam(&model.Team{Id: 103})
	web.arena.Database.CreateTeam(&model.Team{Id: 104})
	web.arena.Database.CreateTeam(&model.Team{Id: 105})
	web.arena.Database.CreateTeam(&model.Team{Id: 106})
	match := model.Match{Type: model.Playoff, ShortName: "QF4-3", Status: game.RedWonMatch, Red1: 101,
		Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	web.arena.Database.CreateMatch(&match)

	matchIdMessage := struct{ MatchId int }{match.Id}
	ws.Write("loadMatch", matchIdMessage)
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketMultiple(t, ws, 3)
	assert.Equal(t, 101, web.arena.CurrentMatch.Red1)
	assert.Equal(t, 102, web.arena.CurrentMatch.Red2)
	assert.Equal(t, 103, web.arena.CurrentMatch.Red3)
	assert.Equal(t, 104, web.arena.CurrentMatch.Blue1)
	assert.Equal(t, 105, web.arena.CurrentMatch.Blue2)
	assert.Equal(t, 106, web.arena.CurrentMatch.Blue3)

	// Load a test match.
	matchIdMessage.MatchId = 0
	ws.Write("loadMatch", matchIdMessage)
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketMultiple(t, ws, 3)
	assert.Equal(t, 0, web.arena.CurrentMatch.Red1)
	assert.Equal(t, 0, web.arena.CurrentMatch.Red2)
	assert.Equal(t, 0, web.arena.CurrentMatch.Red3)
	assert.Equal(t, 0, web.arena.CurrentMatch.Blue1)
	assert.Equal(t, 0, web.arena.CurrentMatch.Blue2)
	assert.Equal(t, 0, web.arena.CurrentMatch.Blue3)

	// Load a nonexistent match.
	matchIdMessage.MatchId = 254
	ws.Write("loadMatch", matchIdMessage)
	assert.Contains(t, readWebsocketError(t, ws), "invalid match ID 254")
}

func TestMatchPlayWebsocketShowAndClearResult(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/match_play/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketMultiple(t, ws, 10)

	matchIdMessage := struct{ MatchId int }{1}
	ws.Write("showResult", matchIdMessage)
	assert.Contains(t, readWebsocketError(t, ws), "invalid match ID 1")

	match := model.Match{Type: model.Qualification, ShortName: "Q1", Status: game.TieMatch}
	web.arena.Database.CreateMatch(&match)
	ws.Write("showResult", matchIdMessage)
	assert.Contains(t, readWebsocketError(t, ws), "No result found")

	web.arena.Database.CreateMatchResult(model.BuildTestMatchResult(match.Id, 1))
	ws.Write("showResult", matchIdMessage)
	readWebsocketType(t, ws, "scorePosted")
	assert.Equal(t, match.Id, web.arena.SavedMatch.Id)
	assert.Equal(t, match.Id, web.arena.SavedMatchResult.MatchId)

	matchIdMessage.MatchId = 0
	ws.Write("showResult", matchIdMessage)
	readWebsocketType(t, ws, "scorePosted")
	assert.Equal(t, model.Match{}, *web.arena.SavedMatch)
	assert.Equal(t, *model.NewMatchResult(), *web.arena.SavedMatchResult)
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
	readWebsocketMultiple(t, ws, 10)

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

func getStatusMatchTime(t *testing.T, messages map[string]any) (bool, field.MatchTimeMessage) {
	_, statusReceived := messages["arenaStatus"]
	message, ok := messages["matchTime"]
	var matchTime field.MatchTimeMessage
	if assert.True(t, ok) {
		err := mapstructure.Decode(message, &matchTime)
		assert.Nil(t, err)
	}
	return statusReceived, matchTime
}
