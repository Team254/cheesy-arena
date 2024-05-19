// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var matchStartTime = time.Unix(10, 0)

func TestAmpSpeaker_CalculationMethods(t *testing.T) {
	ampSpeaker := AmpSpeaker{
		AutoAmpNotes:                  1,
		TeleopAmpNotes:                2,
		AutoSpeakerNotes:              3,
		TeleopUnamplifiedSpeakerNotes: 5,
		TeleopAmplifiedSpeakerNotes:   8,
	}
	assert.Equal(t, 3, ampSpeaker.ampNotesScored())
	assert.Equal(t, 16, ampSpeaker.speakerNotesScored())
	assert.Equal(t, 19, ampSpeaker.TotalNotesScored())
	assert.Equal(t, 17, ampSpeaker.AutoNotePoints())
	assert.Equal(t, 4, ampSpeaker.AmpPoints())
	assert.Equal(t, 65, ampSpeaker.SpeakerPoints())
}

func TestAmpSpeaker_MatchSequence(t *testing.T) {
	var ampSpeaker AmpSpeaker
	assertAmpSpeaker := func(
		autoAmpNotes, teleopAmpNotes, autoSpeakerNotes, teleopUnamplifiedSpeakerNotes, teleopAmplifiedSpeakerNotes int,
	) {
		assert.Equal(t, autoAmpNotes, ampSpeaker.AutoAmpNotes)
		assert.Equal(t, teleopAmpNotes, ampSpeaker.TeleopAmpNotes)
		assert.Equal(t, autoSpeakerNotes, ampSpeaker.AutoSpeakerNotes)
		assert.Equal(t, teleopUnamplifiedSpeakerNotes, ampSpeaker.TeleopUnamplifiedSpeakerNotes)
		assert.Equal(t, teleopAmplifiedSpeakerNotes, ampSpeaker.TeleopAmplifiedSpeakerNotes)
	}

	ampSpeaker.UpdateState(0, 0, false, false, matchStartTime, timeAfterStart(0))
	assertAmpSpeaker(0, 0, 0, 0, 0)

	// Score in the Amp and Speaker during auto.
	ampSpeaker.UpdateState(1, 0, false, false, matchStartTime, timeAfterStart(1))
	assertAmpSpeaker(1, 0, 0, 0, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(1)))
	ampSpeaker.UpdateState(2, 0, false, false, matchStartTime, timeAfterStart(2))
	assertAmpSpeaker(2, 0, 0, 0, 0)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)
	ampSpeaker.UpdateState(2, 3, false, false, matchStartTime, timeAfterStart(3))
	assertAmpSpeaker(2, 0, 3, 0, 0)
	ampSpeaker.UpdateState(3, 4, false, false, matchStartTime, timeAfterStart(4))
	assertAmpSpeaker(3, 0, 4, 0, 0)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)

	// Pressing the buttons during auto should not have any effect.
	ampSpeaker.UpdateState(3, 4, true, true, matchStartTime, timeAfterStart(5))
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.CoopActivated)
	assert.Equal(t, 0.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(5)))

	// Score in the Amp and Speaker around the expiration of the grace period.
	ampSpeaker.UpdateState(4, 6, false, false, matchStartTime, timeAfterStart(17.9))
	assertAmpSpeaker(4, 0, 6, 0, 0)
	ampSpeaker.UpdateState(5, 8, false, false, matchStartTime, timeAfterStart(18.1))
	assertAmpSpeaker(4, 1, 6, 2, 0)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)

	// Activate co-op.
	ampSpeaker.UpdateState(5, 8, false, true, matchStartTime, timeAfterStart(20))
	assertAmpSpeaker(4, 1, 6, 2, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.CoopActivated)

	// Activate co-op a second time.
	ampSpeaker.UpdateState(5, 8, false, false, matchStartTime, timeAfterStart(21))
	assertAmpSpeaker(4, 1, 6, 2, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.CoopActivated)

	// Try to activate amplify with insufficient notes banked.
	ampSpeaker.UpdateState(5, 8, true, false, matchStartTime, timeAfterStart(22))
	assertAmpSpeaker(4, 1, 6, 2, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(22), false))

	// Score more notes in the Amp and amplify.
	ampSpeaker.UpdateState(7, 8, false, false, matchStartTime, timeAfterStart(23))
	assertAmpSpeaker(4, 3, 6, 2, 0)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)
	ampSpeaker.UpdateState(7, 8, true, false, matchStartTime, timeAfterStart(24))
	assertAmpSpeaker(4, 3, 6, 2, 0)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(24.1), false))
	assert.Equal(t, 9.9, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(24.1)))

	// Score in the amplified Speaker and the Amp.
	ampSpeaker.UpdateState(8, 11, false, false, matchStartTime, timeAfterStart(25))
	assertAmpSpeaker(4, 4, 6, 2, 3)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(26), false))
	assert.Equal(t, 8.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(26)))

	// Exceed the note limit for the amplified Speaker.
	ampSpeaker.UpdateState(8, 15, false, false, matchStartTime, timeAfterStart(27))
	assertAmpSpeaker(4, 4, 6, 5, 4)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(27), false))
	assert.Equal(t, 0.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(27)))

	// Do another amplified cycle and test the grace period.
	ampSpeaker.UpdateState(10, 15, true, false, matchStartTime, timeAfterStart(30))
	assertAmpSpeaker(4, 6, 6, 5, 4)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(31), false))
	assert.Equal(t, 9.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(31)))
	ampSpeaker.UpdateState(10, 16, true, false, matchStartTime, timeAfterStart(32))
	assertAmpSpeaker(4, 6, 6, 5, 5)
	ampSpeaker.UpdateState(10, 17, true, false, matchStartTime, timeAfterStart(42.9))
	assertAmpSpeaker(4, 6, 6, 5, 6)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(42.9), true))
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(42.9), false))
	assert.Equal(t, 0.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(42.9)))
	ampSpeaker.UpdateState(10, 18, true, false, matchStartTime, timeAfterStart(43.1))
	assertAmpSpeaker(4, 6, 6, 6, 6)
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(43.1), true))

	// Test around the end of the match and the grace period after.
	ampSpeaker.UpdateState(11, 21, false, false, matchStartTime, timeAfterStart(152.9))
	assertAmpSpeaker(4, 7, 6, 9, 6)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	ampSpeaker.UpdateState(13, 23, true, false, matchStartTime, timeAfterStart(153.1))
	assertAmpSpeaker(4, 7, 6, 11, 6)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	ampSpeaker.UpdateState(13, 24, true, false, matchStartTime, timeAfterStart(157.9))
	assertAmpSpeaker(4, 7, 6, 12, 6)
	ampSpeaker.UpdateState(13, 25, false, false, matchStartTime, timeAfterStart(158.1))
	assertAmpSpeaker(4, 7, 6, 12, 6)

	// Reset the AmpSpeaker to test different conditions and settings.
	ampSpeaker = AmpSpeaker{}
	assertAmpSpeaker(0, 0, 0, 0, 0)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)

	// Attempt to co-op with insufficient notes banked.
	ampSpeaker.UpdateState(0, 0, false, true, matchStartTime, timeAfterStart(20))
	assert.Equal(t, false, ampSpeaker.CoopActivated)

	// Attempt to co-op just before the window has closed.
	assert.Equal(t, true, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(62.9)))
	ampSpeaker.UpdateState(1, 0, false, true, matchStartTime, timeAfterStart(62.9))
	assertAmpSpeaker(0, 1, 0, 0, 0)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.CoopActivated)
	assert.Equal(t, true, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(62.9)))

	// Undo the co-op and try again after the window has closed.
	ampSpeaker = AmpSpeaker{}
	assertAmpSpeaker(0, 0, 0, 0, 0)
	assert.Equal(t, false, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(63.1)))
	ampSpeaker.UpdateState(1, 0, false, true, matchStartTime, timeAfterStart(63.1))
	assertAmpSpeaker(0, 1, 0, 0, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.CoopActivated)

	// Backtrack and disable co-op.
	assertAmpSpeaker(0, 1, 0, 0, 0)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(60)))
	MelodyBonusThresholdWithCoop = 0
	assert.Equal(t, false, ampSpeaker.IsCoopWindowOpen(matchStartTime, timeAfterStart(60)))
	ampSpeaker.UpdateState(2, 0, false, true, matchStartTime, timeAfterStart(60))
	assertAmpSpeaker(0, 2, 0, 0, 0)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.CoopActivated)

	// Test with different amplification note limit and duration.
	AmplificationNoteLimit = 3
	AmplificationDurationSec = 6
	ampSpeaker.UpdateState(2, 0, true, false, matchStartTime, timeAfterStart(70))
	assertAmpSpeaker(0, 2, 0, 0, 0)
	ampSpeaker.UpdateState(2, 1, true, false, matchStartTime, timeAfterStart(71))
	assertAmpSpeaker(0, 2, 0, 0, 1)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(71), false))
	assert.Equal(t, 5.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(71)))
	ampSpeaker.UpdateState(2, 4, false, false, matchStartTime, timeAfterStart(72))
	assertAmpSpeaker(0, 2, 0, 1, 3)
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(72), true))
	assert.Equal(t, 0.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(72)))

	// Test with no amplification note limit and long duration.
	AmplificationNoteLimit = 0
	AmplificationDurationSec = 23
	ampSpeaker.LastAmplifiedTime = time.Time{}
	ampSpeaker.UpdateState(4, 4, true, false, matchStartTime, timeAfterStart(73))
	assertAmpSpeaker(0, 4, 0, 1, 3)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(74), true))
	assert.Equal(t, 22.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(74)))
	ampSpeaker.UpdateState(100, 44, false, false, matchStartTime, timeAfterStart(94))
	assertAmpSpeaker(0, 100, 0, 1, 43)
	assert.Equal(t, 0, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(94), true))
	assert.Equal(t, 2.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(94)))
	ampSpeaker.UpdateState(101, 57, false, false, matchStartTime, timeAfterStart(98.9))
	assertAmpSpeaker(0, 101, 0, 1, 56)
	assert.Equal(t, 1, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(98.9), true))
	assert.Less(t, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(98.9)), 0.2)
	ampSpeaker.UpdateState(102, 60, false, false, matchStartTime, timeAfterStart(99.1))
	assertAmpSpeaker(0, 102, 0, 4, 56)
	assert.Equal(t, 2, ampSpeaker.BankedAmpNotes)
	assert.Equal(t, false, ampSpeaker.isAmplified(timeAfterStart(99.1), true))
	assert.Equal(t, 0.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(99.1)))

	// Restore default settings.
	AmplificationNoteLimit = 4
	AmplificationDurationSec = 10

	// Test hitting the amplification button just before the end of the match.
	ampSpeaker.UpdateState(102, 60, true, false, matchStartTime, timeAfterStart(152))
	ampSpeaker.UpdateState(102, 63, true, false, matchStartTime, timeAfterStart(157))
	assertAmpSpeaker(0, 102, 0, 4, 59)
	assert.Equal(t, true, ampSpeaker.isAmplified(timeAfterStart(157), true))
	assert.Equal(t, 5.0, ampSpeaker.AmplifiedTimeRemaining(timeAfterStart(157)))
	ampSpeaker.UpdateState(102, 66, true, false, matchStartTime, timeAfterStart(157.9))
	assertAmpSpeaker(0, 102, 0, 6, 60)
}

func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}
