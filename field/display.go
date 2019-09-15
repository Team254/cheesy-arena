// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing and methods for controlling a remote web display.

package field

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	minDisplayId = 100
)

type DisplayType int

const (
	InvalidDisplay DisplayType = iota
	PlaceholderDisplay
	AllianceStationDisplay
	AnnouncerDisplay
	AudienceDisplay
	FieldMonitorDisplay
	PitDisplay
	QueueingDisplay
	TwitchStreamDisplay
)

var DisplayTypeNames = map[DisplayType]string{
	PlaceholderDisplay:     "Placeholder",
	AllianceStationDisplay: "Alliance Station",
	AnnouncerDisplay:       "Announcer",
	AudienceDisplay:        "Audience",
	FieldMonitorDisplay:    "Field Monitor",
	PitDisplay:             "Pit",
	QueueingDisplay:        "Queueing",
	TwitchStreamDisplay:    "Twitch Stream",
}

var displayTypePaths = map[DisplayType]string{
	PlaceholderDisplay:     "/display",
	AllianceStationDisplay: "/displays/alliance_station",
	AnnouncerDisplay:       "/displays/announcer",
	AudienceDisplay:        "/displays/audience",
	FieldMonitorDisplay:    "/displays/field_monitor",
	PitDisplay:             "/displays/pit",
	QueueingDisplay:        "/displays/queueing",
	TwitchStreamDisplay:    "/displays/twitch",
}

var displayRegistryMutex sync.Mutex

type Display struct {
	Id              string
	Nickname        string
	Type            DisplayType
	Configuration   map[string]string
	IpAddress       string
	ConnectionCount int
}

// Parses the given display URL path and query string to extract the configuration.
func DisplayFromUrl(path string, query map[string][]string) (*Display, error) {
	if _, ok := query["displayId"]; !ok {
		return nil, fmt.Errorf("Display ID not present in request.")
	}

	var display Display
	display.Id = query["displayId"][0]
	if nickname, ok := query["nickname"]; ok {
		display.Nickname, _ = url.QueryUnescape(nickname[0])
	}

	// Determine type from the websocket connection URL. This way of doing it isn't super efficient, but it's not really
	// a concern since it should happen relatively infrequently.
	for displayType, displayPath := range displayTypePaths {
		if path == displayPath+"/websocket" {
			display.Type = displayType
			break
		}
	}
	if display.Type == InvalidDisplay {
		return nil, fmt.Errorf("Could not determine display type from path %s.", path)
	}

	// Put any remaining query parameters into the per-type configuration map.
	display.Configuration = make(map[string]string)
	for key, value := range query {
		if key != "displayId" && key != "nickname" {
			display.Configuration[key], _ = url.QueryUnescape(value[0])
		}
	}

	return &display, nil
}

// Returns the URL string for the given display that includes all of its configuration parameters.
func (display *Display) ToUrl() string {
	var builder strings.Builder
	builder.WriteString(displayTypePaths[display.Type])
	builder.WriteString("?displayId=")
	builder.WriteString(url.QueryEscape(display.Id))
	if display.Nickname != "" {
		builder.WriteString("&nickname=")
		builder.WriteString(url.QueryEscape(display.Nickname))
	}

	// Sort the keys so that the URL generated is deterministic.
	var keys []string
	for key := range display.Configuration {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString("&")
		builder.WriteString(url.QueryEscape(key))
		builder.WriteString("=")
		builder.WriteString(url.QueryEscape(display.Configuration[key]))
	}
	return builder.String()
}

// Returns an unused ID that can be used for a new display.
func (arena *Arena) NextDisplayId() string {
	displayRegistryMutex.Lock()
	defer displayRegistryMutex.Unlock()

	// Loop until we get an ID that isn't already used. This is inefficient if there is a large number of displays, but
	// that should never be the case.
	candidateId := minDisplayId
	for {
		if _, ok := arena.Displays[strconv.Itoa(candidateId)]; !ok {
			return strconv.Itoa(candidateId)
		}
		candidateId++
	}
}

// Adds the given display to the arena registry and triggers a notification.
func (arena *Arena) RegisterDisplay(display *Display) {
	displayRegistryMutex.Lock()
	defer displayRegistryMutex.Unlock()

	existingDisplay, ok := arena.Displays[display.Id]
	if ok && display.Type == PlaceholderDisplay {
		// Don't rewrite the registered configuration if the new one is a placeholder -- if it is reconnecting after a
		// restart, it should adopt the existing configuration.
		arena.Displays[display.Id].ConnectionCount++
	} else {
		if ok {
			display.ConnectionCount = existingDisplay.ConnectionCount + 1
		} else {
			display.ConnectionCount = 1
		}
		arena.Displays[display.Id] = display
	}
	arena.DisplayConfigurationNotifier.Notify()
}

// Updates the given display in the arena registry. Triggers a notification if the display configuration changed.
func (arena *Arena) UpdateDisplay(display *Display) error {
	displayRegistryMutex.Lock()
	defer displayRegistryMutex.Unlock()

	existingDisplay, ok := arena.Displays[display.Id]
	if !ok {
		return fmt.Errorf("Display %s doesn't exist.", display.Id)
	}
	display.ConnectionCount = existingDisplay.ConnectionCount
	if !reflect.DeepEqual(existingDisplay, display) {
		arena.Displays[display.Id] = display
		arena.DisplayConfigurationNotifier.Notify()
	}
	return nil
}

// Marks the given display as having disconnected in the arena registry and triggers a notification.
func (arena *Arena) MarkDisplayDisconnected(display *Display) {
	displayRegistryMutex.Lock()
	defer displayRegistryMutex.Unlock()

	if existingDisplay, ok := arena.Displays[display.Id]; ok {
		if existingDisplay.Type == PlaceholderDisplay && existingDisplay.Nickname == "" &&
			len(existingDisplay.Configuration) == 0 {
			// If the display is an unconfigured placeholder, just remove it entirely to prevent clutter.
			delete(arena.Displays, existingDisplay.Id)
		} else {
			existingDisplay.ConnectionCount -= 1
		}
		arena.DisplayConfigurationNotifier.Notify()
	}
}
