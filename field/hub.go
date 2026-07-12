// Copyright 2026 Team 254. All Rights Reserved.
//
// Model representing the color status of each physical Hub scoring target on the field. This is used to drive
// tablet displays mounted at each hub that indicate its current state (active for red, active for blue, or under
// manual control for field-reset signaling).

package field

import "fmt"

// HubColor represents the current lighting state of a hub.
type HubColor string

const (
	HubOff    HubColor = "off"
	HubRed    HubColor = "red"
	HubBlue   HubColor = "blue"
	HubGreen  HubColor = "green"
	HubPurple HubColor = "purple"
)

var validHubColors = map[HubColor]struct{}{
	HubOff:    {},
	HubRed:    {},
	HubBlue:   {},
	HubGreen:  {},
	HubPurple: {},
}

// SetHubColor updates the color for the given hub ID and notifies any subscribed hub displays.
func (arena *Arena) SetHubColor(hubId string, color HubColor) error {
	if _, ok := validHubColors[color]; !ok {
		return fmt.Errorf("invalid hub color: %q", color)
	}
	if arena.HubColors == nil {
		arena.HubColors = make(map[string]HubColor)
	}
	arena.HubColors[hubId] = color
	arena.HubColorNotifier.Notify()
	return nil
}

// GetHubColor returns the current color for the given hub ID, defaulting to HubOff if it has never been set.
func (arena *Arena) GetHubColor(hubId string) HubColor {
	if color, ok := arena.HubColors[hubId]; ok {
		return color
	}
	return HubOff
}
