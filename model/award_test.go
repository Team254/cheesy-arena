// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentAward(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	award, err := db.GetAwardById(1114)
	assert.Nil(t, err)
	assert.Nil(t, award)
}

func TestAwardCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	award := Award{0, JudgedAward, "Saftey Award", 254, ""}
	assert.Nil(t, db.CreateAward(&award))
	award2, err := db.GetAwardById(1)
	assert.Nil(t, err)
	assert.Equal(t, award, *award2)

	award2.Id = 0
	award2.AwardName = "Spirit Award"
	assert.Nil(t, db.CreateAward(award2))
	awards, err := db.GetAllAwards()
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(awards)) {
		assert.Equal(t, award, awards[0])
		assert.Equal(t, *award2, awards[1])
	}

	award.TeamId = 0
	award.PersonName = "Travus Cubington"
	assert.Nil(t, db.UpdateAward(&award))
	award2, err = db.GetAwardById(1)
	assert.Nil(t, err)
	assert.Equal(t, award.TeamId, award2.TeamId)
	assert.Equal(t, award.PersonName, award2.PersonName)

	assert.Nil(t, db.DeleteAward(award.Id))
	award2, err = db.GetAwardById(1)
	assert.Nil(t, err)
	assert.Nil(t, award2)
}

func TestTruncateAwards(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	award := Award{0, JudgedAward, "Saftey Award", 254, ""}
	db.CreateAward(&award)
	db.TruncateAwards()
	award2, err := db.GetAwardById(1)
	assert.Nil(t, err)
	assert.Nil(t, award2)
}

func TestGetAwardsByType(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	award1 := Award{0, WinnerAward, "Event Winner", 1114, ""}
	db.CreateAward(&award1)
	award2 := Award{0, FinalistAward, "Event Finalist", 2056, ""}
	db.CreateAward(&award2)
	award3 := Award{0, JudgedAward, "Saftey Award", 254, ""}
	db.CreateAward(&award3)
	award4 := Award{0, WinnerAward, "Event Winner", 254, ""}
	db.CreateAward(&award4)

	awards, err := db.GetAwardsByType(JudgedAward)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(awards)) {
		assert.Equal(t, award3, awards[0])
	}
	awards, err = db.GetAwardsByType(FinalistAward)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(awards)) {
		assert.Equal(t, award2, awards[0])
	}
	awards, err = db.GetAwardsByType(WinnerAward)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(awards)) {
		assert.Equal(t, award1, awards[0])
		assert.Equal(t, award4, awards[1])
	}
}
