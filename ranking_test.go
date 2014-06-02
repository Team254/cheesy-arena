// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
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
		db.CreateRanking(&Ranking{TeamId: i})
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
	rand.Seed(0)

	setupMatchResultsForRankings(db)
	err = db.CalculateRankings()
	assert.Nil(t, err)
	rankings, err := db.GetAllRankings()
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, Ranking{4, 1, 4, 130, 456, 100, 309, 0.897169713149801, 2, 0, 0, 0, 2}, rankings[0])
		assert.Equal(t, Ranking{6, 2, 3, 90, 292, 70, 141, 0.16735444255905835, 1, 0, 1, 0, 2}, rankings[1])
		assert.Equal(t, Ranking{5, 3, 3, 130, 456, 100, 219, 0.2885856518054551, 1, 1, 1, 0, 3}, rankings[2])
		assert.Equal(t, Ranking{2, 4, 3, 40, 164, 30, 168, 0.8497802817628735, 1, 0, 1, 1, 3}, rankings[3])
		assert.Equal(t, Ranking{3, 5, 1, 40, 164, 30, 78, 0.9026048462705047, 0, 1, 1, 0, 2}, rankings[4])
		assert.Equal(t, Ranking{1, 6, 1, 80, 328, 60, 156, 0.2730468047134829, 0, 2, 1, 0, 3}, rankings[5])
	}

	// Test after changing a match result.
	matchResult3 := buildTestMatchResult(3, 3)
	matchResult3.RedScore, matchResult3.BlueScore = matchResult3.BlueScore, matchResult3.RedScore
	err = db.CreateMatchResult(&matchResult3)
	assert.Nil(t, err)
	err = db.CalculateRankings()
	assert.Nil(t, err)
	rankings, err = db.GetAllRankings()
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, Ranking{6, 1, 4, 180, 584, 140, 192, 0.24043190328608438, 2, 0, 0, 0, 2}, rankings[0])
		assert.Equal(t, Ranking{4, 2, 4, 130, 456, 100, 309, 0.5102423328818813, 2, 0, 0, 0, 2}, rankings[1])
		assert.Equal(t, Ranking{5, 3, 4, 220, 748, 170, 270, 0.2092018731282357, 2, 1, 0, 0, 3}, rankings[2])
		assert.Equal(t, Ranking{2, 4, 2, 80, 328, 60, 336, 0.4018978925803393, 1, 1, 0, 1, 3}, rankings[3])
		assert.Equal(t, Ranking{3, 5, 0, 80, 328, 60, 246, 0.6930700440076261, 0, 2, 0, 0, 2}, rankings[4])
		assert.Equal(t, Ranking{1, 6, 0, 120, 492, 90, 324, 0.284824110942037, 0, 3, 0, 0, 3}, rankings[5])
	}
}

func TestSortRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	// Check tiebreakers.
	rankings := make(map[int]*Ranking)
	rankings[1] = &Ranking{1, 0, 20, 50, 50, 50, 50, 0.49, 0, 0, 0, 0, 10}
	rankings[2] = &Ranking{2, 0, 20, 50, 50, 50, 50, 0.51, 0, 0, 0, 0, 10}
	rankings[3] = &Ranking{3, 0, 20, 50, 50, 50, 49, 0.50, 0, 0, 0, 0, 10}
	rankings[4] = &Ranking{4, 0, 20, 50, 50, 50, 51, 0.50, 0, 0, 0, 0, 10}
	rankings[5] = &Ranking{5, 0, 20, 50, 50, 49, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[6] = &Ranking{6, 0, 20, 50, 50, 51, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[7] = &Ranking{7, 0, 20, 50, 49, 50, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[8] = &Ranking{8, 0, 20, 50, 51, 50, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[9] = &Ranking{9, 0, 20, 49, 50, 50, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[10] = &Ranking{10, 0, 20, 51, 50, 50, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[11] = &Ranking{11, 0, 19, 50, 50, 50, 50, 0.50, 0, 0, 0, 0, 10}
	rankings[12] = &Ranking{12, 0, 21, 50, 50, 50, 50, 0.50, 0, 0, 0, 0, 10}
	sortedRankings := sortRankings(rankings)
	assert.Equal(t, 12, sortedRankings[0].TeamId)
	assert.Equal(t, 10, sortedRankings[1].TeamId)
	assert.Equal(t, 8, sortedRankings[2].TeamId)
	assert.Equal(t, 6, sortedRankings[3].TeamId)
	assert.Equal(t, 4, sortedRankings[4].TeamId)
	assert.Equal(t, 2, sortedRankings[5].TeamId)
	assert.Equal(t, 1, sortedRankings[6].TeamId)
	assert.Equal(t, 3, sortedRankings[7].TeamId)
	assert.Equal(t, 5, sortedRankings[8].TeamId)
	assert.Equal(t, 7, sortedRankings[9].TeamId)
	assert.Equal(t, 9, sortedRankings[10].TeamId)
	assert.Equal(t, 11, sortedRankings[11].TeamId)

	// Check with unequal number of matches played.
	rankings = make(map[int]*Ranking)
	rankings[1] = &Ranking{1, 0, 10, 25, 25, 25, 25, 0.49, 0, 0, 0, 0, 5}
	rankings[2] = &Ranking{2, 0, 19, 50, 50, 50, 50, 0.51, 0, 0, 0, 0, 9}
	rankings[3] = &Ranking{3, 0, 20, 50, 50, 50, 50, 0.51, 0, 0, 0, 0, 10}
	sortedRankings = sortRankings(rankings)
	assert.Equal(t, 2, sortedRankings[0].TeamId)
	assert.Equal(t, 3, sortedRankings[1].TeamId)
	assert.Equal(t, 1, sortedRankings[2].TeamId)
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
	matchResult1.Cards = Cards{[]int{}, []int{2}}
	db.CreateMatchResult(&matchResult1)

	match2 := Match{Type: "qualification", DisplayName: "2", Red1: 1, Red2: 3, Red3: 5, Blue1: 2, Blue2: 4,
		Blue3: 6, Status: "complete", Red2IsSurrogate: true, Blue3IsSurrogate: true}
	db.CreateMatch(&match2)
	matchResult2 := buildTestMatchResult(match2.Id, 1)
	matchResult2.BlueScore = matchResult2.RedScore
	db.CreateMatchResult(&matchResult2)

	match3 := Match{Type: "qualification", DisplayName: "3", Red1: 6, Red2: 5, Red3: 4, Blue1: 3, Blue2: 2,
		Blue3: 1, Status: "complete", Red3IsSurrogate: true}
	db.CreateMatch(&match3)
	matchResult3 := buildTestMatchResult(match3.Id, 1)
	db.CreateMatchResult(&matchResult3)
	matchResult3 = MatchResult{MatchId: match3.Id, PlayNumber: 2}
	db.CreateMatchResult(&matchResult3)

	match4 := Match{Type: "practice", DisplayName: "1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: "complete"}
	db.CreateMatch(&match4)
	matchResult4 := buildTestMatchResult(match4.Id, 1)
	matchResult4.Cards = Cards{[]int{}, []int{}}
	db.CreateMatchResult(&matchResult4)

	match5 := Match{Type: "elimination", DisplayName: "F-1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: "complete"}
	db.CreateMatch(&match5)
	matchResult5 := buildTestMatchResult(match5.Id, 1)
	matchResult5.Cards = Cards{[]int{}, []int{}}
	db.CreateMatchResult(&matchResult5)

	match6 := Match{Type: "qualification", DisplayName: "4", Red1: 7, Red2: 8, Red3: 9, Blue1: 10, Blue2: 11,
		Blue3: 12, Status: ""}
	db.CreateMatch(&match6)
	matchResult6 := buildTestMatchResult(match6.Id, 1)
	matchResult6.Cards = Cards{[]int{}, []int{}}
	db.CreateMatchResult(&matchResult6)
}
