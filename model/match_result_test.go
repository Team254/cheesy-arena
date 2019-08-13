// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentMatchResult(t *testing.T) {
	db := setupTestDb(t)

	match, err := db.GetMatchResultForMatch(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchResultCrud(t *testing.T) {
	db := setupTestDb(t)

	matchResult := BuildTestMatchResult(254, 5)
	db.CreateMatchResult(matchResult)
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	matchResult.BlueScore.RobotEndLevels = [3]int{3, 3, 3}
	db.SaveMatchResult(matchResult)
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult, matchResult2)

	db.DeleteMatchResult(matchResult)
	matchResult2, err = db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestTruncateMatchResults(t *testing.T) {
	db := setupTestDb(t)

	matchResult := BuildTestMatchResult(254, 1)
	db.CreateMatchResult(matchResult)
	db.TruncateMatchResults()
	matchResult2, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Nil(t, matchResult2)
}

func TestGetMatchResultForMatch(t *testing.T) {
	db := setupTestDb(t)

	matchResult := BuildTestMatchResult(254, 2)
	db.CreateMatchResult(matchResult)
	matchResult2 := BuildTestMatchResult(254, 5)
	db.CreateMatchResult(matchResult2)
	matchResult3 := BuildTestMatchResult(254, 4)
	db.CreateMatchResult(matchResult3)

	// Should return the match result with the highest play number (i.e. the most recent).
	matchResult4, err := db.GetMatchResultForMatch(254)
	assert.Nil(t, err)
	assert.Equal(t, matchResult2, matchResult4)
}
