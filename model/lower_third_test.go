// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentLowerThird(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	lowerThird, err := db.GetLowerThirdById(1114)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird)
}

func TestLowerThirdCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	lowerThirds, err := db.GetAllLowerThirds()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(lowerThirds))

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 1, 0}
	assert.Nil(t, db.CreateLowerThird(&lowerThird))
	lowerThird2, err := db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Equal(t, lowerThird, *lowerThird2)

	lowerThird.BottomText = "Blorpy"
	assert.Nil(t, db.UpdateLowerThird(&lowerThird))
	lowerThird2, err = db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Equal(t, lowerThird.BottomText, lowerThird2.BottomText)

	lowerThirds, err = db.GetAllLowerThirds()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(lowerThirds))

	assert.Nil(t, db.DeleteLowerThird(lowerThird.Id))
	lowerThird2, err = db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird2)
}

func TestTruncateLowerThirds(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 0, 0}
	assert.Nil(t, db.CreateLowerThird(&lowerThird))
	assert.Nil(t, db.TruncateLowerThirds())
	lowerThird2, err := db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird2)
}

func TestGetLowerThirdsByAwardId(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	nextDisplayOrder := db.GetNextLowerThirdDisplayOrder()
	assert.Equal(t, 1, nextDisplayOrder)
	lowerThird1 := LowerThird{0, "Top Text", "Bottom Text", 1, 0}
	assert.Nil(t, db.CreateLowerThird(&lowerThird1))
	lowerThird2 := LowerThird{0, "Award 1", "", 2, 5}
	assert.Nil(t, db.CreateLowerThird(&lowerThird2))
	lowerThird3 := LowerThird{0, "Award 2", "", 3, 2}
	assert.Nil(t, db.CreateLowerThird(&lowerThird3))
	lowerThird4 := LowerThird{0, "Award 1", "Award 1 Winner", 4, 5}
	assert.Nil(t, db.CreateLowerThird(&lowerThird4))
	lowerThirds, err := db.GetAllLowerThirds()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(lowerThirds))
	nextDisplayOrder = db.GetNextLowerThirdDisplayOrder()
	assert.Equal(t, 5, nextDisplayOrder)

	lowerThirds, err = db.GetLowerThirdsByAwardId(5)
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
	lowerThirds, err = db.GetLowerThirdsByAwardId(39)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(lowerThirds))
}
