// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

var MatchTiming = struct {
	WarmupDurationSec  int
	AutoDurationSec    int
	PauseDurationSec   int
	TeleopDurationSec  int
	EndgameTimeLeftSec int
	TimeoutDurationSec int
}{3, 15, 2, 135, 30, 0}

func GetAutoEndTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec) * time.Second)
}

func GetTeleopStartTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+
		MatchTiming.PauseDurationSec) * time.Second)
}

func GetMatchEndTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+
		MatchTiming.PauseDurationSec+MatchTiming.TeleopDurationSec) * time.Second)
}
