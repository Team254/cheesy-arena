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
var MatchSounds = []*MatchSound{
	{"start", "wav", 0},
	{"resume", "wav", 15},
	{"warning1", "wav", 120},
	{"warning2", "wav", 130},
	{"end", "wav", 150},
	{"abort", "mp3", -1},
}
