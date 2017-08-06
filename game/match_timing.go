// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

var MatchTiming = struct {
	AutoDurationSec    int
	PauseDurationSec   int
	TeleopDurationSec  int
	EndgameTimeLeftSec int
}{15, 2, 135, 30}
