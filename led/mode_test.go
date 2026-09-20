// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package led

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModeNames(t *testing.T) {
	assert.Equal(t, "Test: Facing Scoring Table", RedModeNames[Side2TestMode])
	assert.Equal(t, "Test: Facing Audience", RedModeNames[Side4TestMode])

	assert.Equal(t, "Test: Facing Audience", BlueModeNames[Side2TestMode])
	assert.Equal(t, "Test: Facing Scoring Table", BlueModeNames[Side4TestMode])

	// Verify common modes are present in both
	assert.Equal(t, "Red", RedModeNames[RedMode])
	assert.Equal(t, "Red", BlueModeNames[RedMode])
	assert.Equal(t, "Test: Facing Driver Station", RedModeNames[Side1TestMode])
	assert.Equal(t, "Test: Facing Center", BlueModeNames[Side3TestMode])
}
