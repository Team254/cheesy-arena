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
	teleopStart := MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec
	MatchSounds = []*MatchSound{
		{
			"start",
			"wav",
			0,
		},
		{
			"end",
			"wav",
			float64(MatchTiming.AutoDurationSec),
		},
		{
			"resume",
			"wav",
			float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec),
		},
		// --- 新增：Hub 轉換音效 (change.wav) ---
		// 根據時間表：2:20 -> 2:10 (過 10s 第一次轉換)
		{"change", "wav", float64(teleopStart + 10)},

		// SHIFT 1 -> SHIFT 2 (過 35s)
		{"change", "wav", float64(teleopStart + 35)},

		// SHIFT 2 -> SHIFT 3 (過 60s)
		{"change", "wav", float64(teleopStart + 59)},

		// SHIFT 3 -> SHIFT 4 (過 85s)
		{"change", "wav", float64(teleopStart + 84)},

		// SHIFT 4 -> END GAME (過 110s)
		{"change", "wav", float64(teleopStart + 109)},
		{
			"warning",
			"wav",
			float64(
				MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec + MatchTiming.TeleopDurationSec -
					MatchTiming.WarningRemainingDurationSec,
			),
		},
		{
			"end",
			"wav",
			float64(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec + MatchTiming.TeleopDurationSec),
		},
		{
			"abort",
			"wav",
			-1,
		},
		{
			"match_result",
			"wav",
			-1,
		},
		{
			"pick_clock",
			"wav",
			-1,
		},
		{
			"pick_clock_expired",
			"wav",
			-1,
		},
	}
}
