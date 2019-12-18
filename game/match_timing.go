// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

var MatchTiming = struct {
	WarmupDurationSec  int
	AutoDurationSec    int
	PauseDurationSec   int
	TeleopDurationSec  int
	TimeoutDurationSec int
}{0, 0, 0, 0, 0}
