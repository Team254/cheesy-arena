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

	match := Match{0, Qualification, "254", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), time.Now().UTC(), game.MatchNotPlayed}
	db.CreateMatch(&match)
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match, *match2)
	match3, err := db.GetMatchByName(Qualification, "254")
	assert.Nil(t, err)
	assert.Equal(t, match, *match3)

	match.Status = game.RedWonMatch
	db.UpdateMatch(&match)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match.Status, match2.Status)

	db.DeleteMatch(match.Id)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestTruncateMatches(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{0, Qualification, "254", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), time.Now().UTC(), game.MatchNotPlayed}
	db.CreateMatch(&match)
	db.TruncateMatches()
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestGetMatchesByPlayoffRoundGroup(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{Type: Playoff, DisplayName: "SF1-1", PlayoffRound: 2, PlayoffGroup: 1, PlayoffInstance: 1,
		PlayoffRedAlliance: 8, PlayoffBlueAlliance: 4}
	db.CreateMatch(&match)
	match2 := Match{Type: Playoff, DisplayName: "SF2-2", PlayoffRound: 2, PlayoffGroup: 2, PlayoffInstance: 2,
		PlayoffRedAlliance: 2, PlayoffBlueAlliance: 3}
	db.CreateMatch(&match2)
	match3 := Match{Type: Playoff, DisplayName: "SF2-1", PlayoffRound: 2, PlayoffGroup: 2, PlayoffInstance: 1,
		PlayoffRedAlliance: 8, PlayoffBlueAlliance: 4}
	db.CreateMatch(&match3)
	match4 := Match{Type: Playoff, DisplayName: "QF2-1", PlayoffRound: 4, PlayoffGroup: 2, PlayoffInstance: 1,
		PlayoffRedAlliance: 4, PlayoffBlueAlliance: 5}
	db.CreateMatch(&match4)
	match5 := Match{Type: Practice, DisplayName: "1"}
	db.CreateMatch(&match5)

	matches, err := db.GetMatchesByPlayoffRoundGroup(4, 1)
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByPlayoffRoundGroup(2, 2)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assert.Equal(t, "SF2-1", matches[0].DisplayName)
		assert.Equal(t, "SF2-2", matches[1].DisplayName)
	}
}

func TestGetMatchesByType(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{0, Qualification, "1", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), time.Now().UTC(), game.MatchNotPlayed}
	db.CreateMatch(&match)
	match2 := Match{0, Practice, "1", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, time.Now().UTC(), time.Now().UTC(), time.Now().UTC(), game.MatchNotPlayed}
	db.CreateMatch(&match2)
	match3 := Match{0, Practice, "2", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, time.Now().UTC(), time.Now().UTC(), time.Now().UTC(), game.MatchNotPlayed}
	db.CreateMatch(&match3)

	matches, err := db.GetMatchesByType(Test)
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByType(Practice)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(matches))
	matches, err = db.GetMatchesByType(Qualification)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))
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
