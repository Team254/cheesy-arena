// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var dummyStartTime = time.Unix(0, 0)

func TestSingleEliminationInitialWith2Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(2)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	assertFullFinals(t, matchSpecs, 0)

	finalMatchup.update(map[int]playoffMatchResult{})
	for i := 0; i < 6; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 2}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "F")
}
func TestSingleEliminationInitialWith3Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(3)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 9, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:3],
			[]expectedMatchSpec{
				{"Semifinal 2-1", "SF2-1", "", 38, "SF2", true, false, "sf", 2, 1},
				{"Semifinal 2-2", "SF2-2", "", 40, "SF2", true, false, "sf", 2, 2},
				{"Semifinal 2-3", "SF2-3", "", 42, "SF2", true, false, "sf", 2, 3},
			},
		)
	}
	assertFullFinals(t, matchSpecs, 3)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:3],
		[]expectedAlliances{
			{2, 3},
			{2, 3},
			{2, 3},
		},
	)
	for i := 3; i < 9; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "SF2", "F")
}

func TestSingleEliminationInitialWith4Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(4)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	assertFullSemifinalsOnward(t, matchSpecs, 0)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:6],
		[]expectedAlliances{
			{1, 4},
			{2, 3},
			{1, 4},
			{2, 3},
			{1, 4},
			{2, 3},
		},
	)
	for i := 6; i < 12; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith5Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(5)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 15, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:3],
			[]expectedMatchSpec{
				{"Quarterfinal 2-1", "QF2-1", "", 26, "QF2", true, false, "qf", 2, 1},
				{"Quarterfinal 2-2", "QF2-2", "", 30, "QF2", true, false, "qf", 2, 2},
				{"Quarterfinal 2-3", "QF2-3", "", 34, "QF2", true, false, "qf", 2, 3},
			},
		)
	}
	assertFullSemifinalsOnward(t, matchSpecs, 3)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:9],
		[]expectedAlliances{
			{4, 5},
			{4, 5},
			{4, 5},
			{1, 0},
			{2, 3},
			{1, 0},
			{2, 3},
			{1, 0},
			{2, 3},
		},
	)
	for i := 9; i < 15; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "QF2", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith6Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(6)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 18, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:6],
			[]expectedMatchSpec{
				{"Quarterfinal 2-1", "QF2-1", "", 26, "QF2", true, false, "qf", 2, 1},
				{"Quarterfinal 4-1", "QF4-1", "", 28, "QF4", true, false, "qf", 4, 1},
				{"Quarterfinal 2-2", "QF2-2", "", 30, "QF2", true, false, "qf", 2, 2},
				{"Quarterfinal 4-2", "QF4-2", "", 32, "QF4", true, false, "qf", 4, 2},
				{"Quarterfinal 2-3", "QF2-3", "", 34, "QF2", true, false, "qf", 2, 3},
				{"Quarterfinal 4-3", "QF4-3", "", 36, "QF4", true, false, "qf", 4, 3},
			},
		)
	}
	assertFullSemifinalsOnward(t, matchSpecs, 6)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:12],
		[]expectedAlliances{
			{4, 5},
			{3, 6},
			{4, 5},
			{3, 6},
			{4, 5},
			{3, 6},
			{1, 0},
			{2, 0},
			{1, 0},
			{2, 0},
			{1, 0},
			{2, 0},
		},
	)
	for i := 12; i < 18; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "QF2", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith7Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(7)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 21, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:9],
			[]expectedMatchSpec{
				{"Quarterfinal 2-1", "QF2-1", "", 26, "QF2", true, false, "qf", 2, 1},
				{"Quarterfinal 3-1", "QF3-1", "", 27, "QF3", true, false, "qf", 3, 1},
				{"Quarterfinal 4-1", "QF4-1", "", 28, "QF4", true, false, "qf", 4, 1},
				{"Quarterfinal 2-2", "QF2-2", "", 30, "QF2", true, false, "qf", 2, 2},
				{"Quarterfinal 3-2", "QF3-2", "", 31, "QF3", true, false, "qf", 3, 2},
				{"Quarterfinal 4-2", "QF4-2", "", 32, "QF4", true, false, "qf", 4, 2},
				{"Quarterfinal 2-3", "QF2-3", "", 34, "QF2", true, false, "qf", 2, 3},
				{"Quarterfinal 3-3", "QF3-3", "", 35, "QF3", true, false, "qf", 3, 3},
				{"Quarterfinal 4-3", "QF4-3", "", 36, "QF4", true, false, "qf", 4, 3},
			},
		)
	}
	assertFullSemifinalsOnward(t, matchSpecs, 9)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:15],
		[]expectedAlliances{
			{4, 5},
			{2, 7},
			{3, 6},
			{4, 5},
			{2, 7},
			{3, 6},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 0},
			{0, 0},
			{1, 0},
			{0, 0},
			{1, 0},
			{0, 0},
		},
	)
	for i := 15; i < 21; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith8Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(8)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	assertFullQuarterfinalsOnward(t, matchSpecs, 0)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:12],
		[]expectedAlliances{
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
		},
	)
	for i := 12; i < 24; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith9Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(9)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 27, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:3],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 3)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:15],
		[]expectedAlliances{
			{8, 9},
			{8, 9},
			{8, 9},
			{1, 0},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 0},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 0},
			{4, 5},
			{2, 7},
			{3, 6},
		},
	)
	for i := 15; i < 27; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "EF2", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith10Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(10)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 30, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:6],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 6)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:18],
		[]expectedAlliances{
			{8, 9},
			{7, 10},
			{8, 9},
			{7, 10},
			{8, 9},
			{7, 10},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 6},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 6},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 6},
		},
	)
	for i := 18; i < 30; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "EF2", "EF6", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith11Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(11)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 33, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:9],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 9)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:21],
		[]expectedAlliances{
			{8, 9},
			{7, 10},
			{6, 11},
			{8, 9},
			{7, 10},
			{6, 11},
			{8, 9},
			{7, 10},
			{6, 11},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 0},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 0},
			{1, 0},
			{4, 5},
			{2, 0},
			{3, 0},
		},
	)
	for i := 21; i < 33; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "EF2", "EF6", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith12Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(12)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 36, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:12],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 4-1", "EF4-1", "", 4, "EF4", true, false, "ef", 4, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 4-2", "EF4-2", "", 12, "EF4", true, false, "ef", 4, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 4-3", "EF4-3", "", 20, "EF4", true, false, "ef", 4, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 12)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:24],
		[]expectedAlliances{
			{8, 9},
			{5, 12},
			{7, 10},
			{6, 11},
			{8, 9},
			{5, 12},
			{7, 10},
			{6, 11},
			{8, 9},
			{5, 12},
			{7, 10},
			{6, 11},
			{1, 0},
			{4, 0},
			{2, 0},
			{3, 0},
			{1, 0},
			{4, 0},
			{2, 0},
			{3, 0},
			{1, 0},
			{4, 0},
			{2, 0},
			{3, 0},
		},
	)
	for i := 24; i < 36; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "EF2", "EF4", "EF6", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith13Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(13)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 39, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:15],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 3-1", "EF3-1", "", 3, "EF3", true, false, "ef", 3, 1},
				{"Eighthfinal 4-1", "EF4-1", "", 4, "EF4", true, false, "ef", 4, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 3-2", "EF3-2", "", 11, "EF3", true, false, "ef", 3, 2},
				{"Eighthfinal 4-2", "EF4-2", "", 12, "EF4", true, false, "ef", 4, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 3-3", "EF3-3", "", 19, "EF3", true, false, "ef", 3, 3},
				{"Eighthfinal 4-3", "EF4-3", "", 20, "EF4", true, false, "ef", 4, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 15)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:27],
		[]expectedAlliances{
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{6, 11},
			{1, 0},
			{0, 0},
			{2, 0},
			{3, 0},
			{1, 0},
			{0, 0},
			{2, 0},
			{3, 0},
			{1, 0},
			{0, 0},
			{2, 0},
			{3, 0},
		},
	)
	for i := 27; i < 39; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "EF2", "EF3", "EF4", "EF6", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
}

func TestSingleEliminationInitialWith14Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(14)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 42, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:18],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 3-1", "EF3-1", "", 3, "EF3", true, false, "ef", 3, 1},
				{"Eighthfinal 4-1", "EF4-1", "", 4, "EF4", true, false, "ef", 4, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 7-1", "EF7-1", "", 7, "EF7", true, false, "ef", 7, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 3-2", "EF3-2", "", 11, "EF3", true, false, "ef", 3, 2},
				{"Eighthfinal 4-2", "EF4-2", "", 12, "EF4", true, false, "ef", 4, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 7-2", "EF7-2", "", 15, "EF7", true, false, "ef", 7, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 3-3", "EF3-3", "", 19, "EF3", true, false, "ef", 3, 3},
				{"Eighthfinal 4-3", "EF4-3", "", 20, "EF4", true, false, "ef", 4, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 7-3", "EF7-3", "", 23, "EF7", true, false, "ef", 7, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 18)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:30],
		[]expectedAlliances{
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{3, 14},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{3, 14},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{7, 10},
			{3, 14},
			{6, 11},
			{1, 0},
			{0, 0},
			{2, 0},
			{0, 0},
			{1, 0},
			{0, 0},
			{2, 0},
			{0, 0},
			{1, 0},
			{0, 0},
			{2, 0},
			{0, 0},
		},
	)
	for i := 30; i < 42; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(
		t, matchGroups, "EF2", "EF3", "EF4", "EF6", "EF7", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F",
	)
}

func TestSingleEliminationInitialWith15Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(15)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 45, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:21],
			[]expectedMatchSpec{
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 3-1", "EF3-1", "", 3, "EF3", true, false, "ef", 3, 1},
				{"Eighthfinal 4-1", "EF4-1", "", 4, "EF4", true, false, "ef", 4, 1},
				{"Eighthfinal 5-1", "EF5-1", "", 5, "EF5", true, false, "ef", 5, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 7-1", "EF7-1", "", 7, "EF7", true, false, "ef", 7, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 3-2", "EF3-2", "", 11, "EF3", true, false, "ef", 3, 2},
				{"Eighthfinal 4-2", "EF4-2", "", 12, "EF4", true, false, "ef", 4, 2},
				{"Eighthfinal 5-2", "EF5-2", "", 13, "EF5", true, false, "ef", 5, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 7-2", "EF7-2", "", 15, "EF7", true, false, "ef", 7, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 3-3", "EF3-3", "", 19, "EF3", true, false, "ef", 3, 3},
				{"Eighthfinal 4-3", "EF4-3", "", 20, "EF4", true, false, "ef", 4, 3},
				{"Eighthfinal 5-3", "EF5-3", "", 21, "EF5", true, false, "ef", 5, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 7-3", "EF7-3", "", 23, "EF7", true, false, "ef", 7, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 21)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:33],
		[]expectedAlliances{
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
			{1, 0},
			{0, 0},
			{0, 0},
			{0, 0},
			{1, 0},
			{0, 0},
			{0, 0},
			{0, 0},
			{1, 0},
			{0, 0},
			{0, 0},
			{0, 0},
		},
	)
	for i := 33; i < 45; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(
		t, matchGroups, "EF2", "EF3", "EF4", "EF5", "EF6", "EF7", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F",
	)
}

func TestSingleEliminationInitialWith16Alliances(t *testing.T) {
	finalMatchup, _, err := newSingleEliminationBracket(16)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 48, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:24],
			[]expectedMatchSpec{
				{"Eighthfinal 1-1", "EF1-1", "", 1, "EF1", true, false, "ef", 1, 1},
				{"Eighthfinal 2-1", "EF2-1", "", 2, "EF2", true, false, "ef", 2, 1},
				{"Eighthfinal 3-1", "EF3-1", "", 3, "EF3", true, false, "ef", 3, 1},
				{"Eighthfinal 4-1", "EF4-1", "", 4, "EF4", true, false, "ef", 4, 1},
				{"Eighthfinal 5-1", "EF5-1", "", 5, "EF5", true, false, "ef", 5, 1},
				{"Eighthfinal 6-1", "EF6-1", "", 6, "EF6", true, false, "ef", 6, 1},
				{"Eighthfinal 7-1", "EF7-1", "", 7, "EF7", true, false, "ef", 7, 1},
				{"Eighthfinal 8-1", "EF8-1", "", 8, "EF8", true, false, "ef", 8, 1},
				{"Eighthfinal 1-2", "EF1-2", "", 9, "EF1", true, false, "ef", 1, 2},
				{"Eighthfinal 2-2", "EF2-2", "", 10, "EF2", true, false, "ef", 2, 2},
				{"Eighthfinal 3-2", "EF3-2", "", 11, "EF3", true, false, "ef", 3, 2},
				{"Eighthfinal 4-2", "EF4-2", "", 12, "EF4", true, false, "ef", 4, 2},
				{"Eighthfinal 5-2", "EF5-2", "", 13, "EF5", true, false, "ef", 5, 2},
				{"Eighthfinal 6-2", "EF6-2", "", 14, "EF6", true, false, "ef", 6, 2},
				{"Eighthfinal 7-2", "EF7-2", "", 15, "EF7", true, false, "ef", 7, 2},
				{"Eighthfinal 8-2", "EF8-2", "", 16, "EF8", true, false, "ef", 8, 2},
				{"Eighthfinal 1-3", "EF1-3", "", 17, "EF1", true, false, "ef", 1, 3},
				{"Eighthfinal 2-3", "EF2-3", "", 18, "EF2", true, false, "ef", 2, 3},
				{"Eighthfinal 3-3", "EF3-3", "", 19, "EF3", true, false, "ef", 3, 3},
				{"Eighthfinal 4-3", "EF4-3", "", 20, "EF4", true, false, "ef", 4, 3},
				{"Eighthfinal 5-3", "EF5-3", "", 21, "EF5", true, false, "ef", 5, 3},
				{"Eighthfinal 6-3", "EF6-3", "", 22, "EF6", true, false, "ef", 6, 3},
				{"Eighthfinal 7-3", "EF7-3", "", 23, "EF7", true, false, "ef", 7, 3},
				{"Eighthfinal 8-3", "EF8-3", "", 24, "EF8", true, false, "ef", 8, 3},
			},
		)
	}
	assertFullQuarterfinalsOnward(t, matchSpecs, 24)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:24],
		[]expectedAlliances{
			{1, 16},
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
			{1, 16},
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
			{1, 16},
			{8, 9},
			{4, 13},
			{5, 12},
			{2, 15},
			{7, 10},
			{3, 14},
			{6, 11},
		},
	)
	for i := 24; i < 48; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(
		t,
		matchGroups,
		"EF1", "EF2", "EF3", "EF4", "EF5", "EF6", "EF7", "EF8", "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F",
	)
}

func TestSingleEliminationErrors(t *testing.T) {
	_, _, err := newSingleEliminationBracket(1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "single-elimination bracket must have at least 2 alliances", err.Error())
	}

	_, _, err = newSingleEliminationBracket(17)
	if assert.NotNil(t, err) {
		assert.Equal(t, "single-elimination bracket must have at most 16 alliances", err.Error())
	}
}

func TestSingleEliminationProgression(t *testing.T) {
	playoffTournament, err := NewPlayoffTournament(model.SingleEliminationPlayoff, 3)
	assert.Nil(t, err)
	finalMatchup := playoffTournament.FinalMatchup()
	matchSpecs := playoffTournament.matchSpecs
	matchGroups := playoffTournament.MatchGroups()
	playoffMatchResults := map[int]playoffMatchResult{}

	assertMatchupOutcome(t, matchGroups["SF2"], "", "")

	playoffMatchResults[38] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 3; i < 9; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 0}})
	}
	assertMatchupOutcome(t, matchGroups["SF2"], "", "")

	playoffMatchResults[40] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 3; i < 9; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 2}})
	}
	assertMatchupOutcome(t, matchGroups["SF2"], "Advances to Final 1", "Eliminated")

	// Reverse a previous outcome.
	playoffMatchResults[40] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 3; i < 9; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 0}})
	}
	assertMatchupOutcome(t, matchGroups["SF2"], "", "")

	playoffMatchResults[42] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 3; i < 9; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{1, 3}})
	}
	assertMatchupOutcome(t, matchGroups["SF2"], "Eliminated", "Advances to Final 1")

	playoffMatchResults[43] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[44] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[45] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, finalMatchup.IsComplete())
	assert.Equal(t, 1, finalMatchup.WinningAllianceId())
	assert.Equal(t, 3, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "Tournament Winner", "Tournament Finalist")

	// Unscore the previous match.
	delete(playoffMatchResults, 45)
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[45] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[46] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, finalMatchup.IsComplete())
	assert.Equal(t, 3, finalMatchup.WinningAllianceId())
	assert.Equal(t, 1, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "Tournament Finalist", "Tournament Winner")
}

func assertFullQuarterfinalsOnward(t *testing.T, matchSpecs []*matchSpec, startingIndex int) {
	if assert.Equal(t, startingIndex+24, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[startingIndex:startingIndex+12],
			[]expectedMatchSpec{
				{"Quarterfinal 1-1", "QF1-1", "", 25, "QF1", true, false, "qf", 1, 1},
				{"Quarterfinal 2-1", "QF2-1", "", 26, "QF2", true, false, "qf", 2, 1},
				{"Quarterfinal 3-1", "QF3-1", "", 27, "QF3", true, false, "qf", 3, 1},
				{"Quarterfinal 4-1", "QF4-1", "", 28, "QF4", true, false, "qf", 4, 1},
				{"Quarterfinal 1-2", "QF1-2", "", 29, "QF1", true, false, "qf", 1, 2},
				{"Quarterfinal 2-2", "QF2-2", "", 30, "QF2", true, false, "qf", 2, 2},
				{"Quarterfinal 3-2", "QF3-2", "", 31, "QF3", true, false, "qf", 3, 2},
				{"Quarterfinal 4-2", "QF4-2", "", 32, "QF4", true, false, "qf", 4, 2},
				{"Quarterfinal 1-3", "QF1-3", "", 33, "QF1", true, false, "qf", 1, 3},
				{"Quarterfinal 2-3", "QF2-3", "", 34, "QF2", true, false, "qf", 2, 3},
				{"Quarterfinal 3-3", "QF3-3", "", 35, "QF3", true, false, "qf", 3, 3},
				{"Quarterfinal 4-3", "QF4-3", "", 36, "QF4", true, false, "qf", 4, 3},
			},
		)
	}
	assertFullSemifinalsOnward(t, matchSpecs, startingIndex+12)
}

func assertFullSemifinalsOnward(t *testing.T, matchSpecs []*matchSpec, startingIndex int) {
	if assert.Equal(t, startingIndex+12, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[startingIndex:startingIndex+6],
			[]expectedMatchSpec{
				{"Semifinal 1-1", "SF1-1", "", 37, "SF1", true, false, "sf", 1, 1},
				{"Semifinal 2-1", "SF2-1", "", 38, "SF2", true, false, "sf", 2, 1},
				{"Semifinal 1-2", "SF1-2", "", 39, "SF1", true, false, "sf", 1, 2},
				{"Semifinal 2-2", "SF2-2", "", 40, "SF2", true, false, "sf", 2, 2},
				{"Semifinal 1-3", "SF1-3", "", 41, "SF1", true, false, "sf", 1, 3},
				{"Semifinal 2-3", "SF2-3", "", 42, "SF2", true, false, "sf", 2, 3},
			},
		)
	}
	assertFullFinals(t, matchSpecs, startingIndex+6)
}

func assertFullFinals(t *testing.T, matchSpecs []*matchSpec, startingIndex int) {
	if assert.Equal(t, startingIndex+6, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[startingIndex:startingIndex+6],
			[]expectedMatchSpec{
				{"Final 1", "F1", "", 43, "F", false, false, "f", 1, 1},
				{"Final 2", "F2", "", 44, "F", false, false, "f", 1, 2},
				{"Final 3", "F3", "", 45, "F", false, false, "f", 1, 3},
				{"Overtime 1", "O1", "", 46, "F", true, true, "f", 1, 4},
				{"Overtime 2", "O2", "", 47, "F", true, true, "f", 1, 5},
				{"Overtime 3", "O3", "", 48, "F", true, true, "f", 1, 6},
			},
		)
	}
}
