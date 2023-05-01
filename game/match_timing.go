// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

var ChargeStationTeleopGracePeriod = 3 * time.Second

var MatchTiming = struct {
	WarmupDurationSec                  int
	AutoDurationSec                    int
	PauseDurationSec                   int
	TeleopDurationSec                  int
	WarningRemainingDurationSec        int
	TimeoutDurationSec                 int
	TimeoutWarningRemainingDurationSec int
}{0, 15, 3, 135, 30, 0, 60}

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec) *
		time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
}
