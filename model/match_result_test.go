// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentMatchResult(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match, err := db.GetMatchResultForMatch(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchResultCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	matchResult := BuildTestMatchResult(254, 5)
	assert.Nil(t, db.CreateMatchResult(matchResult))
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	matchResult.BlueScore.EndgameTowerStatuses =
		[3]game.TowerStatus{game.TowerLevel1, game.TowerNone, game.TowerLevel2}
	assert.Nil(t, db.UpdateMatchResult(matchResult))
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	assert.Nil(t, db.DeleteMatchResult(matchResult.Id))
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestTruncateMatchResults(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	matchResult := BuildTestMatchResult(254, 1)
	assert.Nil(t, db.CreateMatchResult(matchResult))
	assert.Nil(t, db.TruncateMatchResults())
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestGetMatchResultForMatch(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	matchResult := BuildTestMatchResult(254, 2)
	assert.Nil(t, db.CreateMatchResult(matchResult))
	matchResult2 := BuildTestMatchResult(254, 5)
	assert.Nil(t, db.CreateMatchResult(matchResult2))
	matchResult3 := BuildTestMatchResult(254, 4)
	assert.Nil(t, db.CreateMatchResult(matchResult3))

	// Should return the match result with the highest play number (i.e. the most recent).
	matchResult4, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult2, matchResult4)
}

func TestCorrectPlayoffScoreResetsDqState(t *testing.T) {
	matchResult := NewMatchResult()
	matchResult.RedScore.PlayoffDq = true
	matchResult.BlueScore.PlayoffDq = true
	matchResult.RedCards = map[string]string{"1": "red"}
	matchResult.BlueCards = map[string]string{}

	matchResult.CorrectPlayoffScore()
	assert.Equal(t, true, matchResult.RedScore.PlayoffDq)
	assert.Equal(t, false, matchResult.BlueScore.PlayoffDq)

	matchResult.RedCards = map[string]string{}
	matchResult.BlueCards = map[string]string{"4": "dq"}

	matchResult.CorrectPlayoffScore()
	assert.Equal(t, false, matchResult.RedScore.PlayoffDq)
	assert.Equal(t, true, matchResult.BlueScore.PlayoffDq)
}
