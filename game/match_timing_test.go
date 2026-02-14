// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Tests for match timing and hub activation logic.

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentShift(t *testing.T) {
	// Before teleop
	assert.Equal(t, -1, GetCurrentShift(0))
	assert.Equal(t, -1, GetCurrentShift(20)) // End of auto
	assert.Equal(t, -1, GetCurrentShift(22)) // During pause (now 3 seconds)

	// During teleop
	teleopStart := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	assert.Equal(t, 0, GetCurrentShift(teleopStart))
	assert.Equal(t, 0, GetCurrentShift(teleopStart+24))
	assert.Equal(t, 1, GetCurrentShift(teleopStart+25))
	assert.Equal(t, 1, GetCurrentShift(teleopStart+49))
	assert.Equal(t, 2, GetCurrentShift(teleopStart+50))
	assert.Equal(t, 5, GetCurrentShift(teleopStart+139))

	// After teleop
	assert.Equal(t, -1, GetCurrentShift(teleopStart+140))
}

func TestIsRedHubActive(t *testing.T) {
	teleopStart := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEnd := teleopStart + float64(MatchTiming.TeleopDurationSec)
	transitionEnd := teleopStart + float64(TransitionDurationSec)
	endGameStart := teleopEnd - EndGameDurationSec

	// During auto and pause, both hubs are active
	assert.Equal(t, true, IsRedHubActive(0, true))               // Auto
	assert.Equal(t, true, IsRedHubActive(10, true))              // Auto
	assert.Equal(t, true, IsRedHubActive(20, true))              // Pause
	assert.Equal(t, true, IsRedHubActive(22, true))              // Pause
	assert.Equal(t, true, IsRedHubActive(teleopStart-0.1, true)) // End of pause

	// During transition period (first 10 seconds of teleop), both hubs are active
	assert.Equal(t, true, IsRedHubActive(teleopStart, true))        // Start of transition
	assert.Equal(t, true, IsRedHubActive(teleopStart+5, true))      // Middle of transition
	assert.Equal(t, true, IsRedHubActive(transitionEnd-0.1, true))  // End of transition
	assert.Equal(t, true, IsRedHubActive(teleopStart, false))       // Start of transition
	assert.Equal(t, true, IsRedHubActive(teleopStart+5, false))     // Middle of transition
	assert.Equal(t, true, IsRedHubActive(transitionEnd-0.1, false)) // End of transition

	// Red won auto: Red is INACTIVE first (shift 0), then alternates every 25 seconds
	// Shift 0 = 10-35 sec into teleop, Shift 1 = 35-60 sec, etc.
	assert.Equal(t, false, IsRedHubActive(transitionEnd, true))    // Start of shift 0 (10 sec into teleop) - INACTIVE
	assert.Equal(t, false, IsRedHubActive(transitionEnd+12, true)) // Middle of shift 0 - INACTIVE
	assert.Equal(t, true, IsRedHubActive(transitionEnd+25, true))  // Start of shift 1 (35 sec into teleop) - ACTIVE
	assert.Equal(t, false, IsRedHubActive(transitionEnd+50, true)) // Start of shift 2 (60 sec into teleop) - INACTIVE
	assert.Equal(t, true, IsRedHubActive(transitionEnd+75, true))  // Start of shift 3 (85 sec into teleop) - ACTIVE

	// Blue won auto or tie: Red is ACTIVE first (shift 0), then alternates every 25 seconds
	assert.Equal(t, true, IsRedHubActive(transitionEnd, false))     // Start of shift 0 - ACTIVE
	assert.Equal(t, true, IsRedHubActive(transitionEnd+12, false))  // Middle of shift 0 - ACTIVE
	assert.Equal(t, false, IsRedHubActive(transitionEnd+25, false)) // Start of shift 1 - INACTIVE
	assert.Equal(t, true, IsRedHubActive(transitionEnd+50, false))  // Start of shift 2 - ACTIVE
	assert.Equal(t, false, IsRedHubActive(transitionEnd+75, false)) // Start of shift 3 - INACTIVE

	// During END GAME (last 30 seconds), both hubs are active
	assert.Equal(t, true, IsRedHubActive(endGameStart, true))   // Start of END GAME
	assert.Equal(t, true, IsRedHubActive(endGameStart, false))  // Start of END GAME
	assert.Equal(t, true, IsRedHubActive(teleopEnd-15, true))   // Middle of END GAME
	assert.Equal(t, true, IsRedHubActive(teleopEnd-15, false))  // Middle of END GAME
	assert.Equal(t, true, IsRedHubActive(teleopEnd-0.1, true))  // End of match
	assert.Equal(t, true, IsRedHubActive(teleopEnd-0.1, false)) // End of match
}

func TestIsBlueHubActive(t *testing.T) {
	teleopStart := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEnd := teleopStart + float64(MatchTiming.TeleopDurationSec)
	transitionEnd := teleopStart + float64(TransitionDurationSec)
	endGameStart := teleopEnd - EndGameDurationSec

	// During auto and pause, both hubs are active
	assert.Equal(t, true, IsBlueHubActive(0, true))               // Auto
	assert.Equal(t, true, IsBlueHubActive(10, true))              // Auto
	assert.Equal(t, true, IsBlueHubActive(20, true))              // Pause
	assert.Equal(t, true, IsBlueHubActive(22, true))              // Pause
	assert.Equal(t, true, IsBlueHubActive(teleopStart-0.1, true)) // End of pause

	// During transition period (first 10 seconds of teleop), both hubs are active
	assert.Equal(t, true, IsBlueHubActive(teleopStart, true))        // Start of transition
	assert.Equal(t, true, IsBlueHubActive(teleopStart+5, true))      // Middle of transition
	assert.Equal(t, true, IsBlueHubActive(transitionEnd-0.1, true))  // End of transition
	assert.Equal(t, true, IsBlueHubActive(teleopStart, false))       // Start of transition
	assert.Equal(t, true, IsBlueHubActive(teleopStart+5, false))     // Middle of transition
	assert.Equal(t, true, IsBlueHubActive(transitionEnd-0.1, false)) // End of transition

	// Blue won auto: Blue is INACTIVE first (shift 0), then alternates every 25 seconds
	assert.Equal(t, false, IsBlueHubActive(transitionEnd, true))    // Start of shift 0 - INACTIVE
	assert.Equal(t, false, IsBlueHubActive(transitionEnd+12, true)) // Middle of shift 0 - INACTIVE
	assert.Equal(t, true, IsBlueHubActive(transitionEnd+25, true))  // Start of shift 1 - ACTIVE
	assert.Equal(t, false, IsBlueHubActive(transitionEnd+50, true)) // Start of shift 2 - INACTIVE
	assert.Equal(t, true, IsBlueHubActive(transitionEnd+75, true))  // Start of shift 3 - ACTIVE

	// Red won auto or tie: Blue is ACTIVE first (shift 0), then alternates every 25 seconds
	assert.Equal(t, true, IsBlueHubActive(transitionEnd, false))     // Start of shift 0 - ACTIVE
	assert.Equal(t, true, IsBlueHubActive(transitionEnd+12, false))  // Middle of shift 0 - ACTIVE
	assert.Equal(t, false, IsBlueHubActive(transitionEnd+25, false)) // Start of shift 1 - INACTIVE
	assert.Equal(t, true, IsBlueHubActive(transitionEnd+50, false))  // Start of shift 2 - ACTIVE
	assert.Equal(t, false, IsBlueHubActive(transitionEnd+75, false)) // Start of shift 3 - INACTIVE

	// During END GAME (last 30 seconds), both hubs are active
	assert.Equal(t, true, IsBlueHubActive(endGameStart, true))   // Start of END GAME
	assert.Equal(t, true, IsBlueHubActive(endGameStart, false))  // Start of END GAME
	assert.Equal(t, true, IsBlueHubActive(teleopEnd-15, true))   // Middle of END GAME
	assert.Equal(t, true, IsBlueHubActive(teleopEnd-15, false))  // Middle of END GAME
	assert.Equal(t, true, IsBlueHubActive(teleopEnd-0.1, true))  // End of match
	assert.Equal(t, true, IsBlueHubActive(teleopEnd-0.1, false)) // End of match
}

func TestIsRedHubActiveForScoring(t *testing.T) {
	teleopStart := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	transitionEnd := teleopStart + float64(TransitionDurationSec)
	teleopEnd := teleopStart + float64(MatchTiming.TeleopDurationSec)

	// Red won auto: Red is INACTIVE first (shift 0), then alternates
	// Transition period (both hubs active)
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopStart, true))
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopStart+9, true))

	// Grace period after transition ends (both hubs were active during transition)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd, true))      // 0 sec after transition (grace period)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd+2.9, true))  // 2.9 sec after transition (grace period)
	assert.Equal(t, false, IsRedHubActiveForScoring(transitionEnd+3.1, true)) // 3.1 sec after transition (shift 0, Red INACTIVE)

	// Shift 0 (10-35 sec into teleop, Red INACTIVE)
	assert.Equal(t, false, IsRedHubActiveForScoring(transitionEnd+12, true)) // Middle of shift 0

	// Shift 1 (35-60 sec into teleop, Red ACTIVE)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd+25, true)) // Start of shift 1 (Red ACTIVE)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd+40, true)) // Middle of shift 1

	// Shift 2 (60-85 sec into teleop, Red INACTIVE, but grace period from shift 1)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd+50, true))    // 0 sec into shift 2 (grace period)
	assert.Equal(t, true, IsRedHubActiveForScoring(transitionEnd+52.9, true))  // 2.9 sec into shift 2 (grace period)
	assert.Equal(t, false, IsRedHubActiveForScoring(transitionEnd+53.1, true)) // 3.1 sec into shift 2 (no longer grace period)

	// During END GAME, hub is active
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopEnd-30, true)) // Start of END GAME
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopEnd-1, true))  // End of match

	// Grace period applies even after match ends
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopEnd+0.5, true))  // 0.5 sec after match (grace period)
	assert.Equal(t, true, IsRedHubActiveForScoring(teleopEnd+2.9, true))  // 2.9 sec after match (grace period)
	assert.Equal(t, false, IsRedHubActiveForScoring(teleopEnd+3.1, true)) // 3.1 sec after match (no longer grace period)
}

func TestIsBlueHubActiveForScoring(t *testing.T) {
	teleopStart := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	transitionEnd := teleopStart + float64(TransitionDurationSec)
	teleopEnd := teleopStart + float64(MatchTiming.TeleopDurationSec)

	// Blue won auto: Blue is INACTIVE first (shift 0), then alternates
	// Transition period (both hubs active)
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopStart, true))
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopStart+9, true))

	// Grace period after transition ends (both hubs were active during transition)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd, true))      // 0 sec after transition (grace period)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd+2.9, true))  // 2.9 sec after transition (grace period)
	assert.Equal(t, false, IsBlueHubActiveForScoring(transitionEnd+3.1, true)) // 3.1 sec after transition (shift 0, Blue INACTIVE)

	// Shift 0 (10-35 sec into teleop, Blue INACTIVE)
	assert.Equal(t, false, IsBlueHubActiveForScoring(transitionEnd+12, true)) // Middle of shift 0

	// Shift 1 (35-60 sec into teleop, Blue ACTIVE)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd+25, true)) // Start of shift 1 (Blue ACTIVE)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd+40, true)) // Middle of shift 1

	// Shift 2 (60-85 sec into teleop, Blue INACTIVE, but grace period from shift 1)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd+50, true))    // 0 sec into shift 2 (grace period)
	assert.Equal(t, true, IsBlueHubActiveForScoring(transitionEnd+52.9, true))  // 2.9 sec into shift 2 (grace period)
	assert.Equal(t, false, IsBlueHubActiveForScoring(transitionEnd+53.1, true)) // 3.1 sec into shift 2 (no longer grace period)

	// During END GAME, hub is active
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopEnd-30, true)) // Start of END GAME
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopEnd-1, true))  // End of match

	// Grace period applies even after match ends
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopEnd+0.5, true))  // 0.5 sec after match (grace period)
	assert.Equal(t, true, IsBlueHubActiveForScoring(teleopEnd+2.9, true))  // 2.9 sec after match (grace period)
	assert.Equal(t, false, IsBlueHubActiveForScoring(teleopEnd+3.1, true)) // 3.1 sec after match (no longer grace period)
}
