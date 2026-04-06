// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2026 Hub element.

package game

import (
	"time"
)

type Hub struct {
	WonAuto     bool
	ShiftCounts [ShiftCount]int
}

// Shift represents a distinct period during the match when Fuel is scored (and tracked separately).
type Shift int

const (
	ShiftAuto Shift = iota
	ShiftTransition
	Shift1
	Shift2
	Shift3
	Shift4
	ShiftEndgame
	ShiftCount
)

// UpdateState updates the internal counting state of the Hub given the current state of the hardware count and the
// match time.
func (hub *Hub) UpdateState(count int, matchStartTime, currentTime time.Time) {
	if currentTime.Before(matchStartTime) {
		return
	}

	shift, ok := getCurrentShift(matchStartTime, currentTime)
	if !ok {
		return
	}

	var existingCount int
	for _, shiftCount := range hub.ShiftCounts {
		existingCount += shiftCount
	}
	newFuel := count - existingCount
	if newFuel <= 0 {
		return
	}

	hub.ShiftCounts[shift] += newFuel
}

// GetTeleopActiveFuelCount returns the number of Fuel scored during the teleop period when the Hub was active.
func (hub *Hub) GetTeleopActiveFuelCount() int {
	var count int
	for shift := ShiftTransition; shift < ShiftCount; shift++ {
		count += hub.GetShiftCount(shift, true)
	}
	return count
}

// GetShiftActiveCount returns the number of Fuel scored during the given shift if the Hub was active, or zero if the
// Hub was not active.
func (hub *Hub) GetShiftCount(shift Shift, activeOnly bool) int {
	switch shift {
	case ShiftAuto, ShiftTransition, ShiftEndgame:
		return hub.ShiftCounts[shift]
	case Shift1, Shift3:
		if !hub.WonAuto || !activeOnly {
			return hub.ShiftCounts[shift]
		}
	case Shift2, Shift4:
		if hub.WonAuto || !activeOnly {
			return hub.ShiftCounts[shift]
		}
	default:
	}
	return 0
}

// getCurrentShift returns the current shift based on the match time, and a boolean indicating whether the time
// corresponds to a valid shift.
func getCurrentShift(matchStartTime, currentTime time.Time) (Shift, bool) {
	gracePeriod := time.Duration(ScoringGracePeriodSec) * time.Second
	if currentTime.Before(matchStartTime.Add(GetDurationToAutoEnd() + gracePeriod)) {
		return ShiftAuto, true
	}

	teleopStartTime := matchStartTime.Add(GetDurationToTeleopStart())
	shiftEndTime := teleopStartTime.Add(time.Duration(MatchTiming.TransitionShiftDurationSec) * time.Second)
	if currentTime.Before(shiftEndTime.Add(gracePeriod)) {
		return ShiftTransition, true
	}

	for shift := Shift1; shift <= Shift4; shift++ {
		shiftEndTime = shiftEndTime.Add(time.Duration(MatchTiming.ShiftDurationSec) * time.Second)
		if currentTime.Before(shiftEndTime.Add(gracePeriod)) {
			return shift, true
		}
	}

	teleopEndTime := matchStartTime.Add(GetDurationToTeleopEnd())
	if currentTime.Before(teleopEndTime.Add(gracePeriod)) {
		return ShiftEndgame, true
	}

	return ShiftCount, false
}
