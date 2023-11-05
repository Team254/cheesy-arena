// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package partner

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLineup(t *testing.T) {
	// Mock the Nexus server.
	nexusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "/v1/my_event_code/")
		if strings.Contains(r.URL.String(), "/v1/my_event_code/p1/lineup") {
			w.Write([]byte("{\"red\":[\"101\",\"102\",\"103\"],\"blue\":[\"104\",\"105\",\"106\"]}"))
		} else if strings.Contains(r.URL.String(), "/v1/my_event_code/p2/lineup") {
			w.Write([]byte("{\"blue\":[\"104\",\"105\",\"106\"]}"))
		} else if strings.Contains(r.URL.String(), "/v1/my_event_code/p3/lineup") {
			w.Write([]byte("{}"))
		} else {
			http.Error(w, "Match not found", 404)
		}
	}))
	defer nexusServer.Close()
	client := NewNexusClient("my_event_code")
	client.BaseUrl = nexusServer.URL

	tbaMatchKey := model.TbaMatchKey{CompLevel: "p", SetNumber: 0, MatchNumber: 1}
	lineup, err := client.GetLineup(tbaMatchKey)
	if assert.Nil(t, err) {
		assert.Equal(t, [6]int{101, 102, 103, 104, 105, 106}, *lineup)
	}

	tbaMatchKey = model.TbaMatchKey{CompLevel: "sf", SetNumber: 6, MatchNumber: 1}
	lineup, err = client.GetLineup(tbaMatchKey)
	assert.Nil(t, lineup)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Match not found")
	}

	tbaMatchKey = model.TbaMatchKey{CompLevel: "p", SetNumber: 0, MatchNumber: 2}
	lineup, err = client.GetLineup(tbaMatchKey)
	assert.Nil(t, lineup)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Lineup not yet submitted")
	}

	tbaMatchKey = model.TbaMatchKey{CompLevel: "p", SetNumber: 0, MatchNumber: 3}
	lineup, err = client.GetLineup(tbaMatchKey)
	assert.Nil(t, lineup)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Lineup not yet submitted")
	}
}
