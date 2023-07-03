// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package partner

import (
	"bytes"
	"encoding/json"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishTeams(t *testing.T) {
	database := setupTestDb(t)

	database.CreateTeam(&model.Team{Id: 254})
	database.CreateTeam(&model.Team{Id: 1114})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "event/my_event_code")
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[\"frc254\",\"frc1114\"]", reader.String())
		assert.Equal(t, "my_secret_id", r.Header["X-Tba-Auth-Id"][0])
		assert.Equal(t, "f5c022fde6d1186ea0719fe28ab6cc63", r.Header["X-Tba-Auth-Sig"][0])
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishTeams(database))
}

func TestPublishMatches(t *testing.T) {
	database := setupTestDb(t)

	match1 := model.Match{
		Type:        model.Qualification,
		ShortName:   "Q2",
		Time:        time.Unix(600, 0),
		Red1:        7,
		Red2:        8,
		Red3:        9,
		Blue1:       10,
		Blue2:       11,
		Blue3:       12,
		Status:      game.RedWonMatch,
		TbaMatchKey: model.TbaMatchKey{"qm", 0, 2},
	}
	match2 := model.Match{Type: model.Playoff, ShortName: "SF2-2", TbaMatchKey: model.TbaMatchKey{"omg", 5, 29}}
	database.CreateMatch(&match1)
	database.CreateMatch(&match2)
	matchResult1 := model.BuildTestMatchResult(match1.Id, 1)
	database.CreateMatchResult(matchResult1)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var matches []*TbaMatch
		json.Unmarshal(body, &matches)
		assert.Equal(t, 2, len(matches))
		assert.Equal(t, "qm", matches[0].CompLevel)
		assert.Equal(t, 0, matches[0].SetNumber)
		assert.Equal(t, 2, matches[0].MatchNumber)
		assert.Equal(t, "omg", matches[1].CompLevel)
		assert.Equal(t, 5, matches[1].SetNumber)
		assert.Equal(t, 29, matches[1].MatchNumber)
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishMatches(database))
}

func TestPublishRankings(t *testing.T) {
	database := setupTestDb(t)

	database.CreateRanking(game.TestRanking2())
	database.CreateRanking(game.TestRanking1())

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var response TbaRankings
		json.Unmarshal(body, &response)
		assert.Equal(t, 2, len(response.Rankings))
		assert.Equal(t, "frc254", response.Rankings[0].TeamKey)
		assert.Equal(t, "frc1114", response.Rankings[1].TeamKey)
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishRankings(database))
}

func TestPublishAlliances(t *testing.T) {
	database := setupTestDb(t)

	model.BuildTestAlliances(database)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(
			t,
			"[[\"frc254\",\"frc469\",\"frc2848\",\"frc74\",\"frc3175\"],[\"frc1718\",\"frc2451\",\"frc1619\"]]",
			reader.String(),
		)
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishAlliances(database))
}

func TestPublishingErrors(t *testing.T) {
	database := setupTestDb(t)

	model.BuildTestAlliances(database)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oh noes", 500)
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.NotNil(t, client.PublishTeams(database))
	assert.NotNil(t, client.PublishMatches(database))
	assert.NotNil(t, client.PublishRankings(database))
	assert.NotNil(t, client.PublishAlliances(database))
}

func TestPublishAwards(t *testing.T) {
	database := setupTestDb(t)

	database.CreateAward(&model.Award{0, model.JudgedAward, "Saftey Award", 254, ""})
	database.CreateAward(&model.Award{0, model.JudgedAward, "Spirt Award", 0, "Bob Dorough"})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "event/my_event_code")
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[{\"name_str\":\"Saftey Award\",\"team_key\":\"frc254\",\"awardee\":\"\"},"+
			"{\"name_str\":\"Spirt Award\",\"team_key\":\"frc0\",\"awardee\":\"Bob Dorough\"}]", reader.String())
	}))
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishAwards(database))
}

func setupTestDb(t *testing.T) *model.Database {
	return model.SetupTestDb(t, "partner")
}
