// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScheduledBreakCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	scheduledBreak1 := ScheduledBreak{0, Qualification, 50, time.Unix(100, 0).UTC(), 600, "Lunch"}
	assert.Nil(t, db.CreateScheduledBreak(&scheduledBreak1))
	scheduledBreak2 := ScheduledBreak{0, Qualification, 25, time.Unix(200, 0).UTC(), 300, "Breakfast"}
	assert.Nil(t, db.CreateScheduledBreak(&scheduledBreak2))
	scheduledBreak3 := ScheduledBreak{0, Playoff, 4, time.Unix(500, 0).UTC(), 900, "Awards"}
	assert.Nil(t, db.CreateScheduledBreak(&scheduledBreak3))

	// Test retrieval of all blocks by match type.
	scheduledBreaks, err := db.GetScheduledBreaksByMatchType(Practice)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(scheduledBreaks))
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Qualification)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(scheduledBreaks)) {
		assert.Equal(t, scheduledBreak2, scheduledBreaks[0])
		assert.Equal(t, scheduledBreak1, scheduledBreaks[1])
	}
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Playoff)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(scheduledBreaks)) {
		assert.Equal(t, scheduledBreak3, scheduledBreaks[0])
	}

	// Test individual retrieval by match type and order.
	scheduledBreak, err := db.GetScheduledBreakByMatchTypeOrder(Qualification, 25)
	assert.Nil(t, err)
	assert.Equal(t, scheduledBreak2, *scheduledBreak)
	scheduledBreak, err = db.GetScheduledBreakByMatchTypeOrder(Playoff, 4)
	assert.Nil(t, err)
	assert.Equal(t, scheduledBreak3, *scheduledBreak)
	scheduledBreak, err = db.GetScheduledBreakByMatchTypeOrder(Qualification, 100)
	assert.Nil(t, err)
	assert.Nil(t, scheduledBreak)

	// Test deletion of breaks.
	assert.Nil(t, db.DeleteScheduledBreaksByMatchType(Playoff))
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(scheduledBreaks))
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Qualification)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(scheduledBreaks))

	assert.Nil(t, db.TruncateScheduledBreaks())
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Qualification)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(scheduledBreaks))
	scheduledBreaks, err = db.GetScheduledBreaksByMatchType(Playoff)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(scheduledBreaks))
}
