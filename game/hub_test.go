// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var hubMatchStartTime = time.Unix(10, 0)

func TestHub_UpdateState(t *testing.T) {
	var hub Hub
	assertHubShiftCounts(t, &hub, [ShiftCount]int{})

	// Fuel scored before the match should not be counted.
	hub.UpdateState(4, hubMatchStartTime, hubTimeAfterStart(-1))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{})

	// Fuel is counted in auto through the end of the grace period.
	hub.UpdateState(0, hubMatchStartTime, hubTimeAfterStart(0))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{})
	hub.UpdateState(3, hubMatchStartTime, hubTimeAfterStart(5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{3, 0, 0, 0, 0, 0, 0})
	hub.UpdateState(4, hubMatchStartTime, hubTimeAfterStart(22.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 0, 0, 0, 0, 0, 0})

	// New Fuel is attributed to the previous shift until the grace period expires.
	hub.UpdateState(7, hubMatchStartTime, hubTimeAfterStart(35.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 0, 0, 0, 0, 0})
	hub.UpdateState(10, hubMatchStartTime, hubTimeAfterStart(60.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 0, 0, 0})
	hub.UpdateState(14, hubMatchStartTime, hubTimeAfterStart(85.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 4, 0, 0, 0})
	hub.UpdateState(19, hubMatchStartTime, hubTimeAfterStart(110.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 4, 5, 0, 0})
	hub.UpdateState(25, hubMatchStartTime, hubTimeAfterStart(135.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 4, 5, 6, 0})

	// Endgame is counted until the end-of-match grace period, after which new Fuel is ignored.
	hub.UpdateState(32, hubMatchStartTime, hubTimeAfterStart(162.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 4, 5, 6, 7})
	hub.UpdateState(36, hubMatchStartTime, hubTimeAfterStart(166.1))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 4, 5, 6, 7})
}

func TestHub_UpdateStateDuringExtendedPauseCountsAsTransitionShift(t *testing.T) {
	originalPauseDurationSec := MatchTiming.PauseDurationSec
	defer func() {
		MatchTiming.PauseDurationSec = originalPauseDurationSec
	}()
	MatchTiming.PauseDurationSec = ScoringGracePeriodSec + 4

	var hub Hub
	hub.UpdateState(5, hubMatchStartTime, hubTimeAfterStart(20.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{5, 0, 0, 0, 0, 0, 0})

	// Even though teleop has not started yet, Fuel scored after the auto grace period is transition Fuel.
	hub.UpdateState(8, hubMatchStartTime, hubTimeAfterStart(24.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{5, 3, 0, 0, 0, 0, 0})
}

func TestHub_GetTeleopActiveFuelCount(t *testing.T) {
	hub := Hub{
		WonAuto:     false,
		ShiftCounts: [ShiftCount]int{9, 10, 20, 30, 40, 50, 60},
	}
	assert.Equal(t, 130, hub.GetTeleopActiveFuelCount())

	hub.WonAuto = true
	assert.Equal(t, 150, hub.GetTeleopActiveFuelCount())
}

func TestHub_GetShiftActiveCount(t *testing.T) {
	hub := Hub{
		WonAuto:     false,
		ShiftCounts: [ShiftCount]int{1, 2, 3, 4, 5, 6, 7},
	}
	assert.Equal(t, 1, hub.GetShiftCount(ShiftAuto, false))
	assert.Equal(t, 1, hub.GetShiftCount(ShiftAuto, true))
	assert.Equal(t, 2, hub.GetShiftCount(ShiftTransition, false))
	assert.Equal(t, 2, hub.GetShiftCount(ShiftTransition, true))
	assert.Equal(t, 3, hub.GetShiftCount(Shift1, false))
	assert.Equal(t, 3, hub.GetShiftCount(Shift1, true))
	assert.Equal(t, 4, hub.GetShiftCount(Shift2, false))
	assert.Equal(t, 0, hub.GetShiftCount(Shift2, true))
	assert.Equal(t, 5, hub.GetShiftCount(Shift3, false))
	assert.Equal(t, 5, hub.GetShiftCount(Shift3, true))
	assert.Equal(t, 6, hub.GetShiftCount(Shift4, false))
	assert.Equal(t, 0, hub.GetShiftCount(Shift4, true))
	assert.Equal(t, 7, hub.GetShiftCount(ShiftEndgame, false))
	assert.Equal(t, 7, hub.GetShiftCount(ShiftEndgame, true))

	hub.WonAuto = true
	assert.Equal(t, 0, hub.GetShiftCount(Shift1, true))
	assert.Equal(t, 4, hub.GetShiftCount(Shift2, true))
	assert.Equal(t, 0, hub.GetShiftCount(Shift3, true))
	assert.Equal(t, 6, hub.GetShiftCount(Shift4, true))
}

func assertHubShiftCounts(t *testing.T, hub *Hub, expectedShiftCounts [ShiftCount]int) {
	for shift := ShiftAuto; shift < ShiftCount; shift++ {
		assert.Equal(t, expectedShiftCounts[shift], hub.ShiftCounts[shift])
	}
}

func hubTimeAfterStart(sec float32) time.Time {
	return hubMatchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}
