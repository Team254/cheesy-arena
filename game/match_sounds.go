// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific audience sound timings.

package game

type MatchSound struct {
	Name         string
	MatchTimeSec float64
}

// List of sounds and how many seconds into the match they are played.
var MatchSounds = []*MatchSound{
	{"match-start", 0},
	{"match-resume", 15},
	{"match-warning1", 120},
	{"match-warning2", 130},
	{"match-end", 150},
}
