// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestConfigureAccessPoint(t *testing.T) {
	model.BaseDir = ".."

	ssidRe := regexp.MustCompile("ssid='([-\\w ]+)'")
	wpaKeyRe := regexp.MustCompile("key='([-\\w ]+)'")
	ap := AccessPoint{teamChannel: 1234, adminChannel: 4321, adminWpaKey: "blorpy"}

	// Should not configure any team SSIDs if there are no teams.
	config, _ := ap.generateAccessPointConfig(nil, nil, nil, nil, nil, nil)
	assert.NotContains(t, config, "set")
	ssids := ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
	assert.Equal(t, 0, len(ssids))
	assert.Equal(t, 0, len(wpaKeys))

	// Should configure two SSID for two teams.
	config, _ = ap.generateAccessPointConfig(&model.Team{Id: 254, WpaKey: "aaaaaaaa"}, nil, nil, nil, nil,
		&model.Team{Id: 1114, WpaKey: "bbbbbbbb"})
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	if assert.Equal(t, 2, len(ssids)) && assert.Equal(t, 2, len(wpaKeys)) {
		assert.Equal(t, "254", ssids[0][1])
		assert.Equal(t, "aaaaaaaa", wpaKeys[0][1])
		assert.Equal(t, "1114", ssids[1][1])
		assert.Equal(t, "bbbbbbbb", wpaKeys[1][1])
	}

	// Should configure all SSIDs for six teams.
	config, _ = ap.generateAccessPointConfig(&model.Team{Id: 1, WpaKey: "11111111"},
		&model.Team{Id: 2, WpaKey: "22222222"}, &model.Team{Id: 3, WpaKey: "33333333"},
		&model.Team{Id: 4, WpaKey: "44444444"}, &model.Team{Id: 5, WpaKey: "55555555"},
		&model.Team{Id: 6, WpaKey: "66666666"})
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	if assert.Equal(t, 6, len(ssids)) && assert.Equal(t, 6, len(wpaKeys)) {
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
	_, err := ap.generateAccessPointConfig(&model.Team{Id: 254}, nil, nil, nil, nil, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid WPA key")
	}
}
