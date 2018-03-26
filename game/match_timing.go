// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

var MatchTiming = struct {
	AutoDurationSec    int
	PauseDurationSec   int
	TeleopDurationSec  int
	EndgameTimeLeftSec int
}{15, 2, 135, 30}

func GetAutoEndTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.AutoDurationSec))
}

func GetTeleopStartTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec))
}

func GetMatchEndTime(matchStartTime time.Time) time.Time {
	return matchStartTime.Add(time.Duration(MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second)
}
