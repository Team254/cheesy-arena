// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring an SCC Switch via SSH.

package network

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	sccSwitchConnectTimeoutSec = 5
	sccSwitchConfigTimeoutSec  = 5
	sccSwitchSSHPort           = 22
)

type SCCSwitch struct {
	address                string
	port                   int
	username               string
	password               string
	mutex                  sync.Mutex
	connectTimeoutDuration time.Duration
	configTimeoutDuration  time.Duration
	upCommands             []string
	downCommands           []string
	Status                 string
}

func NewSCCSwitch(address, username, password string, upCommands, downCommands []string) *SCCSwitch {
	return &SCCSwitch{
		address:                address,
		port:                   sccSwitchSSHPort,
		username:               username,
		password:               password,
		connectTimeoutDuration: sccSwitchConnectTimeoutSec * time.Second,
		configTimeoutDuration:  sccSwitchConfigTimeoutSec * time.Second,
		upCommands:             upCommands,
		downCommands:           downCommands,
		Status:                 "UNKNOWN",
	}
}

func (scc *SCCSwitch) SetTeamEthernetEnabled(enabled bool) error {
	scc.mutex.Lock()
	defer scc.mutex.Unlock()

	scc.Status = "CONFIGURING"

	commandSequence := scc.downCommands
	if enabled {
		commandSequence = scc.upCommands
	}

	_, err := scc.runCommandSequence(commandSequence)
	if err != nil {
		scc.Status = "ERROR"
		return fmt.Errorf("failed to set team ethernet state: %w", err)
	}

	if enabled {
		scc.Status = "ACTIVE"
	} else {
		scc.Status = "DISABLED"
	}

	return nil
}

// Logs into the switch via SSH and runs the given commands in sequence.
// Returns the output of the commands or an error if the operation fails.
func (scc *SCCSwitch) runCommandSequence(commands []string) (string, error) {
	// Open an SSH connection to the switch.
	sshConfig := &ssh.ClientConfig{
		User: scc.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(scc.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Allow any host key for simplicity
		Timeout:         sccSwitchConnectTimeoutSec * time.Second,
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(scc.address, strconv.Itoa(scc.port)), sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH: %w", err)
	}
	defer client.Close()

	// Create an interactive session to run commands
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Capture the session output
	var outputBuffer bytes.Buffer
	session.Stdout = &outputBuffer
	session.Stderr = &outputBuffer

	inputPipe, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create input pipe: %w", err)
	}

	// Launch the switch's interactive shell
	err = session.Shell()
	if err != nil {
		return "", fmt.Errorf("failed to start shell: %w", err)
	}

	// Submit the commands to the switch
	for _, command := range commands {
		if _, err := fmt.Fprintln(inputPipe, command); err != nil {
			return "", fmt.Errorf("failed to write command to switch: %w", err)
		}
	}

	// Wait for the remote to process the commands and exit the shell
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("failed to run command sequence: %w", err)
		}
	case <-time.After(scc.connectTimeoutDuration):
		return "", fmt.Errorf("timed out waiting for command sequence to complete")
	}

	return outputBuffer.String(), nil
}
