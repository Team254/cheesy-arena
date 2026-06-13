package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTeleopDurationSec(t *testing.T) {
	assert.Equal(t, 140, GetTeleopDurationSec())

	originalTransitionShiftDurationSec := MatchTiming.TransitionShiftDurationSec
	originalShiftDurationSec := MatchTiming.ShiftDurationSec
	originalEndgameDurationSec := MatchTiming.EndgameDurationSec
	defer func() {
		MatchTiming.TransitionShiftDurationSec = originalTransitionShiftDurationSec
		MatchTiming.ShiftDurationSec = originalShiftDurationSec
		MatchTiming.EndgameDurationSec = originalEndgameDurationSec
	}()

	MatchTiming.TransitionShiftDurationSec = 8
	MatchTiming.ShiftDurationSec = 20
	MatchTiming.EndgameDurationSec = 18

	assert.Equal(t, 106, GetTeleopDurationSec())
}
