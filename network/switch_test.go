// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestConfigureSwitch(t *testing.T) {
	sw := NewSwitch("127.0.0.1", "password")
	sw.port = 9050
	var command string

	// Should do nothing if current configuration is blank.
	mockTelnet(t, sw.port, "", &command)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, "", command)

	// Should remove any existing teams but not other SSIDs.
	sw.port += 1
	mockTelnet(t, sw.port,
		"interface Vlan100\nip address 10.0.100.2\ninterface Vlan50\nip address 10.2.54.61\n", &command)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, "password\nenable\npassword\nterminal length 0\nconfig terminal\ninterface Vlan50\nno ip"+
		" address\nno access-list 150\nend\ncopy running-config startup-config\n\nexit\n", command)

	// Should configure new teams and leave existing ones alone if still needed.
	sw.port += 1
	mockTelnet(t, sw.port, "interface Vlan50\nip address 10.2.54.61\n", &command)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, &model.Team{Id: 1114}, nil, nil, &model.Team{Id: 254},
		nil}))
	assert.Equal(t, "password\nenable\npassword\nterminal length 0\nconfig terminal\n"+
		"ip dhcp excluded-address 10.11.14.1 10.11.14.100\nno ip dhcp pool dhcp20\nip dhcp pool dhcp20\n"+
		"network 10.11.14.0 255.255.255.0\ndefault-router 10.11.14.61\nlease 7\nno access-list 120\n"+
		"access-list 120 permit ip 10.11.14.0 0.0.0.255 host 10.0.100.5\n"+
		"access-list 120 permit udp any eq bootpc any eq bootps\ninterface Vlan20\n"+
		"ip address 10.11.14.61 255.255.255.0\nend\ncopy running-config startup-config\n\nexit\n", command)
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
