// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentRanking(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	ranking, err := db.GetRankingForTeam(1114)
	assert.Nil(t, err)
	assert.Nil(t, ranking)
}

func TestRankingCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	ranking := Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}
	db.CreateRanking(&ranking)
	ranking2, err := db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Equal(t, ranking, *ranking2)

	ranking.Random = 0.1114
	db.SaveRanking(&ranking)
	ranking2, err = db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Equal(t, ranking.Random, ranking2.Random)

	db.DeleteRanking(&ranking)
	ranking2, err = db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Nil(t, ranking2)
}

func TestTruncateRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	ranking := Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}
	db.CreateRanking(&ranking)
	db.TruncateRankings()
	ranking2, err := db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Nil(t, ranking2)
}

func TestGetAllRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	rankings, err := db.GetAllRankings()
	assert.Nil(t, err)
	assert.Empty(t, rankings)

	numRankings := 20
	for i := 1; i <= numRankings; i++ {
		db.CreateRanking(&Ranking{TeamId:i})
	}
	rankings, err = db.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, numRankings, len(rankings))
	for i := 0; i < numRankings; i++ {
		assert.Equal(t, i+1, rankings[i].TeamId)
	}
}
