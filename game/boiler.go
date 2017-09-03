// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2017 boiler element.

package game

import (
	"time"
)

const (
	BoilerAutoGracePeriodSec   = 5
	BoilerTeleopGracePeriodSec = 5
)

type Boiler struct {
	AutoFuelLow  int
	AutoFuelHigh int
	FuelLow      int
	FuelHigh     int
}

// Updates the internal counting state of the boiler given the current state of the hardware counts. Allows the score to
// accumulate before the match, since the counters will be reset in hardware.
func (boiler *Boiler) UpdateState(lowCount, highCount int, matchStartTime, currentTime time.Time) {
	autoValidityDuration := time.Duration(MatchTiming.AutoDurationSec+BoilerAutoGracePeriodSec) * time.Second
	autoValidityCutoff := matchStartTime.Add(autoValidityDuration)
	teleopValidityDuration := time.Duration(MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec+BoilerTeleopGracePeriodSec) * time.Second
	teleopValidityCutoff := matchStartTime.Add(teleopValidityDuration)

	if currentTime.Before(autoValidityCutoff) {
		boiler.AutoFuelLow = lowCount
		boiler.AutoFuelHigh = highCount
		boiler.FuelLow = 0
		boiler.FuelHigh = 0
	} else if currentTime.Before(teleopValidityCutoff) {
		boiler.FuelLow = lowCount
		boiler.FuelHigh = highCount
	}
}
