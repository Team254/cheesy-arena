// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

var MatchTiming = struct {
	WarmupDurationSec           int
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TimeoutDurationSec          int
}{0, 15, 2, 135, 30, 0}
