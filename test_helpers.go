// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package main

import (
	"github.com/Team254/cheesy-arena/field"
	"testing"
)

func setupTestWeb(t *testing.T) *Web {
	arena := field.SetupTestArena(t, "web")
	return NewWeb(arena)
}
