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
	assertHubShiftCounts(t, &hub, [ShiftCount]int{3, 0, 0, 0, 0, 0, 0, 0})
	hub.UpdateState(4, hubMatchStartTime, hubTimeAfterStart(22.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 0, 0, 0, 0, 0, 0, 0})

	// New Fuel is attributed to the previous active shift until the grace period expires.
	hub.UpdateState(7, hubMatchStartTime, hubTimeAfterStart(35.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 0, 0, 0, 0, 0, 0})
	hub.UpdateState(10, hubMatchStartTime, hubTimeAfterStart(60.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 0, 0, 0, 0})
	hub.UpdateState(14, hubMatchStartTime, hubTimeAfterStart(85.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 4, 0, 0, 0})
	hub.UpdateState(19, hubMatchStartTime, hubTimeAfterStart(110.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 9, 0, 0, 0})
	hub.UpdateState(25, hubMatchStartTime, hubTimeAfterStart(135.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 9, 0, 6, 0})

	// Endgame counts until the match ends; the after-match grace period gets segregated as post-match Fuel.
	hub.UpdateState(32, hubMatchStartTime, hubTimeAfterStart(162.9))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 9, 0, 13, 0})
	hub.UpdateState(36, hubMatchStartTime, hubTimeAfterStart(164.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 9, 0, 13, 4})
	hub.UpdateState(40, hubMatchStartTime, hubTimeAfterStart(166.1))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{4, 3, 3, 0, 9, 0, 13, 4})
}

func TestHub_UpdateStateDuringExtendedPauseCountsAsTransitionShift(t *testing.T) {
	originalPauseDurationSec := MatchTiming.PauseDurationSec
	defer func() {
		MatchTiming.PauseDurationSec = originalPauseDurationSec
	}()
	MatchTiming.PauseDurationSec = ScoringGracePeriodSec + 4

	var hub Hub
	hub.UpdateState(5, hubMatchStartTime, hubTimeAfterStart(20.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{5, 0, 0, 0, 0, 0, 0, 0})

	// Even though teleop has not started yet, Fuel scored after the auto grace period is transition Fuel.
	hub.UpdateState(8, hubMatchStartTime, hubTimeAfterStart(24.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{5, 3, 0, 0, 0, 0, 0, 0})
}

func TestHub_UpdateStateDoesNotApplyGracePeriodToInactiveShift(t *testing.T) {
	hub := Hub{WonAuto: true}

	// Shift 1 is inactive after winning auto, so Fuel just after it ends should count in active Shift 2.
	hub.UpdateState(3, hubMatchStartTime, hubTimeAfterStart(58.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{0, 0, 0, 3, 0, 0, 0, 0})
	// Shift 3 is also inactive after winning auto, so Fuel just after it ends should count in active Shift 4.
	hub.UpdateState(7, hubMatchStartTime, hubTimeAfterStart(108.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{0, 0, 0, 3, 0, 4, 0, 0})
	// Shift 4 gets a normal grace period before endgame takes over.
	hub.UpdateState(11, hubMatchStartTime, hubTimeAfterStart(133.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{0, 0, 0, 3, 0, 8, 0, 0})

	hub = Hub{WonAuto: false}

	// Shift 2 is inactive after losing auto, so Fuel just after it ends should count in active Shift 3.
	hub.UpdateState(4, hubMatchStartTime, hubTimeAfterStart(83.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{0, 0, 0, 0, 4, 0, 0, 0})

	// Shift 4 is inactive after losing auto, so Fuel just after it ends should count in active endgame.
	hub.UpdateState(9, hubMatchStartTime, hubTimeAfterStart(133.5))
	assertHubShiftCounts(t, &hub, [ShiftCount]int{0, 0, 0, 0, 4, 0, 5, 0})
}

func TestHub_UpdateStateTeleopShiftBoundaryGracePeriods(t *testing.T) {
	testCases := []struct {
		name          string
		wonAuto       bool
		timeSec       float32
		expectedShift Shift
	}{
		{"lost auto, shift 1 gets grace", false, 58.5, Shift1},
		{"lost auto, inactive shift 2 gets no grace", false, 83.5, Shift3},
		{"lost auto, shift 3 gets grace", false, 108.5, Shift3},
		{"lost auto, inactive shift 4 gets no grace", false, 133.5, ShiftEndgame},
		{"won auto, inactive shift 1 gets no grace", true, 58.5, Shift2},
		{"won auto, shift 2 gets grace", true, 83.5, Shift2},
		{"won auto, inactive shift 3 gets no grace", true, 108.5, Shift4},
		{"won auto, shift 4 gets grace", true, 133.5, Shift4},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				hub := Hub{WonAuto: testCase.wonAuto}
				hub.UpdateState(1, hubMatchStartTime, hubTimeAfterStart(testCase.timeSec))
				assert.Equal(t, 1, hub.ShiftCounts[testCase.expectedShift])
				for shift := ShiftAuto; shift < ShiftCount; shift++ {
					if shift != testCase.expectedShift {
						assert.Equal(t, 0, hub.ShiftCounts[shift])
					}
				}
			},
		)
	}
}

func TestHub_GetTeleopActiveFuelCount(t *testing.T) {
	hub := Hub{
		WonAuto:     false,
		ShiftCounts: [ShiftCount]int{9, 10, 20, 30, 40, 50, 60, 70},
	}
	assert.Equal(t, 200, hub.GetTeleopActiveFuelCount())

	hub.WonAuto = true
	assert.Equal(t, 220, hub.GetTeleopActiveFuelCount())
}

func TestHub_GetShiftActiveCount(t *testing.T) {
	hub := Hub{
		WonAuto:     false,
		ShiftCounts: [ShiftCount]int{1, 2, 3, 4, 5, 6, 7, 8},
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
	assert.Equal(t, 8, hub.GetShiftCount(ShiftPostMatch, false))
	assert.Equal(t, 8, hub.GetShiftCount(ShiftPostMatch, true))

	hub.WonAuto = true
	assert.Equal(t, 0, hub.GetShiftCount(Shift1, true))
	assert.Equal(t, 4, hub.GetShiftCount(Shift2, true))
	assert.Equal(t, 0, hub.GetShiftCount(Shift3, true))
	assert.Equal(t, 6, hub.GetShiftCount(Shift4, true))
}

func TestHub_GetActiveShiftTiming(t *testing.T) {
	testCases := []struct {
		name          string
		wonAuto       bool
		timeSec       float32
		remaining     time.Duration
		shiftDuration time.Duration
	}{
		{"before match", false, -1, 0, 0},
		{"auto active", false, 5.1, 14900 * time.Millisecond, 20 * time.Second},
		{"transition active", false, 30.1, 2900 * time.Millisecond, 10 * time.Second},
		{"shift 1 active after lost auto", false, 40.1, 17900 * time.Millisecond, 25 * time.Second},
		{"shift 2 inactive after lost auto", false, 60.1, 0, 25 * time.Second},
		{"shift 1 inactive after won auto", true, 40.1, 0, 25 * time.Second},
		{"shift 2 active after won auto", true, 60.1, 22900 * time.Millisecond, 25 * time.Second},
		{"endgame active", false, 160.1, 2900 * time.Millisecond, 30 * time.Second},
		{"after match", false, 166.1, 0, 0},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				hub := Hub{WonAuto: testCase.wonAuto}
				remaining, shiftDuration := hub.GetActiveShiftTiming(
					hubMatchStartTime,
					hubTimeAfterStart(testCase.timeSec),
				)
				assert.Equal(t, testCase.remaining, remaining)
				assert.Equal(t, testCase.shiftDuration, shiftDuration)
			},
		)
	}
}

func TestHub_GetActiveShiftTimingIgnoresGracePeriodsAndExtendedPause(t *testing.T) {
	originalPauseDurationSec := MatchTiming.PauseDurationSec
	defer func() {
		MatchTiming.PauseDurationSec = originalPauseDurationSec
	}()
	MatchTiming.PauseDurationSec = ScoringGracePeriodSec + 4

	hub := Hub{}
	remaining, shiftDuration := hub.GetActiveShiftTiming(hubMatchStartTime, hubTimeAfterStart(20.1))
	assert.Zero(t, remaining)
	assert.Zero(t, shiftDuration)
	remaining, shiftDuration = hub.GetActiveShiftTiming(hubMatchStartTime, hubTimeAfterStart(24.5))
	assert.Zero(t, remaining)
	assert.Zero(t, shiftDuration)
	remaining, shiftDuration = hub.GetActiveShiftTiming(hubMatchStartTime, hubTimeAfterStart(62.1))
	assert.Zero(t, remaining)
	assert.Equal(t, 25*time.Second, shiftDuration)
	remaining, shiftDuration = hub.GetActiveShiftTiming(hubMatchStartTime, hubTimeAfterStart(168.1))
	assert.Zero(t, remaining)
	assert.Zero(t, shiftDuration)
	remaining, shiftDuration = hub.GetActiveShiftTiming(hubMatchStartTime, hubTimeAfterStart(171.1))
	assert.Zero(t, remaining)
	assert.Zero(t, shiftDuration)
}

func assertHubShiftCounts(t *testing.T, hub *Hub, expectedShiftCounts [ShiftCount]int) {
	for shift := ShiftAuto; shift < ShiftCount; shift++ {
		assert.Equal(t, expectedShiftCounts[shift], hub.ShiftCounts[shift])
	}
}

func hubTimeAfterStart(sec float32) time.Time {
	return hubMatchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}
