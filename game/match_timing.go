// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

const (
	powerPortAutoGracePeriodSec   = 5
	PowerPortTeleopGracePeriodSec = 5
	rungAssessmentDelaySec        = 5
	RungAssessmentFlashPeriodMs   = 500 // ms
)

var MatchTiming = struct {
	WarmupDurationSec           int
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TimeoutDurationSec          int
}{0, 15, 2, 135, 30, 0}

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec) *
		time.Second
}

func GetDurationToWarning() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec-MatchTiming.WarningRemainingDurationSec) * time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
}

// Returns true if the given time is within the proper range for assessing the level state of the shield generator rung.
func ShouldAssessRung(matchStartTime, currentTime time.Time) bool {
	return currentTime.After(matchStartTime.Add(GetDurationToWarning())) &&
		currentTime.Before(matchStartTime.Add(GetDurationToTeleopEnd()+rungAssessmentDelaySec*time.Second))
}
