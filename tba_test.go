// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
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
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[\"frc254\",\"frc1114\"]", reader.String())
		assert.Equal(t, "my_secret_id", r.Header["X-Tba-Auth-Id"][0])
		assert.Equal(t, "f5c022fde6d1186ea0719fe28ab6cc63", r.Header["X-Tba-Auth-Sig"][0])
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
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[{\"comp_level\":\"qm\",\"set_number\":0,\"match_number\":2,\"alliances\":{\"blue\""+
			":{\"score\":104,\"teams\":[\"frc10\",\"frc11\",\"frc12\"]},\"red\":{\"score\":92,\"teams\":[\"f"+
			"rc7\",\"frc8\",\"frc9\"]}},\"score_breakdown\":{\"blue\":{\"coopertition_points\":40,\"auto_poi"+
			"nts\":10,\"container_points\":24,\"tote_points\":24,\"litter_points\":6,\"foul_points\":0},\"re"+
			"d\":{\"coopertition_points\":40,\"auto_points\":28,\"container_points\":24,\"tote_points\":12,"+
			"\"litter_points\":6,\"foul_points\":18}},\"time_string\":\"4:10 PM\",\"time_utc\":\"1970-01-01T"+
			"00:10:00\"},{\"comp_level\":\"sf\",\"set_number\":2,\"match_number\":2,\"alliances\":{\"blue\":"+
			"{\"score\":null,\"teams\":[\"frc0\",\"frc0\",\"frc0\"]},\"red\":{\"score\":null,\"teams\":[\"fr"+
			"c0\",\"frc0\",\"frc0\"]}},\"score_breakdown\":null,\"time_string\":\"4:00 PM\",\"time_utc\":\"0"+
			"001-01-01T00:00:00\"}]", reader.String())
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
	db.CreateRanking(&Ranking{1114, 2, 20, 1100, 625, 90, 554, 10, 0.254, 0, 10})
	db.CreateRanking(&Ranking{254, 1, 20, 1100, 625, 90, 554, 10, 0.254, 0, 10})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "{\"breakdowns\":[\"QA\",\"Coopertition\",\"Auto\",\"Container\",\"Tote\","+
			"\"Litter\"],\"rankings\":[{\"team_key\":\"frc254\",\"rank\":1,\"QA\":20,\"Coopertition\":1100,"+
			"\"Auto\":625,\"Container\":90,\"Tote\":554,\"Litter\":10,\"dqs\":0,\"played\":10},"+
			"{\"team_key\":\"frc1114\",\"rank\":2,\"QA\":20,\"Coopertition\":1100,\"Auto\":625,"+
			"\"Container\":90,\"Tote\":554,\"Litter\":10,\"dqs\":0,\"played\":10}]}", reader.String())
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
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[[\"frc254\",\"frc469\",\"frc2848\",\"frc74\"],[\"frc1718\",\"frc2451\"]]",
			reader.String())
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
