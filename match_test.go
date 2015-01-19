// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNonexistentMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match, err := db.GetMatchById(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match := Match{0, "qualification", "254", time.Now().UTC(), 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match, *match2)
	match3, err := db.GetMatchByName("qualification", "254")
	assert.Nil(t, err)
	assert.Equal(t, match, *match3)

	match.Status = "started"
	db.SaveMatch(&match)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match.Status, match2.Status)

	db.DeleteMatch(&match)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestTruncateMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match := Match{0, "qualification", "254", time.Now().UTC(), 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	db.TruncateMatches()
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestGetMatchesByElimRound(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	// TODO(pat): Update for 2015.
	/*
		match := Match{Type: "elimination", DisplayName: "SF1-1", ElimRound: 2, ElimInstance: 1}
		db.CreateMatch(&match)
		match2 := Match{Type: "elimination", DisplayName: "SF2-2", ElimRound: 2, ElimInstance: 2}
		db.CreateMatch(&match2)
		match3 := Match{Type: "elimination", DisplayName: "SF2-1", ElimRound: 2, ElimInstance: 1}
		db.CreateMatch(&match3)
		match4 := Match{Type: "elimination", DisplayName: "QF2-1", ElimRound: 4, ElimInstance: 1}
		db.CreateMatch(&match4)
		match5 := Match{Type: "practice", DisplayName: "1"}
		db.CreateMatch(&match5)

		matches, err := db.GetMatchesByElimRound(4)
		assert.Nil(t, err)
		assert.Empty(t, matches)
		matches, err = db.GetMatchesByElimRound(2)
		assert.Nil(t, err)
		if assert.Equal(t, 2, len(matches)) {
			assert.Equal(t, "SF2-1", matches[0].DisplayName)
			assert.Equal(t, "SF2-2", matches[1].DisplayName)
		}
	*/
}

func TestGetMatchesByType(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match := Match{0, "qualification", "1", time.Now().UTC(), 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	match2 := Match{0, "practice", "1", time.Now().UTC(), 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match2)
	match3 := Match{0, "practice", "2", time.Now().UTC(), 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match3)

	matches, err := db.GetMatchesByType("test")
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByType("practice")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(matches))
	matches, err = db.GetMatchesByType("qualification")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))
}
