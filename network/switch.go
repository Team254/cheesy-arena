// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Layer 3 Cisco Catalyst IOS switch for team VLANs.

package network

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	switchConnectTimeoutSec        = 5
	switchConfigTimeoutSec         = 5
	switchConfigBackoffDurationSec = 5
	switchConfigPauseDurationSec   = 2
	switchTeamGatewayAddress       = 4
	switchSSHPort                  = 22
)

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

type Switch struct {
	address                string
	port                   int
	username               string
	password               string
	mutex                  sync.Mutex
	connectTimeoutDuration time.Duration
	configTimeoutDuration  time.Duration
	configBackoffDuration  time.Duration
	configPauseDuration    time.Duration
	Status                 string
}

const ServerIpAddress = "10.0.100.5" // The DS will try to connect to this address only.

// DevMode allows driver station listeners to bind to all local IP addresses.
var DevMode = false

func NewSwitch(address, username, password string) *Switch {
	return &Switch{
		address:                address,
		port:                   switchSSHPort,
		username:               username,
		password:               password,
		connectTimeoutDuration: switchConnectTimeoutSec * time.Second,
		configTimeoutDuration:  switchConfigTimeoutSec * time.Second,
		configBackoffDuration:  switchConfigBackoffDurationSec * time.Second,
		configPauseDuration:    switchConfigPauseDurationSec * time.Second,
		Status:                 "UNKNOWN",
	}
}

// Sets up wired networks for the given set of teams.
func (sw *Switch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	// Make sure multiple configurations aren't being set at the same time.
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.Status = "CONFIGURING"

	// Remove old team VLANs to reset the switch state.
	removeTeamVlansCommand := ""
	for vlan := 10; vlan <= 60; vlan += 10 {
		removeTeamVlansCommand += fmt.Sprintf(
			"interface Vlan%d\nno ip address\nno ip dhcp pool dhcp%d\n", vlan, vlan,
		)
	}
	_, err := sw.runConfigCommand(removeTeamVlansCommand)
	if err != nil {
		sw.Status = "ERROR"
		return err
	}
	time.Sleep(sw.configPauseDuration)

	// Create the new team VLANs.
	addTeamVlansCommand := ""
	addTeamVlan := func(team *model.Team, vlan int) {
		if team == nil {
			return
		}
		teamPartialIp := fmt.Sprintf("%d.%d", team.Id/100, team.Id%100)
		addTeamVlansCommand += fmt.Sprintf(
			"ip dhcp excluded-address 10.%s.1 10.%s.19\n"+
				"ip dhcp excluded-address 10.%s.200 10.%s.254\n"+
				"ip dhcp pool dhcp%d\n"+
				"network 10.%s.0 255.255.255.0\n"+
				"default-router 10.%s.%d\n"+
				"lease 7\n"+
				"interface Vlan%d\nip address 10.%s.%d 255.255.255.0\n",
			teamPartialIp,
			teamPartialIp,
			teamPartialIp,
			teamPartialIp,
			vlan,
			teamPartialIp,
			teamPartialIp,
			switchTeamGatewayAddress,
			vlan,
			teamPartialIp,
			switchTeamGatewayAddress,
		)
	}
	addTeamVlan(teams[0], red1Vlan)
	addTeamVlan(teams[1], red2Vlan)
	addTeamVlan(teams[2], red3Vlan)
	addTeamVlan(teams[3], blue1Vlan)
	addTeamVlan(teams[4], blue2Vlan)
	addTeamVlan(teams[5], blue3Vlan)
	if len(addTeamVlansCommand) > 0 {
		_, err = sw.runConfigCommand(addTeamVlansCommand)
		if err != nil {
			sw.Status = "ERROR"
			return err
		}
	}

	// Give some time for the configuration to take before another one can be attempted.
	time.Sleep(sw.configBackoffDuration)

	sw.Status = "ACTIVE"
	return nil
}

type synchronizedBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func (b *synchronizedBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.Write(p)
}

func (b *synchronizedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.String()
}

// Logs into the switch via SSH and runs the given command in privileged exec mode. Reads the output and
// returns it as a string.
func (sw *Switch) runCommand(command string) (string, error) {
	// Open an SSH connection to the switch with deadlines covering handshake and command execution.
	addr := net.JoinHostPort(sw.address, strconv.Itoa(sw.port))
	rawConn, err := net.DialTimeout("tcp", addr, sw.connectTimeoutDuration)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH: %w", err)
	}
	defer rawConn.Close()

	sshConfig := &ssh.ClientConfig{
		User: sw.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(sw.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Allow any host key for simplicity
		Timeout:         sw.connectTimeoutDuration,
	}
	clientConn, chans, reqs, err := ssh.NewClientConn(rawConn, addr, sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH: %w", err)
	}
	client := ssh.NewClient(clientConn, chans, reqs)
	defer client.Close()

	if err := rawConn.SetDeadline(time.Now().Add(sw.configTimeoutDuration)); err != nil {
		return "", fmt.Errorf("failed to set connection deadline: %w", err)
	}

	// Create an interactive session to run commands.
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Capture the session output.
	var outputBuffer synchronizedBuffer
	session.Stdout = &outputBuffer
	session.Stderr = &outputBuffer

	inputPipe, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create input pipe: %w", err)
	}

	modes := ssh.TerminalModes{ssh.ECHO: 0}
	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		return "", fmt.Errorf("failed to configure shell: %w", err)
	}

	// Launch the switch's interactive shell.
	err = session.Shell()
	if err != nil {
		return "", fmt.Errorf("failed to start shell: %w", err)
	}

	// Send the command and log out all at once.
	_, err = fmt.Fprintf(inputPipe, "terminal length 0\n%sexit\n", command)
	if err != nil {
		return "", fmt.Errorf("failed to write command to switch: %w", err)
	}

	// Wait for the remote to process the commands and exit the shell.
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case err := <-done:
		var exitMissingErr *ssh.ExitMissingError
		if err != nil && !errors.As(err, &exitMissingErr) {
			return "", fmt.Errorf("failed to run command: %w", err)
		}
	case <-time.After(sw.configTimeoutDuration):
		_ = rawConn.Close()
		return "", fmt.Errorf("timed out waiting for command to complete")
	}

	return outputBuffer.String(), nil
}

// Logs into the switch via SSH and runs the given command in global configuration mode. Reads the output
// and returns it as a string.
func (sw *Switch) runConfigCommand(command string) (string, error) {
	return sw.runCommand(fmt.Sprintf("config terminal\n%send\n", command))
}
