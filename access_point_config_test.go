// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestConfigureAccessPoint(t *testing.T) {
	radioRe := regexp.MustCompile("option device 'radio0'")
	ssidRe := regexp.MustCompile("option ssid '([-\\w ]+)'")
	wpaKeyRe := regexp.MustCompile("option key '([-\\w ]+)'")
	vlanRe := regexp.MustCompile("option network 'vlan(\\d+)'")

	// Should not configure any team SSIDs if there are no teams.
	config, _ := generateAccessPointConfig(nil, nil, nil, nil, nil, nil)
	assert.NotContains(t, config, "option device 'radio0'")
	ssids := ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
	vlans := vlanRe.FindAllStringSubmatch(config, -1)
	assert.Equal(t, "Cheesy Arena", ssids[0][1])
	assert.Equal(t, "1234Five", wpaKeys[0][1])
	assert.Equal(t, "100", vlans[0][1])

	// Should configure two SSID for two teams.
	config, _ = generateAccessPointConfig(&model.Team{Id: 254, WpaKey: "aaaaaaaa"}, nil, nil, nil, nil,
		&model.Team{Id: 1114, WpaKey: "bbbbbbbb"})
	assert.Equal(t, 2, len(radioRe.FindAllString(config, -1)))
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	vlans = vlanRe.FindAllStringSubmatch(config, -1)
	assert.Equal(t, "Cheesy Arena", ssids[0][1])
	assert.Equal(t, "1234Five", wpaKeys[0][1])
	assert.Equal(t, "100", vlans[0][1])
	assert.Equal(t, "254", ssids[1][1])
	assert.Equal(t, "aaaaaaaa", wpaKeys[1][1])
	assert.Equal(t, "10", vlans[1][1])
	assert.Equal(t, "1114", ssids[2][1])
	assert.Equal(t, "bbbbbbbb", wpaKeys[2][1])
	assert.Equal(t, "60", vlans[2][1])

	// Should configure all SSIDs for six teams.
	config, _ = generateAccessPointConfig(&model.Team{Id: 1, WpaKey: "11111111"},
		&model.Team{Id: 2, WpaKey: "22222222"}, &model.Team{Id: 3, WpaKey: "33333333"},
		&model.Team{Id: 4, WpaKey: "44444444"}, &model.Team{Id: 5, WpaKey: "55555555"},
		&model.Team{Id: 6, WpaKey: "66666666"})
	assert.Equal(t, 6, len(radioRe.FindAllString(config, -1)))
	ssids = ssidRe.FindAllStringSubmatch(config, -1)
	wpaKeys = wpaKeyRe.FindAllStringSubmatch(config, -1)
	vlans = vlanRe.FindAllStringSubmatch(config, -1)
	assert.Equal(t, "Cheesy Arena", ssids[0][1])
	assert.Equal(t, "1234Five", wpaKeys[0][1])
	assert.Equal(t, "100", vlans[0][1])
	assert.Equal(t, "1", ssids[1][1])
	assert.Equal(t, "11111111", wpaKeys[1][1])
	assert.Equal(t, "10", vlans[1][1])
	assert.Equal(t, "2", ssids[2][1])
	assert.Equal(t, "22222222", wpaKeys[2][1])
	assert.Equal(t, "20", vlans[2][1])
	assert.Equal(t, "3", ssids[3][1])
	assert.Equal(t, "33333333", wpaKeys[3][1])
	assert.Equal(t, "30", vlans[3][1])
	assert.Equal(t, "4", ssids[4][1])
	assert.Equal(t, "44444444", wpaKeys[4][1])
	assert.Equal(t, "40", vlans[4][1])
	assert.Equal(t, "5", ssids[5][1])
	assert.Equal(t, "55555555", wpaKeys[5][1])
	assert.Equal(t, "50", vlans[5][1])
	assert.Equal(t, "6", ssids[6][1])
	assert.Equal(t, "66666666", wpaKeys[6][1])
	assert.Equal(t, "60", vlans[6][1])

	// Should reject a missing WPA key.
	_, err := generateAccessPointConfig(&model.Team{Id: 254}, nil, nil, nil, nil, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid WPA key")
	}
}
