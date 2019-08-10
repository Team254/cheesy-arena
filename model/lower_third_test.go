// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentLowerThird(t *testing.T) {
	db := setupTestDb(t)

	lowerThird, err := db.GetLowerThirdById(1114)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird)
}

func TestLowerThirdCrud(t *testing.T) {
	db := setupTestDb(t)

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 0, 0}
	db.CreateLowerThird(&lowerThird)
	lowerThird2, err := db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Equal(t, lowerThird, *lowerThird2)

	lowerThird.BottomText = "Blorpy"
	db.SaveLowerThird(&lowerThird)
	lowerThird2, err = db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Equal(t, lowerThird.BottomText, lowerThird2.BottomText)

	db.DeleteLowerThird(&lowerThird)
	lowerThird2, err = db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird2)
}

func TestTruncateLowerThirds(t *testing.T) {
	db := setupTestDb(t)

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 0, 0}
	db.CreateLowerThird(&lowerThird)
	db.TruncateLowerThirds()
	lowerThird2, err := db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird2)
}

func TestGetLowerThirdsByAwardId(t *testing.T) {
	db := setupTestDb(t)
	lowerThird1 := LowerThird{0, "Top Text", "Bottom Text", 0, 0}
	db.CreateLowerThird(&lowerThird1)
	lowerThird2 := LowerThird{0, "Award 1", "", 1, 5}
	db.CreateLowerThird(&lowerThird2)
	lowerThird3 := LowerThird{0, "Award 2", "", 2, 2}
	db.CreateLowerThird(&lowerThird3)
	lowerThird4 := LowerThird{0, "Award 1", "Award 1 Winner", 3, 5}
	db.CreateLowerThird(&lowerThird4)
	nextDisplayOrder := db.GetNextLowerThirdDisplayOrder()
	assert.Equal(t, 4, nextDisplayOrder)

	lowerThirds, err := db.GetLowerThirdsByAwardId(5)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(lowerThirds)) {
		assert.Equal(t, lowerThird2, lowerThirds[0])
		assert.Equal(t, lowerThird4, lowerThirds[1])
	}
	lowerThirds, err = db.GetLowerThirdsByAwardId(2)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(lowerThirds)) {
		assert.Equal(t, lowerThird3, lowerThirds[0])
	}
}
