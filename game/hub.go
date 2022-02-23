// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2022 Hub element.

package game

import (
	"time"
)

type Hub struct {
	AutoCargoLower   [4]int
	AutoCargoUpper   [4]int
	TeleopCargoLower [4]int
	TeleopCargoUpper [4]int
}

type HubQuadrant int

const (
	BlueQuadrant HubQuadrant = iota
	FarQuadrant
	NearQuadrant
	RedQuadrant
)

// Updates the internal counting state of the hub given the current state of the hardware counts. Allows the score to
// accumulate before the match, since the counters will be reset in hardware upon match start.
func (hub *Hub) UpdateState(lowerHubCounts [4]int, upperHubCounts [4]int, matchStartTime, currentTime time.Time) {
	autoValidityDuration := GetDurationToAutoEnd() + hubAutoGracePeriodSec*time.Second
	autoValidityCutoff := matchStartTime.Add(autoValidityDuration)
	teleopValidityDuration := GetDurationToTeleopEnd() + HubTeleopGracePeriodSec*time.Second
	teleopValidityCutoff := matchStartTime.Add(teleopValidityDuration)

	if currentTime.Before(autoValidityCutoff) {
		for i := 0; i < 4; i++ {
			hub.AutoCargoLower[i] = lowerHubCounts[i]
			hub.AutoCargoUpper[i] = upperHubCounts[i]
		}
	} else if currentTime.Before(teleopValidityCutoff) {
		for i := 0; i < 4; i++ {
			hub.TeleopCargoLower[i] = lowerHubCounts[i] - hub.AutoCargoLower[i]
			hub.TeleopCargoUpper[i] = upperHubCounts[i] - hub.AutoCargoUpper[i]
		}
	}
}
