// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

const (
	ScoringGracePeriodSec  = 3
	MotorsOnExtraPeriodSec = 2
)

var MatchTiming = struct {
	AutoDurationSec            int
	PauseDurationSec           int
	TransitionShiftDurationSec int
	ShiftDurationSec           int
	EndgameDurationSec         int
	TimeoutDurationSec         int
}{20, 3, 10, 25, 30, 0}

func GetTeleopDurationSec() int {
	return MatchTiming.TransitionShiftDurationSec + 4*MatchTiming.ShiftDurationSec + MatchTiming.EndgameDurationSec
}

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(
		MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec,
	) * time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(
		MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+GetTeleopDurationSec(),
	) * time.Second
}
