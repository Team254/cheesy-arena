// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	eventSettings.TbaEventCode = "my_event_code"
	eventSettings.TbaSecretId = "my_secret_id"
	eventSettings.TbaSecret = "my_secret"
	db.CreateTeam(&Team{Id: 254})
	db.CreateTeam(&Team{Id: 1114})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "event/my_event_code")
		assert.Equal(t, "my_secret_id", r.PostFormValue("secret-id"))
		assert.Equal(t, "my_secret", r.PostFormValue("secret"))
		assert.Equal(t, "[\"frc254\",\"frc1114\"]", r.PostFormValue("team_list"))
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishTeams())
}

func TestPublishMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	match1 := Match{Type: "qualification", DisplayName: "2", Time: time.Unix(600, 0), Red1: 7, Red2: 8, Red3: 9,
		Blue1: 10, Blue2: 11, Blue3: 12, Status: "complete"}
	match2 := Match{Type: "elimination", DisplayName: "SF2-2", ElimRound: 2, ElimGroup: 2, ElimInstance: 2}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	matchResult1 := buildTestMatchResult(match1.Id, 1)
	db.CreateMatchResult(&matchResult1)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "[{\"comp_level\":\"qm\",\"set_number\":0,\"match_number\":2,\"alliances\":{\"blue\":"+
			"{\"score\":593,\"teams\":[\"frc10\",\"frc11\",\"frc12\"]},\"red\":{\"score\":312,\"teams\":[\"frc7\""+
			",\"frc8\",\"frc9\"]}},\"time_string\":\"4:10 PM\"},{\"comp_level\":\"sf\",\"set_number\":2,"+
			"\"match_number\":2,\"alliances\":{\"blue\":{\"score\":null,\"teams\":[\"frc0\",\"frc0\",\"frc0\"]},"+
			"\"red\":{\"score\":null,\"teams\":[\"frc0\",\"frc0\",\"frc0\"]}},\"time_string\":\"5:00 PM\"}]",
			r.PostFormValue("matches"))
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishMatches())
}

func TestPublishRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	db.CreateRanking(&Ranking{1114, 2, 18, 1100, 625, 90, 554, 0.254, 9, 1, 0, 0, 10})
	db.CreateRanking(&Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "{\"breakdowns\":[\"QS\",\"Assist\",\"Auto\",\"T\\u0026C\",\"G\\u0026F\"],\"rankings\":"+
			"[{\"team_key\":\"frc254\",\"rank\":1,\"wins\":10,\"losses\":0,\"ties\":0,\"played\":10,\"dqs\":0,"+
			"\"QS\":20,\"Assist\":1100,\"Auto\":625,\"T\\u0026C\":90,\"G\\u0026F\":554},{\"team_key\":\"frc1114\""+
			",\"rank\":2,\"wins\":9,\"losses\":1,\"ties\":0,\"played\":10,\"dqs\":0,\"QS\":18,\"Assist\":1100,"+
			"\"Auto\":625,\"T\\u0026C\":90,\"G\\u0026F\":554}]}", r.PostFormValue("rankings"))
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishRankings())
}

func TestPublishAlliances(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	buildTestAlliances(db)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "[[\"frc254\",\"frc469\",\"frc2848\",\"frc74\"],[\"frc1718\",\"frc2451\"]]",
			r.PostFormValue("alliances"))
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishAlliances())
}

func TestPublishingErrors(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	buildTestAlliances(db)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oh noes", 500)
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.NotNil(t, PublishTeams())
	assert.NotNil(t, PublishMatches())
	assert.NotNil(t, PublishRankings())
	assert.NotNil(t, PublishAlliances())
}
