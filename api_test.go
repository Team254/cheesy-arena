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
	assert.Equal(t, "[]", recorder.Body.String())

	ranking1 := Ranking{1114, 2, 18, 1100, 625, 90, 554, 0.254, 9, 1, 0, 0, 10}
	ranking2 := Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}
	db.CreateRanking(&ranking1)
	db.CreateRanking(&ranking2)

	recorder = getHttpResponse("/api/rankings")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/json", recorder.HeaderMap["Content-Type"][0])
	var jsonRankings []Ranking
	err := json.Unmarshal([]byte(recorder.Body.String()), &jsonRankings)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(jsonRankings)) {
		assert.Equal(t, ranking1, jsonRankings[1])
		assert.Equal(t, ranking2, jsonRankings[0])
	}
}
