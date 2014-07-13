// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRankingsApi(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)

	// Test that empty rankings produces an empty array.
	recorder := getHttpResponse("/api/rankings")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/json", recorder.HeaderMap["Content-Type"][0])
	rankingsData := struct {
		Rankings           []RankingWithNickname
		TeamNicknames      map[string]string
		HighestPlayedMatch string
	}{}
	err := json.Unmarshal([]byte(recorder.Body.String()), &rankingsData)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rankingsData.Rankings))
	assert.Equal(t, "", rankingsData.HighestPlayedMatch)

	ranking1 := RankingWithNickname{Ranking{1114, 2, 18, 1100, 625, 90, 554, 0.254, 9, 1, 0, 0, 10}, "Simbots"}
	ranking2 := RankingWithNickname{Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}, "ChezyPof"}
	db.CreateRanking(&ranking1.Ranking)
	db.CreateRanking(&ranking2.Ranking)
	db.CreateMatch(&Match{Type: "qualification", DisplayName: "29", Status: "complete"})
	db.CreateMatch(&Match{Type: "qualification", DisplayName: "30"})
	db.CreateTeam(&Team{Id: 254, Nickname: "ChezyPof"})
	db.CreateTeam(&Team{Id: 1114, Nickname: "Simbots"})

	recorder = getHttpResponse("/api/rankings")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/json", recorder.HeaderMap["Content-Type"][0])
	err = json.Unmarshal([]byte(recorder.Body.String()), &rankingsData)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(rankingsData.Rankings)) {
		assert.Equal(t, ranking1, rankingsData.Rankings[1])
		assert.Equal(t, ranking2, rankingsData.Rankings[0])
	}
	assert.Equal(t, "29", rankingsData.HighestPlayedMatch)
}
