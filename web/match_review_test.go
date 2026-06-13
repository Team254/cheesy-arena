// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMatchReview(t *testing.T) {
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
	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">P1<")
	assert.Contains(t, recorder.Body.String(), ">P2<")
	assert.Contains(t, recorder.Body.String(), ">Q1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-1<")
	assert.Contains(t, recorder.Body.String(), ">SF1-2<")
	assert.Contains(t, recorder.Body.String(), "match-review-rps")
	assert.Contains(t, recorder.Body.String(), "&#x2612;")
}

func TestMatchReviewEditExistingResult(t *testing.T) {
	web := setupTestWeb(t)

	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 8
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	match, _ := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	match.Status = game.RedWonMatch
	web.arena.Database.UpdateMatch(match)
	matchResult := model.BuildTestMatchResult(match.Id, 1)
	matchResult.MatchType = match.Type
	assert.Nil(t, web.arena.Database.CreateMatchResult(matchResult))

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Regexp(t, `(?s)>\s*133\s*</td>`, recorder.Body.String()) // The red score
	assert.Regexp(t, `(?s)>\s*289\s*</td>`, recorder.Body.String()) // The blue score
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")

	// Check response for non-existent match.
	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", 12345))
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No such match")

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Quarterfinal 4-3 ")
	assert.Contains(t, recorder.Body.String(), "AutoTowerStatuses")
	assert.Contains(t, recorder.Body.String(), "HubShiftCount7")
	assert.Contains(t, recorder.Body.String(), "EndgameTowerStatuses")
	assert.Contains(t, recorder.Body.String(), `id="redScore"`)
	assert.Contains(t, recorder.Body.String(), `id="blueScore"`)
	assert.Contains(t, recorder.Body.String(), `id="redSummary"`)
	assert.Contains(t, recorder.Body.String(), `id="blueSummary"`)
	assert.Contains(t, recorder.Body.String(), "score-summary-table-red")
	assert.Contains(t, recorder.Body.String(), "score-summary-table-blue")
	assert.NotContains(t, recorder.Body.String(), "score-summary-rp")
	assert.NotContains(t, recorder.Body.String(), "Energized Bonus")
	assert.NotContains(t, recorder.Body.String(), `data-summary-field="BonusRankingPoints"`)
	assert.NotContains(t, recorder.Body.String(), "scoreTemplate")
	assert.NotContains(t, recorder.Body.String(), "text/x-handlebars-template")
	assert.NotContains(t, recorder.Body.String(), "NumFuelGoal")
	assert.NotContains(t, recorder.Body.String(), "Reef")

	// Update the score to something else.
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameTowerStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"AutoTowerStatuses\":[1,0,0],\"Fouls\":[{\"TeamId\":973,\"RuleId\":4}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Regexp(t, `(?s)>\s*35\s*</td>`, recorder.Body.String()) // The red score
	assert.Regexp(t, `(?s)>\s*15\s*</td>`, recorder.Body.String()) // The blue score
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")
}

func TestMatchReviewCreateNewResult(t *testing.T) {
	web := setupTestWeb(t)

	tournament.CreateTestAlliances(web.arena.Database, 8)
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 8
	web.arena.CreatePlayoffTournament()
	web.arena.CreatePlayoffMatches(time.Now())

	match, _ := web.arena.Database.GetMatchByTypeOrder(model.Playoff, 36)
	match.Status = game.RedWonMatch
	web.arena.Database.UpdateMatch(match)

	recorder := web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.NotRegexp(t, `(?s)>\s*35\s*<span class="match-review-rps`, recorder.Body.String()) // The red score
	assert.NotRegexp(t, `(?s)>\s*15\s*<span class="match-review-rps`, recorder.Body.String()) // The blue score

	recorder = web.getHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id))
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Quarterfinal 4-3 ")

	// Update the score to something else.
	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameTowerStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"AutoTowerStatuses\":[1,0,0],\"Fouls\":[{\"TeamId\":973,\"RuleId\":4}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())

	// Check for the updated scores back on the match list page.
	recorder = web.getHttpResponse("/match_review")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ">QF4-3<")
	assert.Regexp(t, `(?s)>\s*35\s*</td>`, recorder.Body.String()) // The red score
	assert.Regexp(t, `(?s)>\s*15\s*</td>`, recorder.Body.String()) // The blue score
	assert.NotContains(t, recorder.Body.String(), "match-review-rps")
}

func TestMatchReviewEditCurrentMatch(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 352",
		ShortName: "Q352",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)
	web.arena.LoadMatch(&match)
	assert.Equal(t, match, *web.arena.CurrentMatch)

	recorder := web.getHttpResponse("/match_review/current/edit")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), " Qualification 352 ")
	assert.Contains(t, recorder.Body.String(), "score-summary-rp")
	assert.Contains(t, recorder.Body.String(), "Energized Bonus")

	postBody := fmt.Sprintf(
		"matchResultJson={\"MatchId\":%d,\"RedScore\":{\"EndgameTowerStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"AutoTowerStatuses\":[1,0,0],\"Fouls\":[{\"TeamId\":973,\"RuleId\":1}]},"+
			"\"RedCards\":{\"105\":\"yellow\"},\"BlueCards\":{}}",
		match.Id,
	)
	recorder = web.postHttpResponse("/match_review/current/edit", postBody)
	assert.Equal(t, 303, recorder.Code, recorder.Body.String())
	assert.Equal(t, "/match_play", recorder.Header().Get("Location"))

	// Check that the persisted match is still unedited and that the realtime scores have been updated instead.
	match2, _ := web.arena.Database.GetMatchById(match.Id)
	assert.Equal(t, game.MatchScheduled, match2.Status)
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerNone, game.TowerLevel2, game.TowerLevel1},
		web.arena.RedRealtimeScore.CurrentScore.EndgameTowerStatuses,
	)
	assert.Equal(
		t,
		[3]game.TowerStatus{game.TowerLevel1, game.TowerNone, game.TowerNone},
		web.arena.BlueRealtimeScore.CurrentScore.AutoTowerStatuses,
	)
	assert.Equal(t, 0, len(web.arena.RedRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.BlueRealtimeScore.CurrentScore.Fouls))
	assert.Equal(t, 1, len(web.arena.RedRealtimeScore.Cards))
	assert.Equal(t, 0, len(web.arena.BlueRealtimeScore.Cards))
}

func TestMatchReviewSummary(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 1",
		ShortName: "Q1",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)

	postBody := fmt.Sprintf(
		"{\"MatchId\":%d,\"RedScore\":{\"EndgameTowerStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"AutoTowerStatuses\":[1,0,0],\"Fouls\":[{\"TeamId\":1004,\"RuleId\":4}]},"+
			"\"RedCards\":{},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse(fmt.Sprintf("/match_review/%d/summary", match.Id), postBody)
	assert.Equal(t, 200, recorder.Code, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header()["Content-Type"][0])

	var response MatchReviewSummaryResponse
	assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 35, response.RedSummary.Score)
	assert.Equal(t, 30, response.RedSummary.MatchPoints)
	assert.Equal(t, 5, response.RedSummary.FoulPoints)
	assert.Equal(t, 15, response.BlueSummary.Score)

	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)

	recorder = web.postHttpResponse(fmt.Sprintf("/match_review/%d/summary", match.Id), "{\"MatchId\":12345}")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "match ID 12345")
}

func TestMatchReviewSummaryCurrentMatch(t *testing.T) {
	web := setupTestWeb(t)

	match := model.Match{
		Type:      model.Qualification,
		LongName:  "Qualification 1",
		ShortName: "Q1",
		Red1:      1001,
		Red2:      1002,
		Red3:      1003,
		Blue1:     1004,
		Blue2:     1005,
		Blue3:     1006,
	}
	web.arena.Database.CreateMatch(&match)
	web.arena.LoadMatch(&match)

	postBody := fmt.Sprintf(
		"{\"MatchId\":%d,\"RedScore\":{\"EndgameTowerStatuses\":[0,2,1]},\"BlueScore\":{"+
			"\"AutoTowerStatuses\":[1,0,0]},\"RedCards\":{},\"BlueCards\":{}}",
		match.Id,
	)
	recorder := web.postHttpResponse("/match_review/current/summary", postBody)
	assert.Equal(t, 200, recorder.Code, recorder.Body.String())

	var response MatchReviewSummaryResponse
	assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 30, response.RedSummary.Score)
	assert.Equal(t, 15, response.BlueSummary.Score)

	assert.Equal(t, [3]game.TowerStatus{}, web.arena.RedRealtimeScore.CurrentScore.EndgameTowerStatuses)
	assert.Equal(t, [3]game.TowerStatus{}, web.arena.BlueRealtimeScore.CurrentScore.AutoTowerStatuses)
	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	assert.Nil(t, err)
	assert.Nil(t, matchResult)
}
