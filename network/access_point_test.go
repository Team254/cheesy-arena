// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTeamAccessPointConfigForLinksys(t *testing.T) {
	model.BaseDir = ".."
	ap := AccessPoint{isVividType: false}

	ifaceRe := regexp.MustCompile("^set wireless\\.@wifi-iface\\[(\\d)\\]\\.")
	disabledRe := regexp.MustCompile("disabled='([-\\w ]+)'")
	ssidRe := regexp.MustCompile("ssid='([-\\w ]*)'")
	wpaKeyRe := regexp.MustCompile("key='([-\\w ]*)'")

	// Should reject invalid positions.
	for _, position := range []int{-1, 0, 7, 8, 254} {
		_, err := ap.generateTeamAccessPointConfig(nil, position)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), fmt.Sprintf("invalid team position %d", position))
		}
	}

	// Should configure dummy values for all team SSIDs if there are no teams.
	for position := 1; position <= 6; position++ {
		config, _ := ap.generateTeamAccessPointConfig(nil, position)
		ifaces := ifaceRe.FindAllStringSubmatch(config, -1)
		disableds := disabledRe.FindAllStringSubmatch(config, -1)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(disableds)) && assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, strconv.Itoa(position), ifaces[0][1])
			assert.Equal(t, "0", disableds[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), wpaKeys[0][1])
		}
	}

	// Should configure a different SSID for each team.
	for position := 1; position <= 6; position++ {
		team := &model.Team{Id: 254 + position, WpaKey: fmt.Sprintf("aaaaaaa%d", position)}
		config, _ := ap.generateTeamAccessPointConfig(team, position)
		ifaces := ifaceRe.FindAllStringSubmatch(config, -1)
		disableds := disabledRe.FindAllStringSubmatch(config, -1)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, strconv.Itoa(position), ifaces[0][1])
			assert.Equal(t, "0", disableds[0][1])
			assert.Equal(t, strconv.Itoa(team.Id), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("aaaaaaa%d", position), wpaKeys[0][1])
		}
	}

	// Should reject a missing WPA key.
	_, err := ap.generateTeamAccessPointConfig(&model.Team{Id: 254}, 4)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "invalid WPA key")
	}
}

func TestGenerateTeamAccessPointConfigForVividHosting(t *testing.T) {
	model.BaseDir = ".."
	ap := AccessPoint{isVividType: true}

	ifaceRe := regexp.MustCompile("^set wireless\\.@wifi-iface\\[(\\d)\\]\\.")
	disabledRe := regexp.MustCompile("disabled='([-\\w ]+)'")
	ssidRe := regexp.MustCompile("ssid='([-\\w ]*)'")
	wpaKeyRe := regexp.MustCompile("key='([-\\w ]*)'")
	saePasswordRe := regexp.MustCompile("sae_password='([-\\w ]*)'")

	// Should reject invalid positions.
	for _, position := range []int{-1, 0, 7, 8, 254} {
		_, err := ap.generateTeamAccessPointConfig(nil, position)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), fmt.Sprintf("invalid team position %d", position))
		}
	}

	// Should configure dummy values for all team SSIDs if there are no teams.
	for position := 1; position <= 6; position++ {
		config, _ := ap.generateTeamAccessPointConfig(nil, position)
		ifaces := ifaceRe.FindAllStringSubmatch(config, -1)
		disableds := disabledRe.FindAllStringSubmatch(config, -1)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		saePasswords := saePasswordRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(disableds)) && assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, strconv.Itoa(position), ifaces[0][1])
			assert.Equal(t, "0", disableds[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), wpaKeys[0][1])
			assert.Equal(t, fmt.Sprintf("no-team-%d", position), saePasswords[0][1])
		}
	}

	// Should configure a different SSID for each team.
	for position := 1; position <= 6; position++ {
		team := &model.Team{Id: 254 + position, WpaKey: fmt.Sprintf("aaaaaaa%d", position)}
		config, _ := ap.generateTeamAccessPointConfig(team, position)
		ifaces := ifaceRe.FindAllStringSubmatch(config, -1)
		disableds := disabledRe.FindAllStringSubmatch(config, -1)
		ssids := ssidRe.FindAllStringSubmatch(config, -1)
		wpaKeys := wpaKeyRe.FindAllStringSubmatch(config, -1)
		saePasswords := saePasswordRe.FindAllStringSubmatch(config, -1)
		if assert.Equal(t, 1, len(ssids)) && assert.Equal(t, 1, len(wpaKeys)) {
			assert.Equal(t, strconv.Itoa(position), ifaces[0][1])
			assert.Equal(t, "0", disableds[0][1])
			assert.Equal(t, strconv.Itoa(team.Id), ssids[0][1])
			assert.Equal(t, fmt.Sprintf("aaaaaaa%d", position), wpaKeys[0][1])
			assert.Equal(t, fmt.Sprintf("aaaaaaa%d", position), saePasswords[0][1])
		}
	}

	// Should reject a missing WPA key.
	_, err := ap.generateTeamAccessPointConfig(&model.Team{Id: 254}, 4)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "invalid WPA key")
	}
}

func TestDecodeWifiInfo(t *testing.T) {
	statuses := [6]*TeamWifiStatus{
		nil,
		&TeamWifiStatus{},
		&TeamWifiStatus{},
		&TeamWifiStatus{},
		nil,
		&TeamWifiStatus{},
	}
	ap := AccessPoint{isVividType: true, TeamWifiStatuses: statuses}

	// Test with zero team networks configured.
	output, err := ioutil.ReadFile("testdata/iwinfo_0_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, ap.decodeWifiInfo(string(output)))
		assert.Nil(t, statuses[0])
		assert.Equal(t, 0, statuses[1].TeamId)
		assert.Equal(t, 0, statuses[2].TeamId)
		assert.Equal(t, 0, statuses[3].TeamId)
		assert.Nil(t, statuses[4])
		assert.Equal(t, 0, statuses[5].TeamId)
	}

	// Test with two team networks configured.
	output, err = ioutil.ReadFile("testdata/iwinfo_2_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, ap.decodeWifiInfo(string(output)))
		assert.Nil(t, statuses[0])
		assert.Equal(t, 2471, statuses[1].TeamId)
		assert.Equal(t, 0, statuses[2].TeamId)
		assert.Equal(t, 254, statuses[3].TeamId)
		assert.Nil(t, statuses[4])
		assert.Equal(t, 0, statuses[5].TeamId)
	}

	// Test with six team networks configured.
	output, err = ioutil.ReadFile("testdata/iwinfo_6_teams.txt")
	if assert.Nil(t, err) {
		assert.Nil(t, ap.decodeWifiInfo(string(output)))
		assert.Nil(t, statuses[0])
		assert.Equal(t, 1678, statuses[1].TeamId)
		assert.Equal(t, 2910, statuses[2].TeamId)
		assert.Equal(t, 604, statuses[3].TeamId)
		assert.Nil(t, statuses[4])
		assert.Equal(t, 2471, statuses[5].TeamId)
	}

	// Test with invalid input.
	assert.NotNil(t, ap.decodeWifiInfo(""))
	output, err = ioutil.ReadFile("testdata/iwinfo_invalid.txt")
	if assert.Nil(t, err) {
		assert.NotNil(t, ap.decodeWifiInfo(string(output)))
	}
}

func TestParseBtu(t *testing.T) {
	// Response is too short.
	assert.Equal(t, 0.0, parseBtu(""))
	response := "[ 1687496957, 26097, 177, 71670, 865 ],\n" +
		"[ 1687496958, 26097, 177, 71734, 866 ],\n" +
		"[ 1687496959, 26097, 177, 71734, 866 ],\n" +
		"[ 1687496960, 26097, 177, 71798, 867 ],\n" +
		"[ 1687496960, 26097, 177, 71798, 867 ],\n" +
		"[ 1687496961, 26097, 177, 71798, 867 ]"
	assert.Equal(t, 0.0, parseBtu(response))

	// Response is normal.
	response = "[ 1687496917, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496919, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496921, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496922, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496923, 2609700, 177, 7064600, 849 ]"
	assert.Equal(t, 15.0, math.Floor(parseBtu(response)))

	// Response also includes associated client information.
	response = "[ 1687496917, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496919, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496921, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496922, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496923, 2609700, 177, 7064600, 849 ]\n" +
		"48:DA:35:B0:00:CF  -52 dBm / -95 dBm (SNR 43)  1000 ms ago\n" +
		"\tRX: 619.4 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	assert.Equal(t, 15.0, math.Floor(parseBtu(response)))
}

func TestParseAssocList(t *testing.T) {
	var wifiStatus TeamWifiStatus

	wifiStatus.parseAssocList("")
	assert.Equal(t, TeamWifiStatus{}, wifiStatus)

	// MAC address is invalid.
	response := "00:00:00:00:00:00  -53 dBm / -95 dBm (SNR 42)  0 ms ago\n" +
		"\tRX: 550.6 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{}, wifiStatus)

	// Link is valid.
	response = "48:DA:35:B0:00:CF  -53 dBm / -95 dBm (SNR 42)  0 ms ago\n" +
		"\tRX: 550.6 MBit/s                                4095 Pkts.\n" +
		"\tTX: 254.0 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{RadioLinked: true, RxRate: 550.6, TxRate: 254.0, SignalNoiseRatio: 42}, wifiStatus)
	response = "48:DA:35:B0:00:CF  -53 dBm / -95 dBm (SNR 7)  4000 ms ago\n" +
		"\tRX: 123.4 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{RadioLinked: true, RxRate: 123.4, TxRate: 550.6, SignalNoiseRatio: 7}, wifiStatus)

	// Link is stale.
	response = "48:DA:35:B0:00:CF  -53 dBm / -95 dBm (SNR 42)  4001 ms ago\n" +
		"\tRX: 550.6 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{}, wifiStatus)

	// Response also includes BTU information.
	response = "[ 1687496917, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496919, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496921, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496922, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496923, 2609700, 177, 7064600, 849 ]\n" +
		"48:DA:35:B0:00:CF  -52 dBm / -95 dBm (SNR 43)  1000 ms ago\n" +
		"\tRX: 619.4 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{RadioLinked: true, RxRate: 619.4, TxRate: 550.6, SignalNoiseRatio: 43}, wifiStatus)
	response = "[ 1687496917, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496919, 26097, 177, 70454, 846 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496920, 26097, 177, 70518, 847 ],\n" +
		"[ 1687496921, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496922, 26097, 177, 70582, 848 ],\n" +
		"[ 1687496923, 2609700, 177, 7064600, 849 ]\n" +
		"00:00:00:00:00:00  -52 dBm / -95 dBm (SNR 43)  0 ms ago\n" +
		"\tRX: 619.4 MBit/s                                4095 Pkts.\n" +
		"\tTX: 550.6 MBit/s                                   0 Pkts.\n" +
		"\texpected throughput: unknown"
	wifiStatus.parseAssocList(response)
	assert.Equal(t, TeamWifiStatus{}, wifiStatus)
}
