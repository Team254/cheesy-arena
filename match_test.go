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

	match := Match{254, "qualification", "254", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	match2, err := db.GetMatchById(254)
	assert.Nil(t, err)
	assert.Equal(t, match, *match2)

	match.Status = "started"
	db.SaveMatch(&match)
	match2, err = db.GetMatchById(254)
	assert.Nil(t, err)
	assert.Equal(t, match.Status, match2.Status)

	db.DeleteMatch(&match)
	match2, err = db.GetMatchById(254)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestTruncateMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match := Match{254, "qualification", "254", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	db.TruncateMatches()
	match2, err := db.GetMatchById(254)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestGetMatchesByType(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	match := Match{1, "qualification", "1", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	match2 := Match{2, "practice", "1", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match2)
	match3 := Match{3, "practice", "2", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
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
