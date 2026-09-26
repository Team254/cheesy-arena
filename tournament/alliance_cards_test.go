// Copyright 2026 Team 254. All Rights Reserved.

package tournament

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCalculateAllianceCards(t *testing.T) {
	db := setupTestDb(t)
	CreateTestAlliances(db, 2)
	require.NoError(t, db.CreateTeam(&model.Team{Id: 101, YellowCard: true}))
	require.NoError(t, db.CreateTeam(&model.Team{Id: 104}))
	match := &model.Match{Type: model.Playoff, PlayoffRedAlliance: 1, PlayoffBlueAlliance: 2, Status: game.RedWonMatch}
	require.NoError(t, db.CreateMatch(match))
	result := model.NewMatchResult()
	result.MatchId, result.PlayNumber = match.Id, 1
	// Legacy maps must never affect playoff carryover.
	result.RedCards = map[string]string{"201": "red"}
	require.NoError(t, db.CreateMatchResult(result))
	for _, tc := range []struct {
		card   string
		yellow bool
	}{{"yellow", true}, {"red", true}, {"dq", false}, {"", false}} {
		t.Run(tc.card, func(t *testing.T) {
			result.PlayoffRedAllianceCard = tc.card
			require.NoError(t, db.UpdateMatchResult(result))
			require.NoError(t, CalculateAllianceCards(db))
			red, _ := db.GetAllianceById(1)
			blue, _ := db.GetAllianceById(2)
			assert.Equal(t, tc.yellow, red.YellowCard)
			assert.False(t, blue.YellowCard)
			team, _ := db.GetTeamById(101)
			assert.True(t, team.YellowCard)
			team, _ = db.GetTeamById(104)
			assert.False(t, team.YellowCard)
		})
	}
	// The same alliance plays blue in another match; corrections must respect its other cards.
	match2 := &model.Match{Type: model.Playoff, PlayoffRedAlliance: 2, PlayoffBlueAlliance: 1, Status: game.BlueWonMatch}
	require.NoError(t, db.CreateMatch(match2))
	result2 := model.NewMatchResult()
	result2.MatchId, result2.PlayNumber = match2.Id, 1
	result2.PlayoffBlueAllianceCard = "yellow"
	require.NoError(t, db.CreateMatchResult(result2))
	require.NoError(t, CalculateAllianceCards(db))
	alliance, _ := db.GetAllianceById(1)
	assert.True(t, alliance.YellowCard)
	// Removing a card from one match must preserve a yellow established in another.
	result.PlayoffRedAllianceCard = "red"
	require.NoError(t, db.UpdateMatchResult(result))
	require.NoError(t, CalculateAllianceCards(db))
	result.PlayoffRedAllianceCard = ""
	require.NoError(t, db.UpdateMatchResult(result))
	require.NoError(t, CalculateAllianceCards(db))
	alliance, _ = db.GetAllianceById(1)
	assert.True(t, alliance.YellowCard)
	// A replay replaces the effective result, rather than accumulating superseded cards.
	replay := model.NewMatchResult()
	replay.MatchId, replay.PlayNumber = match2.Id, 2
	require.NoError(t, db.CreateMatchResult(replay))
	require.NoError(t, CalculateAllianceCards(db))
	alliance, _ = db.GetAllianceById(1)
	assert.False(t, alliance.YellowCard)
	// An uncompleted match with a saved card must not establish carryover.
	match2.Status = game.MatchScheduled
	require.NoError(t, db.UpdateMatch(match2))
	replay.PlayoffBlueAllianceCard = "red"
	require.NoError(t, db.UpdateMatchResult(replay))
	require.NoError(t, CalculateAllianceCards(db))
	alliance, _ = db.GetAllianceById(1)
	assert.False(t, alliance.YellowCard)
}
