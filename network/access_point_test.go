// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTeamAccessPointConfig(t *testing.T) {
	model.BaseDir = ".."

	disabledRe := regexp.MustCompile("disabled='([-\\w ]+)'")
	ssidRe := regexp.MustCompile("ssid='([-\\w ]*)'")
	wpaKeyRe := regexp.MustCompile("key='([-\\w ]*)'")

	// Should reject invalid positions.
	for _, position := range []int{-1, 0, 7, 8, 254} {
		_, err := generateTeamAccessPointConfig(nil, position)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), fmt.Sprintf("invalid team position %d", position))
		}
	}

	// Should configure dummy values for all team SSIDs if there are no teams.
	for position := 1; position <= 6; position++ {
		config, _ := generateTeamAccessPointConfig(nil, position)
		disableds := disabledRe.FindAllStringSubmatch(config, -1)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(disableds)) && assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, "0", disableds[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), wpaKeys[0][1])
		}
	}

	// Should configure a different SSID for each team.
	for position := 1; position <= 6; position++ {
		team := &model.Team{Id: 254 + position, WpaKey: fmt.Sprintf("aaaaaaa%d", position)}
		config, _ := generateTeamAccessPointConfig(team, position)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, strconv.Itoa(team.Id), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("aaaaaaa%d", position), wpaKeys[0][1])
		}
	}

	// Should reject a missing WPA key.
	_, err := generateTeamAccessPointConfig(&model.Team{Id: 254}, 4)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "invalid WPA key")
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
