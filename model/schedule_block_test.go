// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScheduleBlockCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	scheduleBlock1 := ScheduleBlock{0, Practice, time.Now().UTC(), 10, 600}
	assert.Nil(t, db.CreateScheduleBlock(&scheduleBlock1))
	scheduleBlock2 := ScheduleBlock{0, Qualification, time.Now().UTC(), 20, 480}
	assert.Nil(t, db.CreateScheduleBlock(&scheduleBlock2))
	scheduleBlock3 := ScheduleBlock{0, Qualification, scheduleBlock2.StartTime.Add(time.Second * 20 * 480), 20, 480}
	assert.Nil(t, db.CreateScheduleBlock(&scheduleBlock3))

	// Test retrieval of all blocks by match type.
	blocks, err := db.GetScheduleBlocksByMatchType(Practice)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(blocks)) {
		assert.Equal(t, scheduleBlock1, blocks[0])
	}
	blocks, err = db.GetScheduleBlocksByMatchType(Qualification)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(blocks)) {
		assert.Equal(t, scheduleBlock2, blocks[0])
		assert.Equal(t, scheduleBlock3, blocks[1])
	}

	// Test deletion of blocks.
	assert.Nil(t, db.DeleteScheduleBlocksByMatchType(Practice))
	blocks, err = db.GetScheduleBlocksByMatchType(Practice)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(blocks))
	blocks, err = db.GetScheduleBlocksByMatchType(Qualification)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(blocks))
	assert.Nil(t, db.TruncateScheduleBlocks())
	blocks, err = db.GetScheduleBlocksByMatchType(Qualification)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(blocks))
}
