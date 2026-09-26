// Copyright 2026 Team 254. All Rights Reserved.
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
	"strings"
	"testing"
	"time"
)

func TestPublishTeams(t *testing.T) {
	database := setupTestDb(t)

	database.CreateTeam(&model.Team{Id: 254})
	database.CreateTeam(&model.Team{Id: 1114})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.String(), "event/my_event_code")
				var reader bytes.Buffer
				reader.ReadFrom(r.Body)
				assert.Equal(t, "[\"frc254\",\"frc1114\"]", reader.String())
				assert.Equal(t, "my_secret_id", r.Header["X-Tba-Auth-Id"][0])
				assert.Equal(t, "f5c022fde6d1186ea0719fe28ab6cc63", r.Header["X-Tba-Auth-Sig"][0])
			},
		),
	)
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishTeams(database))
}

func TestPublishMatches(t *testing.T) {
	database := setupTestDb(t)

	match1 := model.Match{
		Type:             model.Qualification,
		ShortName:        "Q2",
		Time:             time.Unix(600, 0),
		StartedAt:        time.Unix(620, 0),
		ScoreCommittedAt: time.Unix(780, 0),
		Red1:             7,
		Red2:             8,
		Red3:             9,
		Blue1:            10,
		Blue2:            11,
		Blue3:            12,
		Status:           game.RedWonMatch,
		TbaMatchKey:      model.TbaMatchKey{"qm", 0, 2},
	}
	match2 := model.Match{Type: model.Playoff, ShortName: "SF2-2", TbaMatchKey: model.TbaMatchKey{"omg", 5, 29}}
	database.CreateMatch(&match1)
	database.CreateMatch(&match2)
	matchResult1 := model.BuildTestMatchResult(match1.Id, 1)
	database.CreateMatchResult(matchResult1)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var matches []*TbaMatch
				json.Unmarshal(body, &matches)
				assert.Equal(t, 2, len(matches))
				assert.Equal(t, "qm", matches[0].CompLevel)
				assert.Equal(t, 0, matches[0].SetNumber)
				assert.Equal(t, 2, matches[0].MatchNumber)
				assert.Equal(t, time.Unix(620, 0).UTC().Format("2006-01-02T15:04:05"), matches[0].ActualStartTimeUtc)
				assert.Equal(t, time.Unix(780, 0).UTC().Format("2006-01-02T15:04:05"), matches[0].PostResultsTimeUtc)
				assert.Equal(t, "omg", matches[1].CompLevel)
				assert.Equal(t, 5, matches[1].SetNumber)
				assert.Equal(t, 29, matches[1].MatchNumber)
				assert.Equal(t, "", matches[1].ActualStartTimeUtc)
				assert.Equal(t, "", matches[1].PostResultsTimeUtc)
			},
		),
	)
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
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var response TbaRankings
				json.Unmarshal(body, &response)
				assert.Equal(t, 2, len(response.Rankings))
				assert.Equal(t, "frc254", response.Rankings[0].TeamKey)
				assert.Equal(t, "frc1114", response.Rankings[1].TeamKey)
			},
		),
	)
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishRankings(database))
}

func TestPublishAlliances(t *testing.T) {
	database := setupTestDb(t)

	model.BuildTestAlliances(database)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var reader bytes.Buffer
				reader.ReadFrom(r.Body)
				if strings.Contains(r.URL.String(), "alliance_selections") {
					assert.Equal(
						t,
						"[[\"frc254\",\"frc469\",\"frc2848\",\"frc74\",\"frc3175\"],[\"frc1718\",\"frc2451\",\"frc1619\"]]",
						reader.String(),
					)
				} else {
					assert.Equal(t, "{\"playoff_type\":10}", reader.String())
				}
			},
		),
	)
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishAlliances(database))
}

func TestPublishingErrors(t *testing.T) {
	database := setupTestDb(t)

	model.BuildTestAlliances(database)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "oh noes", 500)
			},
		),
	)
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.NotNil(t, client.PublishTeams(database))
	assert.NotNil(t, client.PublishMatches(database))
	assert.NotNil(t, client.PublishRankings(database))
	assert.NotNil(t, client.PublishAlliances(database))
}

func TestCheckTbaPostResponseClosesBody(t *testing.T) {
	body := &closeTrackingBody{Reader: strings.NewReader("ok")}
	err := checkTbaPostResponse(&http.Response{StatusCode: 200, Body: body})
	assert.Nil(t, err)
	assert.True(t, body.closed)

	body = &closeTrackingBody{Reader: strings.NewReader("oh noes")}
	err = checkTbaPostResponse(&http.Response{StatusCode: 500, Body: body})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Got status code 500 from TBA: oh noes")
	assert.True(t, body.closed)
}

func TestPublishAwards(t *testing.T) {
	database := setupTestDb(t)

	database.CreateAward(&model.Award{0, model.JudgedAward, "Saftey Award", 254, ""})
	database.CreateAward(&model.Award{0, model.JudgedAward, "Spirt Award", 0, "Bob Dorough"})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.String(), "event/my_event_code")
				var reader bytes.Buffer
				reader.ReadFrom(r.Body)
				var actual []TbaPublishedAward
				err := json.Unmarshal(reader.Bytes(), &actual)
				assert.Nil(t, err)

				expected := []TbaPublishedAward{
					{Name: "Saftey Award", TeamKey: stringPtr("frc254"), Awardee: nil},
					{Name: "Spirt Award", TeamKey: nil, Awardee: stringPtr("Bob Dorough")},
				}
				assert.Equal(t, expected, actual)
			},
		),
	)
	defer tbaServer.Close()
	client := NewTbaClient("my_event_code", "my_secret_id", "my_secret")
	client.BaseUrl = tbaServer.URL

	assert.Nil(t, client.PublishAwards(database))
}

func setupTestDb(t *testing.T) *model.Database {
	return model.SetupTestDb(t)
}

type closeTrackingBody struct {
	*strings.Reader
	closed bool
}

func (body *closeTrackingBody) Close() error {
	body.closed = true
	return nil
}

func TestPublishPlayoffAllianceCards(t *testing.T) {
	database := setupTestDb(t)
	match := &model.Match{Type: model.Playoff, Status: game.BlueWonMatch, Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6}
	assert.NoError(t, database.CreateMatch(match))
	result := model.NewMatchResult()
	result.MatchId, result.PlayNumber = match.Id, 1
	result.PlayoffRedAllianceCard = "red"
	result.PlayoffBlueAllianceCard = "dq"
	result.BlueCards = map[string]string{"4": "red"} // Ignored during playoffs.
	result.CorrectPlayoffScore()
	assert.NoError(t, database.CreateMatchResult(result))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var matches []TbaMatch
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&matches))
		if assert.Len(t, matches, 1) {
			assert.Equal(t, []string{"frc1", "frc2", "frc3"}, matches[0].Alliances["red"].Dqs)
			assert.Empty(t, matches[0].Alliances["blue"].Dqs)
		}
	}))
	defer server.Close()
	client := NewTbaClient("test", "", "")
	client.BaseUrl = server.URL
	assert.NoError(t, client.PublishMatches(database))
}
