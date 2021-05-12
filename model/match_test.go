// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNonexistentMatch(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match, err := db.GetMatchById(1114)
	assert.Nil(t, err)
	assert.Nil(t, match)
}

func TestMatchCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{0, "qualification", "254", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), MatchNotPlayed}
	db.CreateMatch(&match)
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match, *match2)
	match3, err := db.GetMatchByName("qualification", "254")
	assert.Nil(t, err)
	assert.Equal(t, match, *match3)

	match.Status = RedWonMatch
	db.UpdateMatch(&match)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Equal(t, match.Status, match2.Status)

	db.DeleteMatch(match.Id)
	match2, err = db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestTruncateMatches(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{0, "qualification", "254", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), MatchNotPlayed}
	db.CreateMatch(&match)
	db.TruncateMatches()
	match2, err := db.GetMatchById(1)
	assert.Nil(t, err)
	assert.Nil(t, match2)
}

func TestGetMatchesByElimRoundGroup(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{Type: "elimination", DisplayName: "SF1-1", ElimRound: 2, ElimGroup: 1, ElimInstance: 1,
		ElimRedAlliance: 8, ElimBlueAlliance: 4}
	db.CreateMatch(&match)
	match2 := Match{Type: "elimination", DisplayName: "SF2-2", ElimRound: 2, ElimGroup: 2, ElimInstance: 2,
		ElimRedAlliance: 2, ElimBlueAlliance: 3}
	db.CreateMatch(&match2)
	match3 := Match{Type: "elimination", DisplayName: "SF2-1", ElimRound: 2, ElimGroup: 2, ElimInstance: 1,
		ElimRedAlliance: 8, ElimBlueAlliance: 4}
	db.CreateMatch(&match3)
	match4 := Match{Type: "elimination", DisplayName: "QF2-1", ElimRound: 4, ElimGroup: 2, ElimInstance: 1,
		ElimRedAlliance: 4, ElimBlueAlliance: 5}
	db.CreateMatch(&match4)
	match5 := Match{Type: "practice", DisplayName: "1"}
	db.CreateMatch(&match5)

	matches, err := db.GetMatchesByElimRoundGroup(4, 1)
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByElimRoundGroup(2, 2)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(matches)) {
		assert.Equal(t, "SF2-1", matches[0].DisplayName)
		assert.Equal(t, "SF2-2", matches[1].DisplayName)
	}
}

func TestGetMatchesByType(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	match := Match{0, "qualification", "1", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false,
		5, false, 6, false, time.Now().UTC(), time.Now().UTC(), MatchNotPlayed}
	db.CreateMatch(&match)
	match2 := Match{0, "practice", "1", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, time.Now().UTC(), time.Now().UTC(), MatchNotPlayed}
	db.CreateMatch(&match2)
	match3 := Match{0, "practice", "2", time.Now().UTC(), 0, 0, 0, 0, 0, 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, time.Now().UTC(), time.Now().UTC(), MatchNotPlayed}
	db.CreateMatch(&match3)

	matches, err := db.GetMatchesByType("test")
	assert.Nil(t, err)
	assert.Empty(t, matches)
	matches, err = db.GetMatchesByType("practice")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(matches))
	matches, err = db.GetMatchesByType("qualification")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))
}

func TestTbaCode(t *testing.T) {
	match := Match{Type: "practice", DisplayName: "3"}
	assert.Equal(t, "", match.TbaCode())
	match = Match{Type: "qualification", DisplayName: "26"}
	assert.Equal(t, "qm26", match.TbaCode())
	match = Match{Type: "elimination", DisplayName: "EF2-1", ElimRound: 8, ElimGroup: 2, ElimInstance: 1}
	assert.Equal(t, "ef2m1", match.TbaCode())
	match = Match{Type: "elimination", DisplayName: "QF3-2", ElimRound: 4, ElimGroup: 3, ElimInstance: 2}
	assert.Equal(t, "qf3m2", match.TbaCode())
	match = Match{Type: "elimination", DisplayName: "SF1-3", ElimRound: 2, ElimGroup: 1, ElimInstance: 3}
	assert.Equal(t, "sf1m3", match.TbaCode())
	match = Match{Type: "elimination", DisplayName: "F2", ElimRound: 1, ElimGroup: 1, ElimInstance: 2}
	assert.Equal(t, "f1m2", match.TbaCode())
}
