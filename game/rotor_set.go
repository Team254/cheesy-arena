// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2017 rotor elements.

package game

import (
	"time"
)

const rotorGearToothCount = 15

type RotorSet struct {
	AutoRotors     int
	Rotors         int
	counterOffsets [3]int
}

// Updates the internal counting state of the rotors given the current state of the sensors.
func (rotorSet *RotorSet) UpdateState(rotor1 bool, otherRotors [3]int, matchStartTime, currentTime time.Time) {
	autoValidityCutoff := matchStartTime.Add(time.Duration(MatchTiming.AutoDurationSec) * time.Second)
	teleopValidityCutoff := autoValidityCutoff.Add(time.Duration(MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second)

	if currentTime.After(matchStartTime) {
		if currentTime.Before(autoValidityCutoff) {
			if rotorSet.AutoRotors == 0 && rotor1 {
				rotorSet.AutoRotors++
				rotorSet.counterOffsets[0] = otherRotors[0]
			}
			if rotorSet.AutoRotors == 1 && otherRotors[0]-rotorSet.counterOffsets[0] >= rotorGearToothCount {
				rotorSet.AutoRotors++
				rotorSet.counterOffsets[1] = otherRotors[1]
			}
		} else if currentTime.Before(teleopValidityCutoff) {
			if rotorSet.totalRotors() == 0 && rotor1 {
				rotorSet.Rotors++
				rotorSet.counterOffsets[0] = otherRotors[0]
			}
			if rotorSet.totalRotors() == 1 && otherRotors[0]-rotorSet.counterOffsets[0] >= rotorGearToothCount {
				rotorSet.Rotors++
				rotorSet.counterOffsets[1] = otherRotors[1]
			}
			if rotorSet.totalRotors() == 2 && otherRotors[1]-rotorSet.counterOffsets[1] >= rotorGearToothCount {
				rotorSet.Rotors++
				rotorSet.counterOffsets[2] = otherRotors[2]
			}
			if rotorSet.totalRotors() == 3 && otherRotors[2]-rotorSet.counterOffsets[2] >= rotorGearToothCount {
				rotorSet.Rotors++
			}
		}
	}
}

func (rotorSet *RotorSet) totalRotors() int {
	return rotorSet.AutoRotors + rotorSet.Rotors
}
