// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestConfigureAironet(t *testing.T) {
	aironetTelnetPort = 9023
	eventSettings = &EventSettings{ApAddress: "127.0.0.1", ApUsername: "user", ApPassword: "password"}
	var command string

	// Should do nothing if current configuration is blank.
	mockTelnet(t, aironetTelnetPort, "", &command)
	assert.Nil(t, ConfigureTeamWifi(nil, nil, nil, nil, nil, nil))
	assert.Equal(t, "", command)

	// Should remove any existing teams but not other SSIDs.
	aironetTelnetPort += 1
	mockTelnet(t, aironetTelnetPort,
		"dot11 ssid 1\nvlan 1\ndot11 ssid 254\nvlan 12\ndot11 ssid Cheesy Arena\nvlan 17\n", &command)
	assert.Nil(t, ConfigureTeamWifi(nil, nil, nil, nil, nil, nil))
	assert.Equal(t, "user\npassword\nterminal length 0\nconfig terminal\nno dot11 ssid 254\nend\n"+
		"copy running-config startup-config\n\nexit\n", command)

	// Should configure new teams and leave existing ones alone if still needed.
	aironetTelnetPort += 1
	mockTelnet(t, aironetTelnetPort, "dot11 ssid 254\nvlan 11\n", &command)
	assert.Nil(t, ConfigureTeamWifi(&Team{Id: 254, WpaKey: "aaaaaaaa"}, nil, nil, nil, nil,
		&Team{Id: 1114, WpaKey: "bbbbbbbb"}))
	assert.Equal(t, "user\npassword\nterminal length 0\nconfig terminal\ndot11 ssid 1114\nvlan 16\n"+
		"authentication open\nauthentication key-management wpa version 2\nmbssid guest-mode\nwpa-psk ascii "+
		"bbbbbbbb\ninterface Dot11Radio1\nssid 1114\nend\ncopy running-config startup-config\n\nexit\n",
		command)

	// Should reject a missing WPA key.
	aironetTelnetPort += 1
	mockTelnet(t, aironetTelnetPort, "", &command)
	err := ConfigureTeamWifi(&Team{Id: 254}, nil, nil, nil, nil, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid WPA key")
	}
}

func mockTelnet(t *testing.T, port int, response string, command *string) {
	go func() {
		// Fake the first connection which should just get the configuration.
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.Nil(t, err)
		defer ln.Close()
		conn, err := ln.Accept()
		assert.Nil(t, err)
		conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		var reader bytes.Buffer
		reader.ReadFrom(conn)
		assert.Contains(t, reader.String(), "terminal length 0\nshow running-config\nexit\n")
		conn.Write([]byte(response))
		conn.Close()

		// Fake the second connection which should configure stuff.
		conn2, err := ln.Accept()
		assert.Nil(t, err)
		conn2.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		var reader2 bytes.Buffer
		reader2.ReadFrom(conn2)
		*command = reader2.String()
		conn2.Close()
	}()
	time.Sleep(100 * time.Millisecond) // Give it some time to open the socket.
}
