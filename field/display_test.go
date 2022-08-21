// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	display, _ = DisplayFromUrl("/displays/rankings/websocket", query)
	assert.Equal(t, RankingsDisplay, display.Type)

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
	display := &Display{DisplayConfiguration: DisplayConfiguration{Id: "254", Nickname: "Test Nickname",
		Type: RankingsDisplay, Configuration: map[string]string{"f": "1", "z": "#fff", "a": "3", "c": "4"}}}
	assert.Equal(t, "/displays/rankings?displayId=254&nickname=Test+Nickname&a=3&c=4&f=1&z=%23fff", display.ToUrl())
}

func TestNextDisplayId(t *testing.T) {
	arena := setupTestArena(t)

	assert.Equal(t, "100", arena.NextDisplayId())

	displayConfig := &DisplayConfiguration{Id: "100"}
	arena.RegisterDisplay(displayConfig, "")
	assert.Equal(t, "101", arena.NextDisplayId())
}

func TestDisplayRegisterUnregister(t *testing.T) {
	arena := setupTestArena(t)

	displayConfig := &DisplayConfiguration{Id: "254", Nickname: "Placeholder", Type: PlaceholderDisplay,
		Configuration: map[string]string{}}
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Placeholder", arena.Displays["254"].DisplayConfiguration.Nickname)
		assert.Equal(t, PlaceholderDisplay, arena.Displays["254"].DisplayConfiguration.Type)
		assert.Equal(t, 1, arena.Displays["254"].ConnectionCount)
		assert.Equal(t, "1.2.3.4", arena.Displays["254"].IpAddress)
	}
	notifier := arena.Displays["254"].Notifier

	// Register a second instance of the same display.
	displayConfig2 := &DisplayConfiguration{Id: "254", Nickname: "Rankings", Type: RankingsDisplay,
		Configuration: map[string]string{}}
	arena.RegisterDisplay(displayConfig2, "2.3.4.5")
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Rankings", arena.Displays["254"].DisplayConfiguration.Nickname)
		assert.Equal(t, RankingsDisplay, arena.Displays["254"].DisplayConfiguration.Type)
		assert.Equal(t, 2, arena.Displays["254"].ConnectionCount)
		assert.Equal(t, "2.3.4.5", arena.Displays["254"].IpAddress)
		assert.Same(t, notifier, arena.Displays["254"].Notifier)
	}

	// Register a second display.
	displayConfig3 := &DisplayConfiguration{Id: "148", Type: FieldMonitorDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(displayConfig3, "3.4.5.6")
	if assert.Contains(t, arena.Displays, "148") {
		assert.Equal(t, 1, arena.Displays["148"].ConnectionCount)
	}

	// Update the first display.
	displayConfig4 := DisplayConfiguration{Id: "254", Nickname: "Alliance", Type: AllianceStationDisplay,
		Configuration: map[string]string{"station": "B2"}}
	arena.UpdateDisplay(displayConfig4)
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, "Alliance", arena.Displays["254"].DisplayConfiguration.Nickname)
		assert.Equal(t, AllianceStationDisplay, arena.Displays["254"].DisplayConfiguration.Type)
		assert.Equal(t, 2, arena.Displays["254"].ConnectionCount)
	}

	// Disconnect both displays.
	arena.MarkDisplayDisconnected(displayConfig.Id)
	arena.MarkDisplayDisconnected(displayConfig3.Id)
	if assert.Contains(t, arena.Displays, "148") {
		assert.Equal(t, 0, arena.Displays["148"].ConnectionCount)
	}
	if assert.Contains(t, arena.Displays, "254") {
		assert.Equal(t, 1, arena.Displays["254"].ConnectionCount)
	}
}

func TestDisplayUpdateError(t *testing.T) {
	arena := setupTestArena(t)

	displayConfig := DisplayConfiguration{Id: "254", Configuration: map[string]string{}}
	err := arena.UpdateDisplay(displayConfig)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "doesn't exist")
	}
}

func TestDisplayPurge(t *testing.T) {
	arena := setupTestArena(t)

	// Unnamed placeholder gets immediately purged upon disconnection.
	displayConfig := &DisplayConfiguration{Id: "254", Type: PlaceholderDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	assert.Contains(t, arena.Displays, "254")
	arena.MarkDisplayDisconnected(displayConfig.Id)
	assert.NotContains(t, arena.Displays, "254")

	// Named placeholder does not get immediately purged upon disconnection.
	displayConfig.Nickname = "Bob"
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	assert.Contains(t, arena.Displays, "254")
	arena.MarkDisplayDisconnected(displayConfig.Id)
	assert.Contains(t, arena.Displays, "254")

	// Unnamed configured displayConfig does not get immediately purged upon disconnection.
	displayConfig = &DisplayConfiguration{Id: "1114", Type: FieldMonitorDisplay, Configuration: map[string]string{}}
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	assert.Contains(t, arena.Displays, "1114")
	arena.MarkDisplayDisconnected(displayConfig.Id)
	assert.Contains(t, arena.Displays, "1114")
	arena.purgeDisconnectedDisplays()
	assert.Contains(t, arena.Displays, "1114")

	// Unnamed configured displayConfig gets purged by periodic task.
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	assert.Contains(t, arena.Displays, "1114")
	arena.MarkDisplayDisconnected(displayConfig.Id)
	arena.Displays["1114"].lastConnectedTime = time.Now().Add(-displayPurgeTtlMin * time.Minute)
	arena.purgeDisconnectedDisplays()
	assert.NotContains(t, arena.Displays, "1114")

	// Named configured displayConfig does not get purged by periodic task.
	displayConfig.Nickname = "Brunhilda"
	arena.RegisterDisplay(displayConfig, "1.2.3.4")
	assert.Contains(t, arena.Displays, "1114")
	arena.MarkDisplayDisconnected(displayConfig.Id)
	arena.Displays["1114"].lastConnectedTime = time.Now().Add(-displayPurgeTtlMin * time.Minute)
	arena.purgeDisconnectedDisplays()
	assert.Contains(t, arena.Displays, "1114")
}
