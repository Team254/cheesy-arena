// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"path/filepath"
	"testing"
)

func SetupTestArena(t *testing.T) *Arena {
	rand.Seed(0)
	model.BaseDir = ".."
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "test.db")
	arena, err := NewArena(dbPath)
	assert.Nil(t, err)
	t.Cleanup(
		func() {
			arena.Database.Close()
		},
	)
	return arena
}

func setupTestArena(t *testing.T) *Arena {
	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.PauseDurationSec = 2
	return SetupTestArena(t)
}
