// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"regexp"
	"testing"
)

func TestConfigureAccessPoint(t *testing.T) {
	model.BaseDir = ".."

	disabledRe := regexp.MustCompile("disabled='([-\\w ]+)'")
	ssidRe := regexp.MustCompile("ssid='([-\\w ]*)'")
	wpaKeyRe := regexp.MustCompile("key='([-\\w ]*)'")

	// Should put dummy values for all team SSIDs if there are no teams.
	config, _ := generateAccessPointConfig([6]*model.Team{nil, nil, nil, nil, nil, nil})
	disableds := disabledRe.FindAllStringSubmatch(config, -1)
	ssids := ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
	if assert.Equal(t, 6, len(disableds)) && assert.Equal(t, 6, len(ssids)) && assert.Equal(t, 6, len(wpaKeys)) {
		for i := 0; i < 6; i++ {
			assert.Equal(t, "0", disableds[i][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", i+1), ssids[i][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", i+1), wpaKeys[i][1])
		}
	}

	// Should configure two SSIDs for two teams and put dummy values for the rest.
	config, _ = generateAccessPointConfig([6]*model.Team{{Id: 254, WpaKey: "aaaaaaaa"}, nil, nil, nil, nil,
		{Id: 1114, WpaKey: "bbbbbbbb"}})
	disableds = disabledRe.FindAllStringSubmatch(config, -1)
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	if assert.Equal(t, 6, len(disableds)) && assert.Equal(t, 6, len(ssids)) && assert.Equal(t, 6, len(wpaKeys)) {
		assert.Equal(t, "0", disableds[0][1])
		assert.Equal(t, "254", ssids[0][1])
		assert.Equal(t, "aaaaaaaa", wpaKeys[0][1])
		for i := 1; i < 5; i++ {
			assert.Equal(t, "0", disableds[i][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", i+1), ssids[i][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", i+1), wpaKeys[i][1])
		}
		assert.Equal(t, "0", disableds[5][1])
		assert.Equal(t, "1114", ssids[5][1])
		assert.Equal(t, "bbbbbbbb", wpaKeys[5][1])
	}

	// Should configure all SSIDs for six teams.
	config, _ = generateAccessPointConfig([6]*model.Team{{Id: 1, WpaKey: "11111111"}, {Id: 2, WpaKey: "22222222"},
		{Id: 3, WpaKey: "33333333"}, {Id: 4, WpaKey: "44444444"}, {Id: 5, WpaKey: "55555555"},
		{Id: 6, WpaKey: "66666666"}})
	disableds = disabledRe.FindAllStringSubmatch(config, -1)
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	if assert.Equal(t, 6, len(ssids)) && assert.Equal(t, 6, len(wpaKeys)) {
		for i := 0; i < 6; i++ {
			assert.Equal(t, "0", disableds[i][1])
		}
		assert.Equal(t, "1", ssids[0][1])
		assert.Equal(t, "11111111", wpaKeys[0][1])
		assert.Equal(t, "2", ssids[1][1])
		assert.Equal(t, "22222222", wpaKeys[1][1])
		assert.Equal(t, "3", ssids[2][1])
		assert.Equal(t, "33333333", wpaKeys[2][1])
		assert.Equal(t, "4", ssids[3][1])
		assert.Equal(t, "44444444", wpaKeys[3][1])
		assert.Equal(t, "5", ssids[4][1])
		assert.Equal(t, "55555555", wpaKeys[4][1])
		assert.Equal(t, "6", ssids[5][1])
		assert.Equal(t, "66666666", wpaKeys[5][1])
	}

	// Should reject a missing WPA key.
	_, err := generateAccessPointConfig([6]*model.Team{{Id: 254}, nil, nil, nil, nil, nil})
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid WPA key")
	}
}

func TestDecodeWifiInfo(t *testing.T) {
	var statuses [6]TeamWifiStatus

	// Test with zero team networks configured.
	output, err := ioutil.ReadFile("testdata/iwinfo_0_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, decodeWifiInfo(string(output), statuses[:]))
		assertTeamWifiStatus(t, 0, false, statuses[0])
		assertTeamWifiStatus(t, 0, false, statuses[1])
		assertTeamWifiStatus(t, 0, false, statuses[2])
		assertTeamWifiStatus(t, 0, false, statuses[3])
		assertTeamWifiStatus(t, 0, false, statuses[4])
		assertTeamWifiStatus(t, 0, false, statuses[5])
	}

	// Test with two team networks configured.
	output, err = ioutil.ReadFile("testdata/iwinfo_2_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, decodeWifiInfo(string(output), statuses[:]))
		assertTeamWifiStatus(t, 0, false, statuses[0])
		assertTeamWifiStatus(t, 2471, true, statuses[1])
		assertTeamWifiStatus(t, 0, false, statuses[2])
		assertTeamWifiStatus(t, 254, false, statuses[3])
		assertTeamWifiStatus(t, 0, false, statuses[4])
		assertTeamWifiStatus(t, 0, false, statuses[5])
	}

	// Test with six team networks configured.
	output, err = ioutil.ReadFile("testdata/iwinfo_6_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, decodeWifiInfo(string(output), statuses[:]))
		assertTeamWifiStatus(t, 254, false, statuses[0])
		assertTeamWifiStatus(t, 1678, false, statuses[1])
		assertTeamWifiStatus(t, 2910, true, statuses[2])
		assertTeamWifiStatus(t, 604, false, statuses[3])
		assertTeamWifiStatus(t, 8, false, statuses[4])
		assertTeamWifiStatus(t, 2471, true, statuses[5])
	}

	// Test with invalid input.
	assert.NotNil(t, decodeWifiInfo("", statuses[:]))
	output, err = ioutil.ReadFile("testdata/iwinfo_invalid.txt")
	if assert.Nil(t, err) {
		assert.NotNil(t, decodeWifiInfo(string(output), statuses[:]))
	}
}

func assertTeamWifiStatus(t *testing.T, expectedTeamId int, expectedRadioLinked bool, status TeamWifiStatus) {
	assert.Equal(t, expectedTeamId, status.TeamId)
	assert.Equal(t, expectedRadioLinked, status.RadioLinked)
}
