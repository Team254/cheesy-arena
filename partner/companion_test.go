// Copyright 2025 Team 254. All Rights Reserved.
// Author: kyle@team2481.com (Kyle Waremburg)
//
// Tests for the Companion client.

package partner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCompanionClient(t *testing.T) {
	// Test with disabled client (blank address)
	client := NewCompanionClient("", 51234, nil)
	assert.Equal(t, "", client.address)
	assert.Equal(t, 51234, client.port)
	assert.Nil(t, client.events)

	// Test with enabled client and event configs
	eventConfigs := map[CompanionEvent]CompanionEventConfig{
		EventMatchStart: {Page: 1, Row: 2, Column: 3},
		EventMatchEnd:   {Page: 2, Row: 3, Column: 4},
	}
	client = NewCompanionClient("192.168.1.100", 51235, eventConfigs)
	assert.Equal(t, "192.168.1.100", client.address)
	assert.Equal(t, 51235, client.port)
	assert.Equal(t, eventConfigs, client.events)
}

func TestCompanionClient_IsEnabled(t *testing.T) {
	client := NewCompanionClient("", 51234, nil)
	assert.False(t, client.IsEnabled())

	client = NewCompanionClient("127.0.0.1", 51234, nil)
	assert.True(t, client.IsEnabled())
}

func TestCompanionClient_GetEventConfig(t *testing.T) {
	eventConfigs := map[CompanionEvent]CompanionEventConfig{
		EventMatchStart: {Page: 1, Row: 2, Column: 3},
		EventMatchEnd:   {Page: 2, Row: 3, Column: 4},
	}
	client := NewCompanionClient("127.0.0.1", 51234, eventConfigs)

	// Test existing event
	config, exists := client.GetEventConfig(EventMatchStart)
	assert.True(t, exists)
	assert.Equal(t, 1, config.Page)
	assert.Equal(t, 2, config.Row)
	assert.Equal(t, 3, config.Column)

	// Test non-existing event
	config, exists = client.GetEventConfig(EventTeleopStart)
	assert.False(t, exists)
	assert.Equal(t, CompanionEventConfig{}, config)
}

func TestCompanionClient_SendEvent_Disabled(t *testing.T) {
	// Test that disabled client doesn't send events
	client := NewCompanionClient("", 51234, nil)

	// This should not panic or cause errors when client is disabled
	client.SendEvent(EventMatchStart)
}

func TestCompanionClient_SendEvent_UnconfiguredEvent(t *testing.T) {
	// Test that unconfigured events are ignored
	eventConfigs := map[CompanionEvent]CompanionEventConfig{
		EventMatchStart: {Page: 1, Row: 2, Column: 3},
	}
	client := NewCompanionClient("127.0.0.1", 51234, eventConfigs)

	// This should not panic or cause errors for unconfigured events
	client.SendEvent(EventTeleopStart)
}

func TestCompanionClient_SendEvent_InvalidConfig(t *testing.T) {
	// Test that events with invalid coordinates (0 values) are ignored
	eventConfigs := map[CompanionEvent]CompanionEventConfig{
		EventMatchStart:  {Page: 0, Row: 2, Column: 3}, // Invalid page
		EventMatchEnd:    {Page: 1, Row: 0, Column: 3}, // Invalid row
		EventTeleopStart: {Page: 1, Row: 2, Column: 0}, // Invalid column
	}
	client := NewCompanionClient("127.0.0.1", 51234, eventConfigs)

	// These should not panic or cause errors for invalid configurations
	client.SendEvent(EventMatchStart)
	client.SendEvent(EventMatchEnd)
	client.SendEvent(EventTeleopStart)
}
