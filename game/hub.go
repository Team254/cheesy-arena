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
	if hub.isShiftActive(shift) || !activeOnly {
		return hub.ShiftCounts[shift]
	}
	return 0
}

// GetActiveShiftTiming returns the amount of time remaining in the current shift if the Hub is active and the duration
// of the current shift. If the Hub is not active, the remaining time is zero. If the match is not in a valid shift,
// both values are zero.
func (hub *Hub) GetActiveShiftTiming(matchStartTime, currentTime time.Time) (time.Duration, time.Duration) {
	shiftStartTime := matchStartTime
	shiftEndTime := matchStartTime.Add(GetDurationToAutoEnd())
	for _, shift := range []Shift{ShiftAuto, ShiftTransition, Shift1, Shift2, Shift3, Shift4, ShiftEndgame} {
		shiftDuration := shiftEndTime.Sub(shiftStartTime)
		if !currentTime.Before(shiftStartTime) && currentTime.Before(shiftEndTime) {
			if hub.isShiftActive(shift) {
				return shiftEndTime.Sub(currentTime), shiftDuration
			}
			return 0, shiftDuration
		}
		shiftStartTime = shiftEndTime
		switch shift {
		case ShiftAuto:
			shiftStartTime = matchStartTime.Add(GetDurationToTeleopStart())
			shiftEndTime = shiftStartTime.Add(time.Duration(MatchTiming.TransitionShiftDurationSec) * time.Second)
		case ShiftTransition, Shift1, Shift2, Shift3:
			shiftEndTime = shiftEndTime.Add(time.Duration(MatchTiming.ShiftDurationSec) * time.Second)
		case Shift4:
			shiftEndTime = matchStartTime.Add(GetDurationToTeleopEnd())
		default:
			shiftEndTime = shiftStartTime
		}
	}
	return 0, 0
}

// isShiftActive returns true if the Hub is active during the given shift.
func (hub *Hub) isShiftActive(shift Shift) bool {
	switch shift {
	case ShiftAuto, ShiftTransition, ShiftEndgame:
		return true
	case Shift1, Shift3:
		return !hub.WonAuto
	case Shift2, Shift4:
		return hub.WonAuto
	default:
		return false
	}
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
