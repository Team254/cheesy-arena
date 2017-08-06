// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"math/rand"
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
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	ranking := game.TestRanking1()
	db.CreateRanking(ranking)
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
		db.CreateRanking(&game.Ranking{TeamId: i})
	}
	rankings, err = db.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, numRankings, len(rankings))
	for i := 0; i < numRankings; i++ {
		assert.Equal(t, i+1, rankings[i].TeamId)
	}
}

func TestCalculateRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	rand.Seed(0)

	setupMatchResultsForRankings(db)
	err = db.CalculateRankings()
	assert.Nil(t, err)
	rankings, err := db.GetAllRankings()
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, 4, rankings[0].TeamId)
		assert.Equal(t, 5, rankings[1].TeamId)
		assert.Equal(t, 6, rankings[2].TeamId)
		assert.Equal(t, 1, rankings[3].TeamId)
		assert.Equal(t, 3, rankings[4].TeamId)
		assert.Equal(t, 2, rankings[5].TeamId)
	}

	// Test after changing a match result.
	matchResult3 := buildTestMatchResult(3, 3)
	matchResult3.RedScore, matchResult3.BlueScore = matchResult3.BlueScore, matchResult3.RedScore
	err = db.CreateMatchResult(matchResult3)
	assert.Nil(t, err)
	err = db.CalculateRankings()
	assert.Nil(t, err)
	rankings, err = db.GetAllRankings()
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, 6, rankings[0].TeamId)
		assert.Equal(t, 5, rankings[1].TeamId)
		assert.Equal(t, 4, rankings[2].TeamId)
		assert.Equal(t, 1, rankings[3].TeamId)
		assert.Equal(t, 3, rankings[4].TeamId)
		assert.Equal(t, 2, rankings[5].TeamId)
	}
}

func BenchmarkCalculateRankings(b *testing.B) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(b, err)
	defer db.Close()
	setupMatchResultsForRankings(db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.CalculateRankings()
	}
}

// Sets up a schedule and results that touches on all possible variables.
func setupMatchResultsForRankings(db *Database) {
	match1 := Match{Type: "qualification", DisplayName: "1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: "complete"}
	db.CreateMatch(&match1)
	matchResult1 := buildTestMatchResult(match1.Id, 1)
	matchResult1.RedCards = map[string]string{"2": "red"}
	db.CreateMatchResult(matchResult1)

	match2 := Match{Type: "qualification", DisplayName: "2", Red1: 1, Red2: 3, Red3: 5, Blue1: 2, Blue2: 4,
		Blue3: 6, Status: "complete", Red2IsSurrogate: true, Blue3IsSurrogate: true}
	db.CreateMatch(&match2)
	matchResult2 := buildTestMatchResult(match2.Id, 1)
	matchResult2.BlueScore = matchResult2.RedScore
	db.CreateMatchResult(matchResult2)

	match3 := Match{Type: "qualification", DisplayName: "3", Red1: 6, Red2: 5, Red3: 4, Blue1: 3, Blue2: 2,
		Blue3: 1, Status: "complete", Red3IsSurrogate: true}
	db.CreateMatch(&match3)
	matchResult3 := buildTestMatchResult(match3.Id, 1)
	db.CreateMatchResult(matchResult3)
	matchResult3 = NewMatchResult()
	matchResult3.MatchId = match3.Id
	matchResult3.PlayNumber = 2
	db.CreateMatchResult(matchResult3)

	match4 := Match{Type: "practice", DisplayName: "1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: "complete"}
	db.CreateMatch(&match4)
	matchResult4 := buildTestMatchResult(match4.Id, 1)
	db.CreateMatchResult(matchResult4)

	match5 := Match{Type: "elimination", DisplayName: "F-1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: "complete"}
	db.CreateMatch(&match5)
	matchResult5 := buildTestMatchResult(match5.Id, 1)
	db.CreateMatchResult(matchResult5)

	match6 := Match{Type: "qualification", DisplayName: "4", Red1: 7, Red2: 8, Red3: 9, Blue1: 10, Blue2: 11,
		Blue3: 12, Status: ""}
	db.CreateMatch(&match6)
	matchResult6 := buildTestMatchResult(match6.Id, 1)
	db.CreateMatchResult(matchResult6)
}
