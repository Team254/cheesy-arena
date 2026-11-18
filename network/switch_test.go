// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestConfigureSwitch(t *testing.T) {
	username := "username"
	password := "password"
	sw := NewSwitch("127.0.0.1", username, password)
	assert.Equal(t, "UNKNOWN", sw.Status)
	sw.port = 9050
	sw.connectTimeoutDuration = 50 * time.Millisecond
	sw.configTimeoutDuration = 50 * time.Millisecond
	sw.configBackoffDuration = time.Millisecond
	sw.configPauseDuration = time.Millisecond
	var command1, command2 string
	expectedResetCommand := "terminal length 0\nconfig terminal\n" +
		"interface Vlan10\nno ip address\nno ip dhcp pool dhcp10\n" +
		"interface Vlan20\nno ip address\nno ip dhcp pool dhcp20\n" +
		"interface Vlan30\nno ip address\nno ip dhcp pool dhcp30\n" +
		"interface Vlan40\nno ip address\nno ip dhcp pool dhcp40\n" +
		"interface Vlan50\nno ip address\nno ip dhcp pool dhcp50\n" +
		"interface Vlan60\nno ip address\nno ip dhcp pool dhcp60\n" +
		"end\nexit\n"

	// Should remove all previous VLANs and do nothing else if current configuration is blank.
	mockSSH(t, sw.port, username, password, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(t, "", command2)
	assert.Equal(t, "ACTIVE", sw.Status)

	// Should configure one team if only one is present.
	sw.port += 1
	mockSSH(t, sw.port, username, password, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, {Id: 254}, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"terminal length 0\nconfig terminal\n"+
			"ip dhcp excluded-address 10.2.54.1 10.2.54.19\nip dhcp excluded-address 10.2.54.200 10.2.54.254\nip dhcp pool dhcp50\n"+
			"network 10.2.54.0 255.255.255.0\ndefault-router 10.2.54.4\nlease 7\n"+
			"interface Vlan50\nip address 10.2.54.4 255.255.255.0\n"+
			"end\nexit\n",
		command2,
	)

	// Should configure all teams if all are present.
	sw.port += 1
	mockSSH(t, sw.port, username, password, &command1, &command2)
	assert.Nil(
		t,
		sw.ConfigureTeamEthernet([6]*model.Team{{Id: 1114}, {Id: 254}, {Id: 296}, {Id: 1503}, {Id: 1678}, {Id: 1538}}),
	)
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"terminal length 0\nconfig terminal\n"+
			"ip dhcp excluded-address 10.11.14.1 10.11.14.19\nip dhcp excluded-address 10.11.14.200 10.11.14.254\nip dhcp pool dhcp10\n"+
			"network 10.11.14.0 255.255.255.0\ndefault-router 10.11.14.4\nlease 7\n"+
			"interface Vlan10\nip address 10.11.14.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.2.54.1 10.2.54.19\nip dhcp excluded-address 10.2.54.200 10.2.54.254\nip dhcp pool dhcp20\n"+
			"network 10.2.54.0 255.255.255.0\ndefault-router 10.2.54.4\nlease 7\n"+
			"interface Vlan20\nip address 10.2.54.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.2.96.1 10.2.96.19\nip dhcp excluded-address 10.2.96.200 10.2.96.254\nip dhcp pool dhcp30\n"+
			"network 10.2.96.0 255.255.255.0\ndefault-router 10.2.96.4\nlease 7\n"+
			"interface Vlan30\nip address 10.2.96.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.15.3.1 10.15.3.19\nip dhcp excluded-address 10.15.3.200 10.15.3.254\nip dhcp pool dhcp40\n"+
			"network 10.15.3.0 255.255.255.0\ndefault-router 10.15.3.4\nlease 7\n"+
			"interface Vlan40\nip address 10.15.3.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.16.78.1 10.16.78.19\nip dhcp excluded-address 10.16.78.200 10.16.78.254\nip dhcp pool dhcp50\n"+
			"network 10.16.78.0 255.255.255.0\ndefault-router 10.16.78.4\nlease 7\n"+
			"interface Vlan50\nip address 10.16.78.4 255.255.255.0\n"+
			"ip dhcp excluded-address 10.15.38.1 10.15.38.19\nip dhcp excluded-address 10.15.38.200 10.15.38.254\nip dhcp pool dhcp60\n"+
			"network 10.15.38.0 255.255.255.0\ndefault-router 10.15.38.4\nlease 7\n"+
			"interface Vlan60\nip address 10.15.38.4 255.255.255.0\n"+
			"end\nexit\n",
		command2,
	)

	// Should set status to ERROR if connection fails.
	sw.port += 1
	assert.NotNil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, "ERROR", sw.Status)
}

func mockSSH(t *testing.T, port int, username, password string, command1 *string, command2 *string) {
	go func() {
		_, privateKey, err := ed25519.GenerateKey(nil)
		assert.Nil(t, err)
		signer, err := ssh.NewSignerFromKey(privateKey)
		assert.Nil(t, err)
		config := &ssh.ServerConfig{
			PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				assert.Equal(t, username, conn.User())
				assert.Equal(t, password, string(pass))
				return nil, nil
			},
		}
		config.AddHostKey(signer)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.Nil(t, err)
		t.Cleanup(func() {
			listener.Close()
		})
		defer listener.Close()
		*command1 = ""
		*command2 = ""

		acceptSSHConnection(t, listener, config, command1)
		acceptSSHConnection(t, listener, config, command2)
	}()
	time.Sleep(100 * time.Millisecond) // Give it some time to open the socket.
}

func acceptSSHConnection(t *testing.T, listener net.Listener, config *ssh.ServerConfig, command *string) {
	nConn, err := listener.Accept()
	if errors.Is(err, net.ErrClosed) {
		return
	}
	assert.Nil(t, err)
	defer nConn.Close()
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	assert.Nil(t, err)
	defer conn.Close()

	go ssh.DiscardRequests(reqs)

	rawChannel := <-chans
	assert.Equal(t, "session", rawChannel.ChannelType())
	channel, requests, err := rawChannel.Accept()
	assert.Nil(t, err)
	defer channel.Close()

	req := <-requests
	assert.Equal(t, "pty-req", req.Type)
	req.Reply(true, nil)

	req = <-requests
	assert.Equal(t, "shell", req.Type)
	req.Reply(true, nil)

	var receivedData bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 1024)
		for {
			n, err := channel.Read(buffer)
			if n > 0 {
				receivedData.Write(buffer[:n])
				if strings.HasSuffix(receivedData.String(), "exit\n") {
					break
				}
			}
			if err != nil {
				assert.Equal(t, io.EOF, err)
				break
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		channel.Close()
		<-done
		assert.Fail(t, "timed out waiting for SSH command")
		return
	}

	*command = receivedData.String()
	channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
}
