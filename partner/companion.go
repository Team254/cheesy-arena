// Copyright 2025 Team 254. All Rights Reserved.
// Author: kyle@team2481.com (Kyle Waremburg)
//
// Client for interfacing with Bitfocus Companion to automatically trigger events during matches.

package partner

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	companionConnectTimeoutMs = 1000
)

// CompanionEvent represents the different events that can be sent to Companion
type CompanionEvent string

const (
	EventMatchPreview      CompanionEvent = "matchPreview"
	EventSetAudience       CompanionEvent = "setAudience"
	EventMatchStart        CompanionEvent = "matchStart"
	EventTeleopStart       CompanionEvent = "teleopStart"
	EventEndgameStart      CompanionEvent = "endgameStart"
	EventMatchEnd          CompanionEvent = "matchEnd"
	EventPostResult        CompanionEvent = "postResult"
	EventAllianceSelection CompanionEvent = "allianceSelection"
	EventMatchAbort        CompanionEvent = "matchAbort"
)

// CompanionEventConfig holds the page/row/column configuration for a specific event
type CompanionEventConfig struct {
	Page   int
	Row    int
	Column int
}

type CompanionClient struct {
	address string
	port    int
	events  map[CompanionEvent]CompanionEventConfig
}

// Creates a new Companion client with the given configuration.
func NewCompanionClient(
	address string,
	port int,
	eventConfigs map[CompanionEvent]CompanionEventConfig,
) *CompanionClient {
	return &CompanionClient{
		address: address,
		port:    port,
		events:  eventConfigs,
	}
}

// Sends an event to Companion if enabled and the event is configured.
func (client *CompanionClient) SendEvent(event CompanionEvent) {
	if !client.IsEnabled() {
		return
	}

	config, exists := client.events[event]
	if !exists || config.Page == 0 {
		// Event not configured or has invalid coordinates
		return
	}

	command := fmt.Sprintf("LOCATION %d/%d/%d PRESS\n", config.Page, config.Row, config.Column)
	client.sendCommand(command)
}

// Connects to Companion and executes the given command.
func (client *CompanionClient) sendCommand(command string) {
	if !client.IsEnabled() {
		return
	}

	address := fmt.Sprintf("%s:%d", client.address, client.port)
	conn, err := net.DialTimeout("tcp", address, companionConnectTimeoutMs*time.Millisecond)
	if err != nil {
		log.Printf("Failed to connect to Companion at %s: %v", address, err)
		return
	}
	defer conn.Close()

	_, err = fmt.Fprint(conn, command)
	if err != nil {
		log.Printf("Failed to send command '%s' to Companion at %s: %v", command, address, err)
	}
}

// IsEnabled returns whether the Companion client is enabled (address is not blank)
func (client *CompanionClient) IsEnabled() bool {
	return client.address != ""
}

// GetEventConfig returns the configuration for a specific event
func (client *CompanionClient) GetEventConfig(event CompanionEvent) (CompanionEventConfig, bool) {
	config, exists := client.events[event]
	return config, exists
}
