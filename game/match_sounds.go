// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific audience sound timings.

package game

type MatchSound struct {
	Name          string
	FileExtension string
	MatchTimeSec  float64
}

// List of sounds and how many seconds into the match they are played. A negative time indicates that the sound can only
// be triggered explicitly.
var MatchSounds []*MatchSound

func UpdateMatchSounds() {
	MatchSounds = []*MatchSound{
		{"start", "wav", 0},
		{"end", "wav", float64(MatchTiming.AutoDurationSec)},
		{"resume", "wav", float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)},
		{"warning", "wav", float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec +
			MatchTiming.TeleopDurationSec - MatchTiming.WarningRemainingDurationSec)},
		{"end", "wav", float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec +
			MatchTiming.TeleopDurationSec)},
		{"abort", "wav", -1},
		{"rotation", "wav", -1},
		{"position", "wav", -1},
	}
}
