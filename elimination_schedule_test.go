// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEliminationScheduleInitial(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 2)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err := db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 3, len(matches)) {
		assertMatch(t, matches[0], "F-1", 1, 2)
		assertMatch(t, matches[1], "F-2", 1, 2)
		assertMatch(t, matches[2], "F-3", 1, 2)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 3)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[0], "SF2-1", 2, 3)
		assertMatch(t, matches[1], "SF2-2", 2, 3)
		assertMatch(t, matches[2], "SF2-3", 2, 3)
		assertMatch(t, matches[3], "F-1", 1, 0)
		assertMatch(t, matches[4], "F-2", 1, 0)
		assertMatch(t, matches[5], "F-3", 1, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 4)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[0], "SF1-1", 1, 4)
		assertMatch(t, matches[1], "SF2-1", 2, 3)
		assertMatch(t, matches[2], "SF1-2", 1, 4)
		assertMatch(t, matches[3], "SF2-2", 2, 3)
		assertMatch(t, matches[4], "SF1-3", 1, 4)
		assertMatch(t, matches[5], "SF2-3", 2, 3)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 5)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF2-2", 4, 5)
		assertMatch(t, matches[2], "QF2-3", 4, 5)
		assertMatch(t, matches[3], "SF1-1", 1, 0)
		assertMatch(t, matches[4], "SF2-1", 2, 3)
		assertMatch(t, matches[5], "SF1-2", 1, 0)
		assertMatch(t, matches[6], "SF2-2", 2, 3)
		assertMatch(t, matches[7], "SF1-3", 1, 0)
		assertMatch(t, matches[8], "SF2-3", 2, 3)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 6)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF4-1", 3, 6)
		assertMatch(t, matches[2], "QF2-2", 4, 5)
		assertMatch(t, matches[3], "QF4-2", 3, 6)
		assertMatch(t, matches[4], "QF2-3", 4, 5)
		assertMatch(t, matches[5], "QF4-3", 3, 6)
		assertMatch(t, matches[6], "SF1-1", 1, 0)
		assertMatch(t, matches[7], "SF2-1", 2, 0)
		assertMatch(t, matches[8], "SF1-2", 1, 0)
		assertMatch(t, matches[9], "SF2-2", 2, 0)
		assertMatch(t, matches[10], "SF1-3", 1, 0)
		assertMatch(t, matches[11], "SF2-3", 2, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 7)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[0], "QF2-1", 4, 5)
		assertMatch(t, matches[1], "QF3-1", 2, 7)
		assertMatch(t, matches[2], "QF4-1", 3, 6)
		assertMatch(t, matches[3], "QF2-2", 4, 5)
		assertMatch(t, matches[4], "QF3-2", 2, 7)
		assertMatch(t, matches[5], "QF4-2", 3, 6)
		assertMatch(t, matches[6], "QF2-3", 4, 5)
		assertMatch(t, matches[7], "QF3-3", 2, 7)
		assertMatch(t, matches[8], "QF4-3", 3, 6)
		assertMatch(t, matches[9], "SF1-1", 1, 0)
		assertMatch(t, matches[10], "SF1-2", 1, 0)
		assertMatch(t, matches[11], "SF1-3", 1, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 8)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[0], "QF1-1", 1, 8)
		assertMatch(t, matches[1], "QF2-1", 4, 5)
		assertMatch(t, matches[2], "QF3-1", 2, 7)
		assertMatch(t, matches[3], "QF4-1", 3, 6)
		assertMatch(t, matches[4], "QF1-2", 1, 8)
		assertMatch(t, matches[5], "QF2-2", 4, 5)
		assertMatch(t, matches[6], "QF3-2", 2, 7)
		assertMatch(t, matches[7], "QF4-2", 3, 6)
		assertMatch(t, matches[8], "QF1-3", 1, 8)
		assertMatch(t, matches[9], "QF2-3", 4, 5)
		assertMatch(t, matches[10], "QF3-3", 2, 7)
		assertMatch(t, matches[11], "QF4-3", 3, 6)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 9)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 15, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF2-2", 8, 9)
		assertMatch(t, matches[2], "EF2-3", 8, 9)
		assertMatch(t, matches[3], "QF1-1", 1, 0)
		assertMatch(t, matches[4], "QF2-1", 4, 5)
		assertMatch(t, matches[5], "QF3-1", 2, 7)
		assertMatch(t, matches[6], "QF4-1", 3, 6)
		assertMatch(t, matches[7], "QF1-2", 1, 0)
		assertMatch(t, matches[8], "QF2-2", 4, 5)
		assertMatch(t, matches[9], "QF3-2", 2, 7)
		assertMatch(t, matches[10], "QF4-2", 3, 6)
		assertMatch(t, matches[11], "QF1-3", 1, 0)
		assertMatch(t, matches[12], "QF2-3", 4, 5)
		assertMatch(t, matches[13], "QF3-3", 2, 7)
		assertMatch(t, matches[14], "QF4-3", 3, 6)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 10)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 18, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF6-1", 7, 10)
		assertMatch(t, matches[2], "EF2-2", 8, 9)
		assertMatch(t, matches[3], "EF6-2", 7, 10)
		assertMatch(t, matches[4], "EF2-3", 8, 9)
		assertMatch(t, matches[5], "EF6-3", 7, 10)
		assertMatch(t, matches[6], "QF1-1", 1, 0)
		assertMatch(t, matches[7], "QF2-1", 4, 5)
		assertMatch(t, matches[8], "QF3-1", 2, 0)
		assertMatch(t, matches[9], "QF4-1", 3, 6)
		assertMatch(t, matches[10], "QF1-2", 1, 0)
		assertMatch(t, matches[11], "QF2-2", 4, 5)
		assertMatch(t, matches[12], "QF3-2", 2, 0)
		assertMatch(t, matches[13], "QF4-2", 3, 6)
		assertMatch(t, matches[14], "QF1-3", 1, 0)
		assertMatch(t, matches[15], "QF2-3", 4, 5)
		assertMatch(t, matches[16], "QF3-3", 2, 0)
		assertMatch(t, matches[17], "QF4-3", 3, 6)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 11)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 21, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF6-1", 7, 10)
		assertMatch(t, matches[2], "EF8-1", 6, 11)
		assertMatch(t, matches[3], "EF2-2", 8, 9)
		assertMatch(t, matches[4], "EF6-2", 7, 10)
		assertMatch(t, matches[5], "EF8-2", 6, 11)
		assertMatch(t, matches[6], "EF2-3", 8, 9)
		assertMatch(t, matches[7], "EF6-3", 7, 10)
		assertMatch(t, matches[8], "EF8-3", 6, 11)
		assertMatch(t, matches[9], "QF1-1", 1, 0)
		assertMatch(t, matches[10], "QF2-1", 4, 5)
		assertMatch(t, matches[11], "QF3-1", 2, 0)
		assertMatch(t, matches[12], "QF4-1", 3, 0)
		assertMatch(t, matches[13], "QF1-2", 1, 0)
		assertMatch(t, matches[14], "QF2-2", 4, 5)
		assertMatch(t, matches[15], "QF3-2", 2, 0)
		assertMatch(t, matches[16], "QF4-2", 3, 0)
		assertMatch(t, matches[17], "QF1-3", 1, 0)
		assertMatch(t, matches[18], "QF2-3", 4, 5)
		assertMatch(t, matches[19], "QF3-3", 2, 0)
		assertMatch(t, matches[20], "QF4-3", 3, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 12)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 24, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF4-1", 5, 12)
		assertMatch(t, matches[2], "EF6-1", 7, 10)
		assertMatch(t, matches[3], "EF8-1", 6, 11)
		assertMatch(t, matches[4], "EF2-2", 8, 9)
		assertMatch(t, matches[5], "EF4-2", 5, 12)
		assertMatch(t, matches[6], "EF6-2", 7, 10)
		assertMatch(t, matches[7], "EF8-2", 6, 11)
		assertMatch(t, matches[8], "EF2-3", 8, 9)
		assertMatch(t, matches[9], "EF4-3", 5, 12)
		assertMatch(t, matches[10], "EF6-3", 7, 10)
		assertMatch(t, matches[11], "EF8-3", 6, 11)
		assertMatch(t, matches[12], "QF1-1", 1, 0)
		assertMatch(t, matches[13], "QF2-1", 4, 0)
		assertMatch(t, matches[14], "QF3-1", 2, 0)
		assertMatch(t, matches[15], "QF4-1", 3, 0)
		assertMatch(t, matches[16], "QF1-2", 1, 0)
		assertMatch(t, matches[17], "QF2-2", 4, 0)
		assertMatch(t, matches[18], "QF3-2", 2, 0)
		assertMatch(t, matches[19], "QF4-2", 3, 0)
		assertMatch(t, matches[20], "QF1-3", 1, 0)
		assertMatch(t, matches[21], "QF2-3", 4, 0)
		assertMatch(t, matches[22], "QF3-3", 2, 0)
		assertMatch(t, matches[23], "QF4-3", 3, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 13)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 24, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF6-1", 7, 10)
		assertMatch(t, matches[4], "EF8-1", 6, 11)
		assertMatch(t, matches[5], "EF2-2", 8, 9)
		assertMatch(t, matches[6], "EF3-2", 4, 13)
		assertMatch(t, matches[7], "EF4-2", 5, 12)
		assertMatch(t, matches[8], "EF6-2", 7, 10)
		assertMatch(t, matches[9], "EF8-2", 6, 11)
		assertMatch(t, matches[10], "EF2-3", 8, 9)
		assertMatch(t, matches[11], "EF3-3", 4, 13)
		assertMatch(t, matches[12], "EF4-3", 5, 12)
		assertMatch(t, matches[13], "EF6-3", 7, 10)
		assertMatch(t, matches[14], "EF8-3", 6, 11)
		assertMatch(t, matches[15], "QF1-1", 1, 0)
		assertMatch(t, matches[16], "QF3-1", 2, 0)
		assertMatch(t, matches[17], "QF4-1", 3, 0)
		assertMatch(t, matches[18], "QF1-2", 1, 0)
		assertMatch(t, matches[19], "QF3-2", 2, 0)
		assertMatch(t, matches[20], "QF4-2", 3, 0)
		assertMatch(t, matches[21], "QF1-3", 1, 0)
		assertMatch(t, matches[22], "QF3-3", 2, 0)
		assertMatch(t, matches[23], "QF4-3", 3, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 14)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 24, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF6-1", 7, 10)
		assertMatch(t, matches[4], "EF7-1", 3, 14)
		assertMatch(t, matches[5], "EF8-1", 6, 11)
		assertMatch(t, matches[6], "EF2-2", 8, 9)
		assertMatch(t, matches[7], "EF3-2", 4, 13)
		assertMatch(t, matches[8], "EF4-2", 5, 12)
		assertMatch(t, matches[9], "EF6-2", 7, 10)
		assertMatch(t, matches[10], "EF7-2", 3, 14)
		assertMatch(t, matches[11], "EF8-2", 6, 11)
		assertMatch(t, matches[12], "EF2-3", 8, 9)
		assertMatch(t, matches[13], "EF3-3", 4, 13)
		assertMatch(t, matches[14], "EF4-3", 5, 12)
		assertMatch(t, matches[15], "EF6-3", 7, 10)
		assertMatch(t, matches[16], "EF7-3", 3, 14)
		assertMatch(t, matches[17], "EF8-3", 6, 11)
		assertMatch(t, matches[18], "QF1-1", 1, 0)
		assertMatch(t, matches[19], "QF3-1", 2, 0)
		assertMatch(t, matches[20], "QF1-2", 1, 0)
		assertMatch(t, matches[21], "QF3-2", 2, 0)
		assertMatch(t, matches[22], "QF1-3", 1, 0)
		assertMatch(t, matches[23], "QF3-3", 2, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 15)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 24, len(matches)) {
		assertMatch(t, matches[0], "EF2-1", 8, 9)
		assertMatch(t, matches[1], "EF3-1", 4, 13)
		assertMatch(t, matches[2], "EF4-1", 5, 12)
		assertMatch(t, matches[3], "EF5-1", 2, 15)
		assertMatch(t, matches[4], "EF6-1", 7, 10)
		assertMatch(t, matches[5], "EF7-1", 3, 14)
		assertMatch(t, matches[6], "EF8-1", 6, 11)
		assertMatch(t, matches[7], "EF2-2", 8, 9)
		assertMatch(t, matches[8], "EF3-2", 4, 13)
		assertMatch(t, matches[9], "EF4-2", 5, 12)
		assertMatch(t, matches[10], "EF5-2", 2, 15)
		assertMatch(t, matches[11], "EF6-2", 7, 10)
		assertMatch(t, matches[12], "EF7-2", 3, 14)
		assertMatch(t, matches[13], "EF8-2", 6, 11)
		assertMatch(t, matches[14], "EF2-3", 8, 9)
		assertMatch(t, matches[15], "EF3-3", 4, 13)
		assertMatch(t, matches[16], "EF4-3", 5, 12)
		assertMatch(t, matches[17], "EF5-3", 2, 15)
		assertMatch(t, matches[18], "EF6-3", 7, 10)
		assertMatch(t, matches[19], "EF7-3", 3, 14)
		assertMatch(t, matches[20], "EF8-3", 6, 11)
		assertMatch(t, matches[21], "QF1-1", 1, 0)
		assertMatch(t, matches[22], "QF1-2", 1, 0)
		assertMatch(t, matches[23], "QF1-3", 1, 0)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()

	createTestAlliances(db, 16)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 24, len(matches)) {
		assertMatch(t, matches[0], "EF1-1", 1, 16)
		assertMatch(t, matches[1], "EF2-1", 8, 9)
		assertMatch(t, matches[2], "EF3-1", 4, 13)
		assertMatch(t, matches[3], "EF4-1", 5, 12)
		assertMatch(t, matches[4], "EF5-1", 2, 15)
		assertMatch(t, matches[5], "EF6-1", 7, 10)
		assertMatch(t, matches[6], "EF7-1", 3, 14)
		assertMatch(t, matches[7], "EF8-1", 6, 11)
		assertMatch(t, matches[8], "EF1-2", 1, 16)
		assertMatch(t, matches[9], "EF2-2", 8, 9)
		assertMatch(t, matches[10], "EF3-2", 4, 13)
		assertMatch(t, matches[11], "EF4-2", 5, 12)
		assertMatch(t, matches[12], "EF5-2", 2, 15)
		assertMatch(t, matches[13], "EF6-2", 7, 10)
		assertMatch(t, matches[14], "EF7-2", 3, 14)
		assertMatch(t, matches[15], "EF8-2", 6, 11)
		assertMatch(t, matches[16], "EF1-3", 1, 16)
		assertMatch(t, matches[17], "EF2-3", 8, 9)
		assertMatch(t, matches[18], "EF3-3", 4, 13)
		assertMatch(t, matches[19], "EF4-3", 5, 12)
		assertMatch(t, matches[20], "EF5-3", 2, 15)
		assertMatch(t, matches[21], "EF6-3", 7, 10)
		assertMatch(t, matches[22], "EF7-3", 3, 14)
		assertMatch(t, matches[23], "EF8-3", 6, 11)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()
}

func TestEliminationScheduleErrors(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 1)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have at least 2 alliances", err.Error())
	}
	db.TruncateAllianceTeams()

	createTestAlliances(db, 17)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Round of depth 32 is not supported", err.Error())
	}
	db.TruncateAllianceTeams()

	db.CreateAllianceTeam(&AllianceTeam{0, 1, 0, 1})
	db.CreateAllianceTeam(&AllianceTeam{0, 1, 1, 2})
	db.CreateAllianceTeam(&AllianceTeam{0, 2, 0, 3})
	db.CreateAllianceTeam(&AllianceTeam{0, 2, 1, 4})
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Alliances must consist of at least 3 teams", err.Error())
	}
	db.TruncateAllianceTeams()
}

func TestEliminationSchedulePopulatePartialMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	// Final should be updated after semifinal is concluded.
	createTestAlliances(db, 3)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "SF2-1", "B")
	scoreMatch(db, "SF2-2", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err := db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 5, len(matches)) {
		assertMatch(t, matches[2], "F-1", 1, 3)
		assertMatch(t, matches[3], "F-2", 1, 3)
		assertMatch(t, matches[4], "F-3", 1, 3)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()
	db.TruncateMatchResults()

	// Final should be generated and populated as both semifinals conclude.
	createTestAlliances(db, 4)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "SF2-1", "R")
	scoreMatch(db, "SF2-2", "R")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[5], "F-1", 0, 2)
		assertMatch(t, matches[6], "F-2", 0, 2)
		assertMatch(t, matches[7], "F-3", 0, 2)
	}
	scoreMatch(db, "SF1-1", "R")
	scoreMatch(db, "SF1-2", "R")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 7, len(matches)) {
		assertMatch(t, matches[4], "F-1", 1, 2)
		assertMatch(t, matches[5], "F-2", 1, 2)
		assertMatch(t, matches[6], "F-3", 1, 2)
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()
	db.TruncateMatchResults()
}

func TestEliminationScheduleCreateNextRound(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 4)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "SF1-1", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ := db.GetMatchesByType("elimination")
	assert.Equal(t, 6, len(matches))
	scoreMatch(db, "SF2-1", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 6, len(matches))
	scoreMatch(db, "SF1-2", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 8, len(matches))
	scoreMatch(db, "SF2-2", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = db.GetMatchesByType("elimination")
	if assert.Equal(t, 7, len(matches)) {
		assertMatch(t, matches[4], "F-1", 4, 3)
		assertMatch(t, matches[5], "F-2", 4, 3)
		assertMatch(t, matches[6], "F-3", 4, 3)
	}
}

func TestEliminationScheduleDetermineWinner(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	// Round with one tie and a sweep.
	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "T")
	winner, err := db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ := db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-2", "B")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-3", "B")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.Nil(t, err) {
		if assert.Equal(t, 3, len(winner)) {
			assert.Equal(t, 2, winner[0].TeamId)
		}
	}
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	db.TruncateAllianceTeams()
	db.TruncateMatches()
	db.TruncateMatchResults()

	// Round with one tie and a split.
	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "R")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-2", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-3", "B")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))
	assert.Equal(t, "F-4", matches[3].DisplayName)
	scoreMatch(db, "F-4", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	scoreMatch(db, "F-5", "R")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.Nil(t, err) {
		if assert.Equal(t, 3, len(winner)) {
			assert.Equal(t, 1, winner[0].TeamId)
		}
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()
	db.TruncateMatchResults()

	// Round with two ties.
	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-2", "B")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 3, len(matches))
	scoreMatch(db, "F-3", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	assert.Empty(t, winner)
	matches, _ = db.GetMatchesByType("elimination")
	assert.Equal(t, 5, len(matches))
	assert.Equal(t, "F-4", matches[3].DisplayName)
	assert.Equal(t, "F-5", matches[4].DisplayName)
	scoreMatch(db, "F-4", "B")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.Nil(t, err) {
		if assert.Equal(t, 3, len(winner)) {
			assert.Equal(t, 2, winner[0].TeamId)
		}
	}
	db.TruncateAllianceTeams()
	db.TruncateMatches()
	db.TruncateMatchResults()

	// Round with repeated ties.
	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "T")
	scoreMatch(db, "F-2", "T")
	scoreMatch(db, "F-3", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-4", "T")
	scoreMatch(db, "F-5", "T")
	scoreMatch(db, "F-6", "T")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-7", "R")
	scoreMatch(db, "F-8", "B")
	scoreMatch(db, "F-9", "R")
	winner, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.Nil(t, err) {
		if assert.Equal(t, 3, len(winner)) {
			assert.Equal(t, 1, winner[0].TeamId)
		}
	}
}

func TestEliminationScheduleRemoveUnneededMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "R")
	scoreMatch(db, "F-2", "R")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ := db.GetMatchesByType("elimination")
	assert.Equal(t, 2, len(matches))

	// Check that the deleted match is recreated if the score is changed.
	scoreMatch(db, "F-2", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, _ = db.GetMatchesByType("elimination")
	if assert.Equal(t, 3, len(matches)) {
		assert.Equal(t, "F-3", matches[2].DisplayName)
	}
}

func TestEliminationScheduleChangePreviousRoundResult(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 4)
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(db, "SF2-1", "R")
	scoreMatch(db, "SF2-2", "B")
	scoreMatch(db, "SF2-3", "R")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(db, "SF2-3", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err := db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[6], "F-1", 0, 3)
		assertMatch(t, matches[7], "F-2", 0, 3)
		assertMatch(t, matches[8], "F-3", 0, 3)
	}

	scoreMatch(db, "SF1-1", "R")
	scoreMatch(db, "SF1-2", "R")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(db, "SF1-2", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	scoreMatch(db, "SF1-3", "B")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	assert.Nil(t, err)
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[6], "F-1", 4, 3)
		assertMatch(t, matches[7], "F-2", 4, 3)
		assertMatch(t, matches[8], "F-3", 4, 3)
	}
}

func TestEliminationScheduleUnscoredMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 2)
	db.UpdateEliminationSchedule(time.Unix(0, 0))
	scoreMatch(db, "F-1", "blorpy")
	_, err = db.UpdateEliminationSchedule(time.Unix(0, 0))
	if assert.NotNil(t, err) {
		assert.Equal(t, "Completed match 1 has invalid winner 'blorpy'", err.Error())
	}
}

func TestEliminationScheduleTiming(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	createTestAlliances(db, 4)
	db.UpdateEliminationSchedule(time.Unix(1000, 0))
	matches, err := db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assert.True(t, time.Unix(1000, 0).Equal(matches[0].Time))
		assert.True(t, time.Unix(1600, 0).Equal(matches[1].Time))
		assert.True(t, time.Unix(2200, 0).Equal(matches[2].Time))
		assert.True(t, time.Unix(2800, 0).Equal(matches[3].Time))
		assert.True(t, time.Unix(3400, 0).Equal(matches[4].Time))
		assert.True(t, time.Unix(4000, 0).Equal(matches[5].Time))
	}
	scoreMatch(db, "SF1-1", "R")
	scoreMatch(db, "SF1-3", "B")
	db.UpdateEliminationSchedule(time.Unix(5000, 0))
	matches, err = db.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assert.True(t, time.Unix(1000, 0).Equal(matches[0].Time))
		assert.True(t, time.Unix(5000, 0).Equal(matches[1].Time))
		assert.True(t, time.Unix(5600, 0).Equal(matches[2].Time))
		assert.True(t, time.Unix(6200, 0).Equal(matches[3].Time))
		assert.True(t, time.Unix(3400, 0).Equal(matches[4].Time))
		assert.True(t, time.Unix(6800, 0).Equal(matches[5].Time))
	}
}

func createTestAlliances(db *Database, allianceCount int) {
	for i := 1; i <= allianceCount; i++ {
		db.CreateAllianceTeam(&AllianceTeam{0, i, 0, i})
		db.CreateAllianceTeam(&AllianceTeam{0, i, 1, i})
		db.CreateAllianceTeam(&AllianceTeam{0, i, 2, i})
	}
}

func assertMatch(t *testing.T, match Match, displayName string, redAlliance int, blueAlliance int) {
	assert.Equal(t, displayName, match.DisplayName)
	assert.Equal(t, redAlliance, match.Red1)
	assert.Equal(t, blueAlliance, match.Blue1)
}

func scoreMatch(db *Database, displayName string, winner string) {
	match, _ := db.GetMatchByName("elimination", displayName)
	match.Status = "complete"
	match.Winner = winner
	db.SaveMatch(match)
}
