// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessPoint_ConfigureTeamWifi(t *testing.T) {
	var ap AccessPoint
	var request configurationRequest
	wifiStatuses := [6]*TeamWifiStatus{{}, {}, {}, {}, {}, {}}
	ap.SetSettings("dummy", "password1", 123, true, wifiStatuses)
	ap.Status = "INITIAL"

	// Mock the radio API server.
	radioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/configuration")
		assert.Equal(t, "Bearer password1", r.Header.Get("Authorization"))
		assert.Nil(t, json.NewDecoder(r.Body).Decode(&request))
	}))
	ap.apiUrl = radioServer.URL

	// All stations assigned.
	team1 := &model.Team{Id: 254, WpaKey: "11111111"}
	team2 := &model.Team{Id: 1114, WpaKey: "22222222"}
	team3 := &model.Team{Id: 469, WpaKey: "33333333"}
	team4 := &model.Team{Id: 2046, WpaKey: "44444444"}
	team5 := &model.Team{Id: 2056, WpaKey: "55555555"}
	team6 := &model.Team{Id: 1678, WpaKey: "66666666"}
	assert.Nil(t, ap.ConfigureTeamWifi([6]*model.Team{team1, team2, team3, team4, team5, team6}))
	assert.Equal(
		t,
		configurationRequest{
			Channel: 123,
			StationConfigurations: map[string]stationConfiguration{
				"red1":  {"254", "11111111"},
				"red2":  {"1114", "22222222"},
				"red3":  {"469", "33333333"},
				"blue1": {"2046", "44444444"},
				"blue2": {"2056", "55555555"},
				"blue3": {"1678", "66666666"},
			},
		},
		request,
	)

	// Different channel and only some stations assigned.
	ap.channel = 456
	request = configurationRequest{}
	assert.Nil(t, ap.ConfigureTeamWifi([6]*model.Team{nil, nil, team2, nil, team1, nil}))
	assert.Equal(
		t,
		configurationRequest{
			Channel: 456,
			StationConfigurations: map[string]stationConfiguration{
				"red3":  {"1114", "22222222"},
				"blue2": {"254", "11111111"},
			},
		},
		request,
	)

	// Radio API returns an error.
	radioServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/configuration")
		http.Error(w, "oh noes", 507)
	}))
	ap.apiUrl = radioServer.URL
	err := ap.ConfigureTeamWifi([6]*model.Team{team1, team2, team3, team4, team5, team6})
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "returned status 507: oh noes")
	}
	assert.Equal(t, "INITIAL", ap.Status)
}

func TestAccessPoint_updateMonitoring(t *testing.T) {
	var ap AccessPoint
	wifiStatuses := [6]*TeamWifiStatus{{}, {}, {}, {}, {}, {}}
	ap.SetSettings("dummy", "password2", 123, true, wifiStatuses)

	apStatus := accessPointStatus{
		Channel: 456,
		Status:  "ACTIVE",
		StationStatuses: map[string]*stationStatus{
			"red1":  {"254", "hash111", "salt1", true, 1, 2, 3, 4},
			"red2":  {"1114", "hash222", "salt2", false, 5, 6, 7, 8},
			"red3":  {"469", "hash333", "salt3", true, 9, 10, 11, 12},
			"blue1": {"2046", "hash444", "salt4", false, 13, 14, 15, 16},
			"blue2": {"2056", "hash555", "salt5", true, 17, 18, 19, 20},
			"blue3": {"1678", "hash666", "salt6", false, 21, 22, 23, 24},
		},
	}

	// Mock the radio API server.
	radioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/status")
		assert.Equal(t, "Bearer password2", r.Header.Get("Authorization"))
		assert.Nil(t, json.NewEncoder(w).Encode(apStatus))
	}))
	ap.apiUrl = radioServer.URL

	// All stations assigned.
	assert.Nil(t, ap.updateMonitoring())
	assert.Equal(t, 123, ap.channel) // Should not have changed to reflect the radio API.
	assert.Equal(t, "ACTIVE", ap.Status)
	assert.Equal(t, TeamWifiStatus{254, true, 4, 1, 2, 3}, *wifiStatuses[0])
	assert.Equal(t, TeamWifiStatus{1114, false, 8, 5, 6, 7}, *wifiStatuses[1])
	assert.Equal(t, TeamWifiStatus{469, true, 12, 9, 10, 11}, *wifiStatuses[2])
	assert.Equal(t, TeamWifiStatus{2046, false, 16, 13, 14, 15}, *wifiStatuses[3])
	assert.Equal(t, TeamWifiStatus{2056, true, 20, 17, 18, 19}, *wifiStatuses[4])
	assert.Equal(t, TeamWifiStatus{1678, false, 24, 21, 22, 23}, *wifiStatuses[5])

	// Only some stations assigned.
	apStatus.Status = "CONFIGURING"
	apStatus.StationStatuses = map[string]*stationStatus{
		"red1":  nil,
		"red2":  nil,
		"red3":  {"469", "hash333", "salt3", true, 9, 10, 11, 12},
		"blue1": nil,
		"blue2": {"2056", "hash555", "salt5", true, 17, 18, 19, 20},
		"blue3": nil,
	}
	assert.Nil(t, ap.updateMonitoring())
	assert.Equal(t, "CONFIGURING", ap.Status)
	assert.Equal(t, TeamWifiStatus{}, *wifiStatuses[0])
	assert.Equal(t, TeamWifiStatus{}, *wifiStatuses[1])
	assert.Equal(t, TeamWifiStatus{469, true, 12, 9, 10, 11}, *wifiStatuses[2])
	assert.Equal(t, TeamWifiStatus{}, *wifiStatuses[3])
	assert.Equal(t, TeamWifiStatus{2056, true, 20, 17, 18, 19}, *wifiStatuses[4])
	assert.Equal(t, TeamWifiStatus{}, *wifiStatuses[5])

	// Radio API returns an error.
	radioServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/status")
		http.Error(w, "gosh darn", 404)
	}))
	ap.apiUrl = radioServer.URL
	err := ap.updateMonitoring()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "returned status 404: gosh darn")
	}
	assert.Equal(t, "ERROR", ap.Status)
}
