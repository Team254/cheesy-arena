// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentLowerThird(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	lowerThird, err := db.GetLowerThirdById(1114)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird)
}

func TestLowerThirdCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 0}
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
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	lowerThird := LowerThird{0, "Top Text", "Bottom Text", 0}
	db.CreateLowerThird(&lowerThird)
	db.TruncateLowerThirds()
	lowerThird2, err := db.GetLowerThirdById(1)
	assert.Nil(t, err)
	assert.Nil(t, lowerThird2)
}
