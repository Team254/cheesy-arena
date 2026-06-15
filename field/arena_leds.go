// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific control of the 2026 Hub DMX lighting.

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/led"
	"log"
	"time"
)

const (
	hubLightWarningSec = 3
	hubLightScoringAssessmentSec
)

// SetLedMode updates the Hub LED mode and notifies listeners if the published mode changed.
func (arena *Arena) SetLedMode(redMode, blueMode led.Mode) {
	arena.Leds.SetMode(redMode, blueMode)
	currentRed, currentBlue := arena.Leds.GetModes()
	if currentRed != arena.lastRedLedMode || currentBlue != arena.lastBlueLedMode {
		arena.LedChangeNotifier.Notify()
	}
}

// updateHubLeds updates Hub LEDs based on the current match state and active scoring shift.
func (arena *Arena) updateHubLeds(currentTime time.Time) {
	switch arena.MatchState {
	case AutoPeriod:
		arena.SetLedMode(led.RedStartupMode, led.BlueStartupMode)
	case PausePeriod:
		arena.SetLedMode(led.RedMode, led.BlueMode)
	case TeleopPeriod:
		arena.updateTeleopHubLeds(currentTime)
	case PostMatch:
		if arena.lastMatchState != PostMatch {
			// Set the Hub LEDs to white at the end of the match, and then turn them off when the referees are supposed
			// to assess tower climbs.
			arena.SetLedMode(led.WhiteMode, led.WhiteMode)
			go func() {
				time.Sleep(hubLightScoringAssessmentSec * time.Second)
				arena.SetLedMode(led.OffMode, led.OffMode)
			}()
		}
	}

	if err := arena.Leds.Update(); err != nil {
		log.Printf("Failed to update hub LEDs: %s", err)
	}
}

// updateTeleopHubLeds updates teleop LED modes using the active Hub shift, auto winner, and warning window.
func (arena *Arena) updateTeleopHubLeds(currentTime time.Time) {
	shift, remaining, _, ok := arena.RedRealtimeScore.CurrentScore.Hub.GetCurrentShiftTiming(
		arena.MatchStartTime, currentTime,
	)
	if !ok {
		return
	}

	redRemaining, _ := arena.RedRealtimeScore.CurrentScore.Hub.GetActiveShiftTiming(
		arena.MatchStartTime, currentTime,
	)
	blueRemaining, _ := arena.BlueRealtimeScore.CurrentScore.Hub.GetActiveShiftTiming(
		arena.MatchStartTime, currentTime,
	)

	redMode := led.OffMode
	if redRemaining > 0 {
		redMode = led.RedMode
	}
	blueMode := led.OffMode
	if blueRemaining > 0 {
		blueMode = led.BlueMode
	}

	// Pulse the LEDs when the Hub is about to go inactive.
	if remaining <= time.Duration(hubLightWarningSec)*time.Second {
		switch shift {
		case game.ShiftTransition:
			if arena.redWonAuto {
				redMode = led.RedPulseMode
			} else {
				blueMode = led.BluePulseMode
			}
		case game.Shift1, game.Shift3:
			if arena.redWonAuto {
				blueMode = led.BluePulseMode
			} else {
				redMode = led.RedPulseMode
			}
		case game.Shift2:
			if arena.redWonAuto {
				redMode = led.RedPulseMode
			} else {
				blueMode = led.BluePulseMode
			}
		case game.ShiftEndgame:
			redMode = led.RedPulseMode
			blueMode = led.BluePulseMode
		default:
		}
	} else if shift == game.ShiftTransition {
		if arena.redWonAuto {
			redMode = led.RedAdvantageMode
		} else {
			blueMode = led.BlueAdvantageMode
		}
	}
	arena.SetLedMode(redMode, blueMode)
}
