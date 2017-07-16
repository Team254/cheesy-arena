// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMatchesApi(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	match1 := Match{Type: "qualification", DisplayName: "1", Time: time.Unix(0, 0), Red1: 1, Red2: 2, Red3: 3,
		Blue1: 4, Blue2: 5, Blue3: 6, Blue1IsSurrogate: true, Blue2IsSurrogate: true, Blue3IsSurrogate: true}
	match2 := Match{Type: "qualification", DisplayName: "2", Time: time.Unix(600, 0), Red1: 7, Red2: 8, Red3: 9,
		Blue1: 10, Blue2: 11, Blue3: 12, Red1IsSurrogate: true, Red2IsSurrogate: true, Red3IsSurrogate: true}
	match3 := Match{Type: "practice", DisplayName: "1", Time: time.Now(), Red1: 6, Red2: 5, Red3: 4,
		Blue1: 3, Blue2: 2, Blue3: 1}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	db.CreateMatch(&match3)
	matchResult1 := buildTestMatchResult(match1.Id, 1)
	db.CreateMatchResult(&matchResult1)

	recorder := getHttpResponse("/api/matches/qualification")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/json", recorder.HeaderMap["Content-Type"][0])
	var matchesData []MatchWithResult
	err := json.Unmarshal([]byte(recorder.Body.String()), &matchesData)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matchesData)) {
		assert.Equal(t, match1.Id, matchesData[0].Match.Id)
		assert.Equal(t, matchResult1, matchesData[0].Result.MatchResult)
		assert.Equal(t, match2.Id, matchesData[1].Match.Id)
		assert.Nil(t, matchesData[1].Result)
	}
}

func TestRankingsApi(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	eventSettings, _ = db.GetEventSettings()

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

	ranking1 := RankingWithNickname{Ranking{1114, 2, 18, 700, 625, 90, 554, 9, 0.254, 3, 2, 1, 0, 10}, "Simbots"}
	ranking2 := RankingWithNickname{Ranking{254, 1, 20, 700, 625, 90, 554, 10, 0.254, 1, 2, 3, 0, 10}, "ChezyPof"}
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

func TestSponsorSlides(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)

	slide1 := SponsorSlide{1, "subtitle", "line1", "line2", "image", 2}
	slide2 := SponsorSlide{2, "Chezy Sponsaur", "Teh", "Chezy Pofs", "ejface.jpg", 54}
	db.CreateSponsorSlide(&slide1)
	db.CreateSponsorSlide(&slide2)

	recorder := getHttpResponse("/api/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/json", recorder.HeaderMap["Content-Type"][0])
	var sponsorSlides []SponsorSlide
	err := json.Unmarshal([]byte(recorder.Body.String()), &sponsorSlides)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(sponsorSlides)) {
		assert.Equal(t, slide1, sponsorSlides[0])
		assert.Equal(t, slide2, sponsorSlides[1])
	}
}
