// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJudgingSlotCrud(t *testing.T) {
	database := setupTestDb(t)

	// Test creation of a judging slot with all fields populated.
	visitTime := time.Unix(100, 0).UTC()
	prevMatchTime := time.Unix(50, 0).UTC()
	nextMatchTime := time.Unix(150, 0).UTC()
	judgingSlot := JudgingSlot{
		Time:                visitTime,
		TeamId:              1503,
		PreviousMatchNumber: 5,
		PreviousMatchTime:   prevMatchTime,
		NextMatchNumber:     6,
		NextMatchTime:       nextMatchTime,
		JudgeNumber:         2,
	}
	assert.Nil(t, database.CreateJudgingSlot(&judgingSlot))
	assert.NotEqual(t, 0, judgingSlot.Id)

	// Test retrieving all judging slots and verify all fields.
	slots, err := database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(slots))
	assert.Equal(t, visitTime, slots[0].Time)
	assert.Equal(t, 1503, slots[0].TeamId)
	assert.Equal(t, 5, slots[0].PreviousMatchNumber)
	assert.Equal(t, prevMatchTime, slots[0].PreviousMatchTime)
	assert.Equal(t, 6, slots[0].NextMatchNumber)
	assert.Equal(t, nextMatchTime, slots[0].NextMatchTime)
	assert.Equal(t, 2, slots[0].JudgeNumber)

	// Test creating additional judging slots.
	slot1 := JudgingSlot{Time: time.Unix(300, 0), TeamId: 1678, JudgeNumber: 1}
	slot2 := JudgingSlot{Time: time.Unix(400, 0), TeamId: 1114, JudgeNumber: 2}
	assert.Nil(t, database.CreateJudgingSlot(&slot1))
	assert.Nil(t, database.CreateJudgingSlot(&slot2))
	slots, err = database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(slots))
	assert.Equal(t, 1114, slots[0].TeamId)
	assert.Equal(t, 1503, slots[1].TeamId)
	assert.Equal(t, 1678, slots[2].TeamId)

	// Test truncating all judging slots.
	assert.Nil(t, database.TruncateJudgingSlots())
	slots, err = database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(slots))
}
