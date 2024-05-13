// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2024 Amp and Speaker elements.

package game

import (
	"time"
)

const bankedAmpNoteLimit = 2

type AmpSpeaker struct {
	BankedAmpNotes                int
	CoopActivated                 bool
	autoAmpNotes                  int
	teleopAmpNotes                int
	autoSpeakerNotes              int
	teleopUnamplifiedSpeakerNotes int
	teleopAmplifiedSpeakerNotes   int
	lastAmplifiedTime             time.Time
	lastAmplifiedSpeakerNotes     int
}

// Updates the internal state of the AmpSpeaker based on the PLC inputs.
func (ampSpeaker *AmpSpeaker) UpdateState(
	ampNoteCount, speakerNoteCount int, amplifyButton, coopButton bool, matchStartTime, currentTime time.Time,
) {
	newAmpNotes := ampNoteCount - ampSpeaker.ampNotesScored()
	newSpeakerNotes := speakerNoteCount - ampSpeaker.speakerNotesScored()

	// Handle the autonomous period.
	autoValidityCutoff := matchStartTime.Add(GetDurationToAutoEnd() + speakerAutoGracePeriodSec*time.Second)
	if currentTime.Before(autoValidityCutoff) {
		ampSpeaker.autoAmpNotes += newAmpNotes
		ampSpeaker.BankedAmpNotes = min(ampSpeaker.BankedAmpNotes+newAmpNotes, bankedAmpNoteLimit)
		ampSpeaker.autoSpeakerNotes += newSpeakerNotes

		// Bail out to avoid exercising the teleop logic.
		return
	}

	// Handle the Amp.
	teleopAmpValidityCutoff := matchStartTime.Add(GetDurationToTeleopEnd())
	if currentTime.Before(teleopAmpValidityCutoff) {
		// Handle incoming Amp notes.
		ampSpeaker.teleopAmpNotes += newAmpNotes
		if !ampSpeaker.isAmplified(currentTime, false) {
			ampSpeaker.BankedAmpNotes = min(ampSpeaker.BankedAmpNotes+newAmpNotes, bankedAmpNoteLimit)
		}

		// Handle the co-op button.
		if coopButton && !ampSpeaker.CoopActivated && ampSpeaker.BankedAmpNotes >= 1 &&
			ampSpeaker.IsCoopWindowOpen(matchStartTime, currentTime) {
			ampSpeaker.CoopActivated = true
			ampSpeaker.BankedAmpNotes--
		}

		// Handle the amplify button.
		if amplifyButton && !ampSpeaker.isAmplified(currentTime, false) && ampSpeaker.BankedAmpNotes >= 2 {
			ampSpeaker.lastAmplifiedTime = currentTime
			ampSpeaker.lastAmplifiedSpeakerNotes = 0
			ampSpeaker.BankedAmpNotes -= 2
		}
	}

	// Handle the Speaker.
	teleopSpeakerValidityCutoff := matchStartTime.Add(
		GetDurationToTeleopEnd() + speakerTeleopGracePeriodSec*time.Second,
	)
	if currentTime.Before(teleopSpeakerValidityCutoff) {
		for newSpeakerNotes > 0 && ampSpeaker.isAmplified(currentTime, true) {
			ampSpeaker.teleopAmplifiedSpeakerNotes++
			ampSpeaker.lastAmplifiedSpeakerNotes++
			newSpeakerNotes--
		}
		ampSpeaker.teleopUnamplifiedSpeakerNotes += newSpeakerNotes
	}
}

// Returns the amount of time remaining in the current amplification period, or zero if not currently amplified.
func (ampSpeaker *AmpSpeaker) AmplifiedTimeRemaining(currentTime time.Time) float64 {
	if !ampSpeaker.isAmplified(currentTime, false) {
		return 0
	}
	return float64(AmplificationDurationSec) - currentTime.Sub(ampSpeaker.lastAmplifiedTime).Seconds()
}

// Returns true if the co-op window during the match is currently open.
func (ampSpeaker *AmpSpeaker) IsCoopWindowOpen(matchStartTime, currentTime time.Time) bool {
	coopValidityCutoff := matchStartTime.Add(GetDurationToTeleopStart() + coopTeleopWindowSec*time.Second)
	return MelodyBonusWithCoop > 0 && currentTime.Before(coopValidityCutoff)
}

// Returns the total number of notes scored in the Amp and Speaker.
func (ampSpeaker *AmpSpeaker) TotalNotesScored() int {
	return ampSpeaker.ampNotesScored() + ampSpeaker.speakerNotesScored()
}

// Returns the total points scored in the Amp and Speaker during the autonomous period.
func (ampSpeaker *AmpSpeaker) AutoNotePoints() int {
	return 2*ampSpeaker.autoAmpNotes + 5*ampSpeaker.autoSpeakerNotes
}

// Returns the total points scored in the Amp and Speaker during the teleoperated period.
func (ampSpeaker *AmpSpeaker) TeleopNotePoints() int {
	return ampSpeaker.teleopAmpNotes +
		2*ampSpeaker.teleopUnamplifiedSpeakerNotes +
		5*ampSpeaker.teleopAmplifiedSpeakerNotes
}

// Returns the total points scored in the Amp.
func (ampSpeaker *AmpSpeaker) AmpPoints() int {
	return 2*ampSpeaker.autoAmpNotes + ampSpeaker.teleopAmpNotes
}

// Returns the total points scored in the Speaker.
func (ampSpeaker *AmpSpeaker) SpeakerPoints() int {
	return 5*ampSpeaker.autoSpeakerNotes +
		2*ampSpeaker.teleopUnamplifiedSpeakerNotes +
		5*ampSpeaker.teleopAmplifiedSpeakerNotes
}

// Returns the total number of notes scored in the Amp.
func (ampSpeaker *AmpSpeaker) ampNotesScored() int {
	return ampSpeaker.autoAmpNotes + ampSpeaker.teleopAmpNotes
}

// Returns the total number of notes scored in the Speaker.
func (ampSpeaker *AmpSpeaker) speakerNotesScored() int {
	return ampSpeaker.autoSpeakerNotes +
		ampSpeaker.teleopUnamplifiedSpeakerNotes +
		ampSpeaker.teleopAmplifiedSpeakerNotes
}

// Returns whether the Speaker should be counting new incoming notes as amplified.
func (ampSpeaker *AmpSpeaker) isAmplified(currentTime time.Time, includeGracePeriod bool) bool {
	amplifiedValidityCutoff := ampSpeaker.lastAmplifiedTime.Add(time.Duration(AmplificationDurationSec) * time.Second)
	if includeGracePeriod {
		amplifiedValidityCutoff = amplifiedValidityCutoff.Add(
			time.Duration(speakerAmplifiedGracePeriodSec) * time.Second,
		)
	}
	meetsTimeCriterion := currentTime.After(ampSpeaker.lastAmplifiedTime) && currentTime.Before(amplifiedValidityCutoff)
	meetsNoteCriterion := AmplificationNoteLimit == 0 || ampSpeaker.lastAmplifiedSpeakerNotes < AmplificationNoteLimit
	return meetsTimeCriterion && meetsNoteCriterion
}
