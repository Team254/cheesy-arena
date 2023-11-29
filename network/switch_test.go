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
	sw.configBackoffDuration = time.Millisecond
	sw.configPauseDuration = time.Millisecond
	var command1, command2 string
	expectedResetCommand := "password\nenable\npassword\nterminal length 0\nconfig terminal\n" +
		"interface Vlan10\nno ip address\nno access-list 110\nno ip dhcp pool dhcp10\n" +
		"interface Vlan20\nno ip address\nno access-list 120\nno ip dhcp pool dhcp20\n" +
		"interface Vlan30\nno ip address\nno access-list 130\nno ip dhcp pool dhcp30\n" +
		"interface Vlan40\nno ip address\nno access-list 140\nno ip dhcp pool dhcp40\n" +
		"interface Vlan50\nno ip address\nno access-list 150\nno ip dhcp pool dhcp50\n" +
		"interface Vlan60\nno ip address\nno access-list 160\nno ip dhcp pool dhcp60\n" +
		"end\ncopy running-config startup-config\n\nexit\n"

	// Should remove all previous VLANs and do nothing else if current configuration is blank.
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(t, "", command2)

	// Should configure one team if only one is present.
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, {Id: 254}, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"password\nenable\npassword\nterminal length 0\nconfig terminal\n"+
			"ip dhcp excluded-address 10.2.54.1 10.2.54.19\nip dhcp excluded-address 10.2.54.200 10.2.54.254\nip dhcp pool dhcp50\n"+
			"network 10.2.54.0 255.255.255.0\ndefault-router 10.2.54.4\nlease 7\n"+
			"access-list 150 permit ip 10.2.54.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 150 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan50\nip address 10.2.54.4 255.255.255.0\n"+
			"end\ncopy running-config startup-config\n\nexit\n",
		command2,
	)

	// Should configure all teams if all are present.
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(
		t,
		sw.ConfigureTeamEthernet([6]*model.Team{{Id: 1114}, {Id: 254}, {Id: 296}, {Id: 1503}, {Id: 1678}, {Id: 1538}}),
	)
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"password\nenable\npassword\nterminal length 0\nconfig terminal\n"+
			"ip dhcp excluded-address 10.11.14.1 10.11.14.19\nip dhcp excluded-address 10.11.14.200 10.11.14.254\nip dhcp pool dhcp10\n"+
			"network 10.11.14.0 255.255.255.0\ndefault-router 10.11.14.4\nlease 7\n"+
			"access-list 110 permit ip 10.11.14.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 110 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan10\nip address 10.11.14.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.2.54.1 10.2.54.19\nip dhcp excluded-address 10.2.54.200 10.2.54.254\nip dhcp pool dhcp20\n"+
			"network 10.2.54.0 255.255.255.0\ndefault-router 10.2.54.4\nlease 7\n"+
			"access-list 120 permit ip 10.2.54.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 120 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan20\nip address 10.2.54.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.2.96.1 10.2.96.19\nip dhcp excluded-address 10.2.96.200 10.2.96.254\nip dhcp pool dhcp30\n"+
			"network 10.2.96.0 255.255.255.0\ndefault-router 10.2.96.4\nlease 7\n"+
			"access-list 130 permit ip 10.2.96.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 130 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan30\nip address 10.2.96.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.15.3.1 10.15.3.19\nip dhcp excluded-address 10.15.3.200 10.15.3.254\nip dhcp pool dhcp40\n"+
			"network 10.15.3.0 255.255.255.0\ndefault-router 10.15.3.4\nlease 7\n"+
			"access-list 140 permit ip 10.15.3.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 140 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan40\nip address 10.15.3.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.16.78.1 10.16.78.19\nip dhcp excluded-address 10.16.78.200 10.16.78.254\nip dhcp pool dhcp50\n"+
			"network 10.16.78.0 255.255.255.0\ndefault-router 10.16.78.4\nlease 7\n"+
			"access-list 150 permit ip 10.16.78.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 150 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan50\nip address 10.16.78.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.15.38.1 10.15.38.19\nip dhcp excluded-address 10.15.38.200 10.15.38.254\nip dhcp pool dhcp60\n"+
			"network 10.15.38.0 255.255.255.0\ndefault-router 10.15.38.4\nlease 7\n"+
			"access-list 160 permit ip 10.15.38.0 0.0.0.255 host 10.0.100.5\n"+
			"access-list 160 permit udp any eq bootpc any eq bootps\n"+
			"interface Vlan60\nip address 10.15.38.4 255.255.255.0\n"+
			"end\ncopy running-config startup-config\n\nexit\n",
		command2,
	)
}

func mockTelnet(t *testing.T, port int, command1 *string, command2 *string) {
	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.Nil(t, err)
		defer ln.Close()
		*command1 = ""
		*command2 = ""

		// Fake the first connection.
		conn1, err := ln.Accept()
		assert.Nil(t, err)
		conn1.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		var reader bytes.Buffer
		reader.ReadFrom(conn1)
		*command1 = reader.String()
		conn1.Close()

		// Fake the second connection.
		conn2, err := ln.Accept()
		assert.Nil(t, err)
		conn2.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		reader.Reset()
		reader.ReadFrom(conn2)
		*command2 = reader.String()
		conn2.Close()
	}()
	time.Sleep(100 * time.Millisecond) // Give it some time to open the socket.
}
