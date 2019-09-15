// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplayFromUrl(t *testing.T) {
	query := map[string][]string{}
	display, err := DisplayFromUrl("/display", query)
	assert.Nil(t, display)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "ID not present")
	}

	// Test the various types.
	query["displayId"] = []string{"123"}
	display, err = DisplayFromUrl("/blorpy", query)
	assert.Nil(t, display)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Could not determine display type")
	}
	display, _ = DisplayFromUrl("/display/websocket", query)
	assert.Equal(t, PlaceholderDisplay, display.Type)
	display, _ = DisplayFromUrl("/displays/alliance_station/websocket", query)
	assert.Equal(t, AllianceStationDisplay, display.Type)
	display, _ = DisplayFromUrl("/displays/announcer/websocket", query)
	assert.Equal(t, AnnouncerDisplay, display.Type)
	display, _ = DisplayFromUrl("/displays/audience/websocket", query)
	assert.Equal(t, AudienceDisplay, display.Type)
	display, _ = DisplayFromUrl("/displays/field_monitor/websocket", query)
	assert.Equal(t, FieldMonitorDisplay, display.Type)
	display, _ = DisplayFromUrl("/displays/pit/websocket", query)
	assert.Equal(t, PitDisplay, display.Type)

	// Test the nickname and arbitrary parameters.
	query["nickname"] = []string{"Test Nickname"}
	query["key1"] = []string{"value1"}
	query["key2"] = []string{"value2"}
	query["color"] = []string{"%230f0"}
	display, _ = DisplayFromUrl("/display/websocket", query)
	assert.Equal(t, "Test Nickname", display.Nickname)
	if assert.Equal(t, 3, len(display.Configuration)) {
		assert.Equal(t, "value1", display.Configuration["key1"])
		assert.Equal(t, "value2", display.Configuration["key2"])
		assert.Equal(t, "#0f0", display.Configuration["color"])
	}
}

func TestDisplayToUrl(t *testing.T) {
	display := &Display{Id: "254", Nickname: "Test Nickname", Type: PitDisplay,
		Configuration: map[string]string{"f": "1", "z": "#fff", "a": "3", "c": "4"}}
	assert.Equal(t, "/displays/pit?displayId=254&nickname=Test+Nickname&a=3&c=4&f=1&z=%23fff", display.ToUrl())
}

func TestNextDisplayId(t *testing.T) {
	arena := setupTestArena(t)

	assert.Equal(t, "100", arena.NextDisplayId())

	display := &Display{Id: "100"}
	arena.RegisterDisplay(display)
	assert.Equal(t, "101", arena.NextDisplayId())
}

func TestDisplayRegisterUnregister(t *testing.T) {
	arena := setupTestArena(t)

	display := &Display{Id: "254", Nickname: "Placeholder", Type: PlaceholderDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(display)
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Placeholder", arena.Displays["254"].Nickname)
		assert.Equal(t, PlaceholderDisplay, arena.Displays["254"].Type)
		assert.Equal(t, 1, arena.Displays["254"].ConnectionCount)
	}

	// Register a second instance of the same display.
	display2 := &Display{Id: "254", Nickname: "Pit", Type: PitDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(display2)
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Pit", arena.Displays["254"].Nickname)
		assert.Equal(t, PitDisplay, arena.Displays["254"].Type)
		assert.Equal(t, 2, arena.Displays["254"].ConnectionCount)
	}

	// Register a second display.
	display3 := &Display{Id: "148", Type: FieldMonitorDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(display3)
	if assert.Contains(t, arena.Displays, "148") {
		assert.Equal(t, 1, arena.Displays["148"].ConnectionCount)
	}

	// Update the first display.
	display4 := &Display{Id: "254", Nickname: "Alliance", Type: AllianceStationDisplay,
		Configuration: map[string]string{"station": "B2"}}
	arena.UpdateDisplay(display4)
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Alliance", arena.Displays["254"].Nickname)
		assert.Equal(t, AllianceStationDisplay, arena.Displays["254"].Type)
		assert.Equal(t, 2, arena.Displays["254"].ConnectionCount)
	}

	// Disconnect both displays.
	arena.MarkDisplayDisconnected(display)
	arena.MarkDisplayDisconnected(display3)
	if assert.Contains(t, arena.Displays, "148") {
		assert.Equal(t, 0, arena.Displays["148"].ConnectionCount)
	}
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, 1, arena.Displays["254"].ConnectionCount)
	}
}

func TestDisplayUpdateError(t *testing.T) {
	arena := setupTestArena(t)

	display := &Display{Id: "254", Configuration: map[string]string{}}
	err := arena.UpdateDisplay(display)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "doesn't exist")
	}
}
