// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client for interfacing with one or more Blackmagic HyperDeck devices to automatically record matches.

package partner

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	blackmagicPort             = 9993
	blackmagicConnectTimeoutMs = 100
	blackmagicStopDelaySec     = 10
)

type BlackmagicClient struct {
	deviceAddresses []string
}

// Creates a new Blackmagic client with the given device addresses as a comma-separated string.
func NewBlackmagicClient(addresses string) *BlackmagicClient {
	var deviceAddresses []string
	for _, address := range strings.Split(addresses, ",") {
		trimmedAddress := strings.TrimSpace(address)
		if trimmedAddress != "" {
			deviceAddresses = append(deviceAddresses, trimmedAddress)
		}
	}
	return &BlackmagicClient{deviceAddresses: deviceAddresses}
}

// Starts recording across all devices.
func (client *BlackmagicClient) StartRecording() {
	client.sendCommand("record")
}

// Stops recording across all devices after a delay.
func (client *BlackmagicClient) StopRecording() {
	time.Sleep(blackmagicStopDelaySec * time.Second)
	client.sendCommand("stop")
}

// Connects to all devices and executes the given command.
func (client *BlackmagicClient) sendCommand(command string) {
	for _, address := range client.deviceAddresses {
		conn, err := net.DialTimeout(
			"tcp", fmt.Sprintf("%s:%d", address, blackmagicPort), blackmagicConnectTimeoutMs*time.Millisecond,
		)
		if err != nil {
			log.Printf("Failed to connect to Blackmagic device at %s: %v", address, err)
			continue
		}
		defer conn.Close()
		_, err = fmt.Fprint(conn, command+"\n")
		if err != nil {
			log.Printf("Failed to send '%s' command to Blackmagic device at %s: %v", command, address, err)
		}
	}
}
