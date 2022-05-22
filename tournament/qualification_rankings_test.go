// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateRankings(t *testing.T) {
	database := setupTestDb(t)

	setupMatchResultsForRankings(database)
	updatedRankings, err := CalculateRankings(database, false)
	assert.Nil(t, err)
	rankings, err := database.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, updatedRankings, rankings)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, 4, rankings[0].TeamId)
		assert.Equal(t, 0, rankings[0].PreviousRank)
		assert.Equal(t, 5, rankings[1].TeamId)
		assert.Equal(t, 0, rankings[1].PreviousRank)
		assert.Equal(t, 6, rankings[2].TeamId)
		assert.Equal(t, 0, rankings[2].PreviousRank)
		assert.Equal(t, 1, rankings[3].TeamId)
		assert.Equal(t, 0, rankings[3].PreviousRank)
		assert.Equal(t, 3, rankings[4].TeamId)
		assert.Equal(t, 0, rankings[4].PreviousRank)
		assert.Equal(t, 2, rankings[5].TeamId)
		assert.Equal(t, 0, rankings[5].PreviousRank)
	}

	for _, ranking := range rankings {
		fmt.Printf("%+v\n", ranking)
	}
	fmt.Println()

	// Test after changing a match result.
	matchResult3 := model.BuildTestMatchResult(3, 3)
	matchResult3.RedScore, matchResult3.BlueScore = matchResult3.BlueScore, matchResult3.RedScore
	err = database.CreateMatchResult(matchResult3)
	assert.Nil(t, err)
	updatedRankings, err = CalculateRankings(database, false)
	assert.Nil(t, err)
	rankings, err = database.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, updatedRankings, rankings)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, 6, rankings[0].TeamId)
		assert.Equal(t, 3, rankings[0].PreviousRank)
		assert.Equal(t, 5, rankings[1].TeamId)
		assert.Equal(t, 2, rankings[1].PreviousRank)
		assert.Equal(t, 4, rankings[2].TeamId)
		assert.Equal(t, 1, rankings[2].PreviousRank)
		assert.Equal(t, 1, rankings[3].TeamId)
		assert.Equal(t, 4, rankings[3].PreviousRank)
		assert.Equal(t, 3, rankings[4].TeamId)
		assert.Equal(t, 5, rankings[4].PreviousRank)
		assert.Equal(t, 2, rankings[5].TeamId)
		assert.Equal(t, 6, rankings[5].PreviousRank)
	}

	for _, ranking := range rankings {
		fmt.Printf("%+v\n", ranking)
	}
	fmt.Println()

	matchResult3 = model.BuildTestMatchResult(3, 4)
	err = database.CreateMatchResult(matchResult3)
	assert.Nil(t, err)
	updatedRankings, err = CalculateRankings(database, true)
	assert.Nil(t, err)
	rankings, err = database.GetAllRankings()
	assert.Nil(t, err)
	assert.Equal(t, updatedRankings, rankings)
	if assert.Equal(t, 6, len(rankings)) {
		assert.Equal(t, 4, rankings[0].TeamId)
		assert.Equal(t, 1, rankings[0].PreviousRank)
		assert.Equal(t, 5, rankings[1].TeamId)
		assert.Equal(t, 2, rankings[1].PreviousRank)
		assert.Equal(t, 1, rankings[2].TeamId)
		assert.Equal(t, 4, rankings[2].PreviousRank)
		assert.Equal(t, 3, rankings[3].TeamId)
		assert.Equal(t, 5, rankings[3].PreviousRank)
		assert.Equal(t, 6, rankings[4].TeamId)
		assert.Equal(t, 3, rankings[4].PreviousRank)
		assert.Equal(t, 2, rankings[5].TeamId)
		assert.Equal(t, 6, rankings[5].PreviousRank)
	}

	for _, ranking := range rankings {
		fmt.Printf("%+v\n", ranking)
	}
	fmt.Println()

}

// Sets up a schedule and results that touches on all possible variables.
func setupMatchResultsForRankings(database *model.Database) {
	match1 := model.Match{Type: "qualification", DisplayName: "1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: model.RedWonMatch}
	database.CreateMatch(&match1)
	matchResult1 := model.BuildTestMatchResult(match1.Id, 1)
	matchResult1.RedCards = map[string]string{"2": "red"}
	database.CreateMatchResult(matchResult1)

	match2 := model.Match{Type: "qualification", DisplayName: "2", Red1: 1, Red2: 3, Red3: 5, Blue1: 2, Blue2: 4,
		Blue3: 6, Status: model.BlueWonMatch, Red2IsSurrogate: true, Blue3IsSurrogate: true}
	database.CreateMatch(&match2)
	matchResult2 := model.BuildTestMatchResult(match2.Id, 1)
	matchResult2.BlueScore = matchResult2.RedScore
	database.CreateMatchResult(matchResult2)

	match3 := model.Match{Type: "qualification", DisplayName: "3", Red1: 6, Red2: 5, Red3: 4, Blue1: 3, Blue2: 2,
		Blue3: 1, Status: model.TieMatch, Red3IsSurrogate: true}
	database.CreateMatch(&match3)
	matchResult3 := model.BuildTestMatchResult(match3.Id, 1)
	database.CreateMatchResult(matchResult3)
	matchResult3 = model.NewMatchResult()
	matchResult3.MatchId = match3.Id
	matchResult3.PlayNumber = 2
	database.CreateMatchResult(matchResult3)

	match4 := model.Match{Type: "practice", DisplayName: "1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: model.RedWonMatch}
	database.CreateMatch(&match4)
	matchResult4 := model.BuildTestMatchResult(match4.Id, 1)
	database.CreateMatchResult(matchResult4)

	match5 := model.Match{Type: "elimination", DisplayName: "F-1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5,
		Blue3: 6, Status: model.BlueWonMatch}
	database.CreateMatch(&match5)
	matchResult5 := model.BuildTestMatchResult(match5.Id, 1)
	database.CreateMatchResult(matchResult5)

	match6 := model.Match{Type: "qualification", DisplayName: "4", Red1: 7, Red2: 8, Red3: 9, Blue1: 10, Blue2: 11,
		Blue3: 12, Status: model.MatchNotPlayed}
	database.CreateMatch(&match6)
	matchResult6 := model.BuildTestMatchResult(match6.Id, 1)
	database.CreateMatchResult(matchResult6)
}
