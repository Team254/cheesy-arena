// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2017 rotor elements.

package game

import (
	"time"
)

type RotorSet struct {
	AutoRotors int
	Rotors     int
}

// Updates the internal counting state of the rotors given the current state of the sensors.
func (rotorSet *RotorSet) UpdateState(rotors [4]bool, matchStartTime, currentTime time.Time) {
	autoValidityCutoff := matchStartTime.Add(time.Duration(MatchTiming.AutoDurationSec) * time.Second)
	teleopValidityCutoff := autoValidityCutoff.Add(time.Duration(MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second)

	if currentTime.After(matchStartTime) {
		if currentTime.Before(autoValidityCutoff) {
			if rotorSet.AutoRotors == 0 && rotors[0] {
				rotorSet.AutoRotors++
			}
			if rotorSet.AutoRotors == 1 && rotors[1] {
				rotorSet.AutoRotors++
			}
		} else if currentTime.Before(teleopValidityCutoff) {
			if rotorSet.totalRotors() == 0 && rotors[0] {
				rotorSet.Rotors++
			}
			if rotorSet.totalRotors() == 1 && rotors[1] {
				rotorSet.Rotors++
			}
			if rotorSet.totalRotors() == 2 && rotors[2] {
				rotorSet.Rotors++
			}
			if rotorSet.totalRotors() == 3 && rotors[3] {
				rotorSet.Rotors++
			}
		}
	}
}

func (rotorSet *RotorSet) totalRotors() int {
	return rotorSet.AutoRotors + rotorSet.Rotors
}
