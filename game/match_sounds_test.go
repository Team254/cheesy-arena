// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUniqueMatchSounds(t *testing.T) {
	UpdateMatchSounds()

	uniqueSounds := UniqueMatchSounds()

	assert.Equal(
		t,
		[]string{
			"start",
			"end",
			"resume",
			"shift_change",
			"warning",
			"abort",
			"match_result",
			"pick_clock",
			"pick_clock_expired",
		},
		matchSoundNames(uniqueSounds),
	)
	assert.Len(t, uniqueSounds, 9)
	assert.Same(t, MatchSounds[0], uniqueSounds[0])
	assert.Same(t, MatchSounds[1], uniqueSounds[1])
	assert.Same(t, MatchSounds[3], uniqueSounds[3])
}

func matchSoundNames(matchSounds []*MatchSound) []string {
	names := make([]string, 0, len(matchSounds))
	for _, sound := range matchSounds {
		names = append(names, sound.Name)
	}
	return names
}
