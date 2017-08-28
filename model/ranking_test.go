// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentRanking(t *testing.T) {
	db := setupTestDb(t)

	ranking, err := db.GetRankingForTeam(1114)
	assert.Nil(t, err)
	assert.Nil(t, ranking)
}

func TestRankingCrud(t *testing.T) {
	db := setupTestDb(t)

	ranking := game.TestRanking1()
	db.CreateRanking(ranking)
	ranking2, err := db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Equal(t, ranking, ranking2)

	ranking.Random = 0.1114
	db.SaveRanking(ranking)
	ranking2, err = db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Equal(t, ranking.Random, ranking2.Random)

	db.DeleteRanking(ranking)
	ranking2, err = db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Nil(t, ranking2)
}

func TestTruncateRankings(t *testing.T) {
	db := setupTestDb(t)

	ranking := game.TestRanking1()
	db.CreateRanking(ranking)
	db.TruncateRankings()
	ranking2, err := db.GetRankingForTeam(254)
	assert.Nil(t, err)
	assert.Nil(t, ranking2)
}

func TestGetAllRankings(t *testing.T) {
	db := setupTestDb(t)

	rankings, err := db.GetAllRankings()
	assert.Nil(t, err)
	assert.Empty(t, rankings)

	numRankings := 20
	for i := 1; i <= numRankings; i++ {
		db.CreateRanking(&game.Ranking{TeamId: i})
	}
	rankings, err = db.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, numRankings, len(rankings))
	for i := 0; i < numRankings; i++ {
		assert.Equal(t, i+1, rankings[i].TeamId)
	}
}
