// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

func setupPlayoffCardTest(t *testing.T) (*Web, *model.Match) {
	web := setupTestWeb(t)
	tournament.CreateTestAlliances(web.arena.Database, 2)
	for _, id := range []int{101, 102, 103, 104, 201, 202, 203, 204, 999} {
		require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: id, YellowCard: id == 201}))
	}
	web.arena.EventSettings.PlayoffType = model.SingleEliminationPlayoff
	web.arena.EventSettings.NumPlayoffAlliances = 2
	require.NoError(t, web.arena.CreatePlayoffTournament())
	require.NoError(t, web.arena.CreatePlayoffMatches(time.Now()))
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, false)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	match := &matches[0]
	require.NoError(t, web.arena.LoadMatch(match))
	return web, match
}

func cardMessage(t *testing.T, message any) map[string]any {
	data, err := json.Marshal(message)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	return decoded
}

func TestPlayoffAllianceCardReviewAndCommit(t *testing.T) {
	web, match := setupPlayoffCardTest(t)
	db := web.arena.Database
	page := web.getHttpResponse("/match_review/current/edit")
	require.Equal(t, 200, page.Code)
	assert.Contains(t, page.Body.String(), `name="redPlayoffAllianceCard"`)
	assert.NotContains(t, page.Body.String(), `name="redTeam102Card"`)
	result := model.NewMatchResult()
	result.MatchId = match.Id
	result.PlayoffRedAllianceCard = "yellow"
	result.PlayoffBlueAllianceCard = "dq"
	result.RedScore.AutoTowerStatuses[0] = game.TowerLevel1
	// Conflicting maps must neither cause a red DQ nor establish yellow carryover for blue.
	result.RedCards = map[string]string{"102": "red"}
	result.BlueCards = map[string]string{"201": "yellow"}
	data, err := json.Marshal(result)
	require.NoError(t, err)
	response := web.postHttpResponse("/match_review/current/edit", "matchResultJson="+url.QueryEscape(string(data)))
	require.Equal(t, 303, response.Code, response.Body.String())
	assert.Equal(t, "yellow", web.arena.RedRealtimeScore.PlayoffAllianceCard)
	assert.Empty(t, web.arena.RedRealtimeScore.Cards)
	assert.Empty(t, web.arena.BlueRealtimeScore.Cards)
	assert.False(t, web.arena.RedRealtimeScore.CurrentScore.PlayoffDq)
	assert.True(t, web.arena.BlueRealtimeScore.CurrentScore.PlayoffDq)
	alliance, _ := db.GetAllianceById(1)
	assert.False(t, alliance.YellowCard) // No carryover before commit.
	require.NoError(t, web.commitCurrentMatchScore())
	saved, err := db.GetMatchResultForMatch(match.Id)
	require.NoError(t, err)
	assert.Equal(t, "yellow", saved.PlayoffRedAllianceCard)
	assert.Equal(t, "dq", saved.PlayoffBlueAllianceCard)
	assert.Empty(t, saved.RedCards)
	assert.Empty(t, saved.BlueCards)
	alliance, _ = db.GetAllianceById(1)
	assert.True(t, alliance.YellowCard)
	alliance, _ = db.GetAllianceById(2)
	assert.False(t, alliance.YellowCard)
	team, _ := db.GetTeamById(104)
	assert.False(t, team.YellowCard)
	team, _ = db.GetTeamById(201)
	assert.True(t, team.YellowCard)
	posted := cardMessage(t, web.arena.GenerateScorePostedMessage())
	assert.Equal(t, "yellow", posted["PlayoffRedAllianceCard"])
	assert.Equal(t, []any{float64(104)}, posted["RedOffFieldTeamIds"])
	announcer := web.getHttpResponse("/displays/announcer/score_posted")
	require.Equal(t, 200, announcer.Code, announcer.Body.String())
	assert.Contains(t, announcer.Body.String(), "Alliance</div>")
	assert.Contains(t, announcer.Body.String(), ">yellow</div>")
	// Substitute the fourth team, then a new team not yet in the roster.
	for _, substitute := range []int{104, 999} {
		require.NoError(t, web.arena.SubstituteTeams(substitute, 101, 103, 202, 201, 203))
		loaded := cardMessage(t, web.arena.GenerateMatchLoadMessage())
		assert.Equal(t, true, loaded["PlayoffRedAllianceYellowCard"])
		assert.Equal(t, false, loaded["PlayoffBlueAllianceYellowCard"])
	}
	// Historical corrections must refresh carryover without altering current uncommitted state.
	web.arena.RedRealtimeScore.PlayoffAllianceCard = "red"
	web.arena.RedRealtimeScore.CurrentScore.AutoTowerStatuses[0] = game.TowerLevel3
	saved.PlayoffRedAllianceCard = ""
	data, err = json.Marshal(saved)
	require.NoError(t, err)
	response = web.postHttpResponse(fmt.Sprintf("/match_review/%d/edit", match.Id), "matchResultJson="+url.QueryEscape(string(data)))
	require.Equal(t, 303, response.Code, response.Body.String())
	loaded := cardMessage(t, web.arena.GenerateMatchLoadMessage())
	assert.Equal(t, false, loaded["PlayoffRedAllianceYellowCard"])
	assert.Equal(t, "red", web.arena.RedRealtimeScore.PlayoffAllianceCard)
	assert.Equal(t, game.TowerLevel3, web.arena.RedRealtimeScore.CurrentScore.AutoTowerStatuses[0])
}

func TestPlayoffCardPreviewAndCorrection(t *testing.T) {
	web, match := setupPlayoffCardTest(t)
	result := model.NewMatchResult()
	result.MatchId = match.Id
	result.RedScore.AutoTowerStatuses[0] = game.TowerLevel1
	result.BlueScore.AutoTowerStatuses[0] = game.TowerLevel1
	result.RedCards = map[string]string{"102": "red"}
	for _, tc := range []struct {
		card       string
		dq, yellow bool
	}{
		{"red", true, true}, {"yellow", false, true}, {"", false, false}, {"dq", true, false}, {"", false, false},
	} {
		t.Run(tc.card, func(t *testing.T) {
			result.PlayoffRedAllianceCard = tc.card
			data, err := json.Marshal(result)
			require.NoError(t, err)
			response := web.postHttpResponse("/match_review/current/summary", string(data))
			require.Equal(t, 200, response.Code, response.Body.String())
			var preview MatchReviewSummaryResponse
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
			assert.Equal(t, tc.dq, preview.RedSummary.PlayoffDq)
			require.NoError(t, web.commitMatchScore(match, result, true))
			assert.Equal(t, *result.RedScoreSummary(), *preview.RedSummary)
			alliance, _ := web.arena.Database.GetAllianceById(1)
			assert.Equal(t, tc.yellow, alliance.YellowCard)
		})
	}
}
