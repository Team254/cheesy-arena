// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNonexistentMatch(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match, err := db.GetMatchById(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{
		Type:                Qualification,
		TypeOrder:           254,
		Time:                time.Unix(1114, 0).UTC(),
		LongName:            "Qualification 254",
		ShortName:           "Q254",
		NameDetail:          "Qual Round",
		Red1:                1,
		Red2:                2,
		Red3:                3,
		Blue1:               4,
		Blue2:               5,
		Blue3:               6,
		UseTiebreakCriteria: true,
		TbaMatchKey:         TbaMatchKey{"qm", 0, 254},
	}
	assert.Nil(t, db.CreateMatch(&match))
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match, *match2)
	match3, err := db.GetMatchByTypeOrder(Qualification, 254)
	assert.Nil(t, err)
	assert.Equal(t, match, *match3)

	match.Status = game.RedWonMatch
	assert.Nil(t, db.UpdateMatch(&match))
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match.Status, match2.Status)

	assert.Nil(t, db.DeleteMatch(match.Id))
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestTruncateMatches(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{
		Type:      Qualification,
		TypeOrder: 254,
		ShortName: "Q254",
		LongName:  "Qualification 254",
		Red1:      1,
		Red2:      2,
		Red3:      3,
		Blue1:     4,
		Blue2:     5,
		Blue3:     6,
	}
	assert.Nil(t, db.CreateMatch(&match))
	assert.Nil(t, db.TruncateMatches())
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestGetMatchByTypeOrder(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match1 := Match{
		Type:      Practice,
		TypeOrder: 2,
		ShortName: "P2",
	}
	assert.Nil(t, db.CreateMatch(&match1))
	match2 := Match{
		Type:      Qualification,
		TypeOrder: 2,
		ShortName: "Q2",
	}
	assert.Nil(t, db.CreateMatch(&match2))

	match, err := db.GetMatchByTypeOrder(Qualification, 1)
	assert.Nil(t, err)
	assert.Nil(t, match)

	match, err = db.GetMatchByTypeOrder(Qualification, 2)
	assert.Nil(t, err)
	assert.Equal(t, match2, *match)

	match, err = db.GetMatchByTypeOrder(Practice, 2)
	assert.Nil(t, err)
	assert.Equal(t, match1, *match)
}

func TestGetMatchesByType(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match1 := Match{
		Type:      Qualification,
		TypeOrder: 1,
		ShortName: "Q1",
	}
	assert.Nil(t, db.CreateMatch(&match1))
	match3 := Match{
		Type:      Practice,
		TypeOrder: 2,
		ShortName: "P2",
	}
	assert.Nil(t, db.CreateMatch(&match3))
	match2 := Match{
		Type:      Practice,
		TypeOrder: 1,
		ShortName: "P1",
	}
	assert.Nil(t, db.CreateMatch(&match2))

	matches, err := db.GetMatchesByType(Test, false)
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByType(Practice, false)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assert.Equal(t, match2, matches[0])
		assert.Equal(t, match3, matches[1])
	}
	matches, err = db.GetMatchesByType(Qualification, false)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(matches)) {
		assert.Equal(t, match1, matches[0])
	}

	// Test filtering of hidden matches.
	match3.Status = game.MatchHidden
	assert.Nil(t, db.UpdateMatch(&match3))
	matches, err = db.GetMatchesByType(Practice, false)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(matches)) {
		assert.Equal(t, match2, matches[0])
	}
	matches, err = db.GetMatchesByType(Practice, true)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assert.Equal(t, match2, matches[0])
		assert.Equal(t, match3, matches[1])
	}
}

func TestMatchTypeFromString(t *testing.T) {
	matchType, err := MatchTypeFromString("test")
	assert.Nil(t, err)
	assert.Equal(t, Test, matchType)

	matchType, err = MatchTypeFromString("practice")
	assert.Nil(t, err)
	assert.Equal(t, Practice, matchType)

	matchType, err = MatchTypeFromString("qualification")
	assert.Nil(t, err)
	assert.Equal(t, Qualification, matchType)

	matchType, err = MatchTypeFromString("Qualification")
	assert.Nil(t, err)
	assert.Equal(t, Qualification, matchType)

	matchType, err = MatchTypeFromString("playoff")
	assert.Nil(t, err)
	assert.Equal(t, Playoff, matchType)

	matchType, err = MatchTypeFromString("blorpy")
	if assert.NotNil(t, err) {
		assert.Equal(t, "invalid match type \"blorpy\"", err.Error())
	}

	matchType, err = MatchTypeFromString("elimination")
	if assert.NotNil(t, err) {
		assert.Equal(t, "invalid match type \"elimination\"", err.Error())
	}
}
