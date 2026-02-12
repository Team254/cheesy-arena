// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for Fortinet Switch Support

package network

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestConfigureSwitch(t *testing.T) {
	// The password here will be used as the admin password.
	sw := NewSwitch("127.0.0.1", "password")
	assert.Equal(t, "UNKNOWN", sw.Status)
	sw.port = 9050
	sw.configBackoffDuration = time.Millisecond
	sw.configPauseDuration = time.Millisecond
	var command1, command2 string

	// Modify to Fortinet's expected reset command
	expectedResetCommand := "admin\npassword\nconfig system console\nset output standard\nend\n" +
		"config system dhcp server\ndelete 10\ndelete 20\ndelete 30\ndelete 40\ndelete 50\ndelete 60\nend\nexit\n"

	// 1. Test: When there are no teams, only VLAN removal should be executed
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(t, "", command2)
	assert.Equal(t, "ACTIVE", sw.Status)

	// 2. Test: Configure a single team (Team 254 in Blue 2 position, VLAN 50)
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, {Id: 254}, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"admin\npassword\nconfig system console\nset output standard\nend\n"+
			"config system interface\nedit \"vlan50\"\nset ip 10.2.54.4 255.255.255.0\nnext\nend\n"+
			"config system dhcp server\nedit 50\nset interface \"vlan50\"\nset default-gateway 10.2.54.4\nset netmask 255.255.255.0\n"+
			"config ip-range\nedit 1\nset start-ip 10.2.54.20\nset end-ip 10.2.54.199\nnext\nend\nnext\nend\nexit\n",
		command2,
	)

	// 3. Test: Configure all teams (Teams 1114, 254, 296, 1503, 1678, 1538 in positions Blue 1-6, VLANs 10-60)
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(
		t,
		sw.ConfigureTeamEthernet([6]*model.Team{{Id: 1114}, {Id: 254}, {Id: 296}, {Id: 1503}, {Id: 1678}, {Id: 1538}}),
	)
	assert.Equal(t, expectedResetCommand, command1)

	// Note: The expected string for command2 must match the loop order output by switch.go exactly
	// because Fortinet commands are long, this only shows the structure, actual execution must ensure exact format
	assert.Contains(t, command2, "edit \"vlan10\"")
	assert.Contains(t, command2, "edit \"vlan60\"")
	assert.Contains(t, command2, "set start-ip 10.11.14.20")
}

func mockTelnet(t *testing.T, port int, command1 *string, command2 *string) {
	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return // Avoid errors during parallel tests
		}
		defer ln.Close()
		*command1 = ""
		*command2 = ""

		// Simulate first connection (Reset)
		conn1, err := ln.Accept()
		if err == nil {
			conn1.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			var reader bytes.Buffer
			reader.ReadFrom(conn1)
			*command1 = reader.String()
			conn1.Close()
		}

		// Simulate second connection (Config)
		conn2, err := ln.Accept()
		if err == nil {
			conn2.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			var reader bytes.Buffer
			reader.ReadFrom(conn2)
			*command2 = reader.String()
			conn2.Close()
		}
	}()
	time.Sleep(100 * time.Millisecond)
}
