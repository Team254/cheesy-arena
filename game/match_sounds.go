// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific audience sound timings.

package game

type MatchSound struct {
	Name          string
	FileExtension string
	MatchTimeSec  float64
	Timeout       bool
}

// List of sounds and how many seconds into the match they are played. A negative time indicates that the sound can only
// be triggered explicitly.
var MatchSounds []*MatchSound

func UpdateMatchSounds() {
	MatchSounds = []*MatchSound{
		{
			"start",
			"wav",
			0,
			false,
		},
		{
			"end",
			"wav",
			float64(MatchTiming.AutoDurationSec),
			false,
		},
		{
			"resume",
			"wav",
			float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec),
			false,
		},
		{
			"warning",
			"wav",
			float64(
				MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec + MatchTiming.TeleopDurationSec -
					MatchTiming.WarningRemainingDurationSec,
			),
			false,
		},
		{
			"end",
			"wav",
			float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec + MatchTiming.TeleopDurationSec),
			false,
		},
		{
			"timeout_warning",
			"wav",
			float64(MatchTiming.TimeoutDurationSec - MatchTiming.TimeoutWarningRemainingDurationSec),
			true,
		},
		{
			"end",
			"wav",
			float64(MatchTiming.TimeoutDurationSec),
			true,
		},
		{
			"abort",
			"wav",
			-1,
			false,
		},
		{
			"match_result",
			"wav",
			-1,
			false,
		},
	}
}
