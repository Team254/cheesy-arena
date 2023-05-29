// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentAlliance(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	alliance, err := db.GetAllianceById(1114)
	assert.Nil(t, err)
	assert.Nil(t, alliance)
}

func TestAllianceCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	alliance := Alliance{Id: 3, TeamIds: []int{254, 1114, 296, 1503}, Lineup: [3]int{1114, 254, 296}}
	assert.Nil(t, db.CreateAlliance(&alliance))
	alliance2, err := db.GetAllianceById(3)
	assert.Nil(t, err)
	assert.Equal(t, alliance, *alliance2)

	alliance.TeamIds = append(alliance.TeamIds, 296)
	assert.Nil(t, db.UpdateAlliance(&alliance))
	alliance2, err = db.GetAllianceById(3)
	assert.Nil(t, err)
	assert.Equal(t, alliance, *alliance2)

	assert.Nil(t, db.DeleteAlliance(alliance.Id))
	alliance2, err = db.GetAllianceById(3)
	assert.Nil(t, err)
	assert.Nil(t, alliance2)
}

func TestUpdateAllianceFromMatch(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	alliance := Alliance{Id: 3, TeamIds: []int{254, 1114, 296, 1503}, Lineup: [3]int{1114, 254, 296}}
	assert.Nil(t, db.CreateAlliance(&alliance))
	assert.Nil(t, db.UpdateAllianceFromMatch(3, [3]int{1503, 188, 296}))
	alliance2, err := db.GetAllianceById(3)
	assert.Nil(t, err)
	assert.Equal(t, []int{254, 1114, 296, 1503, 188}, alliance2.TeamIds)
	assert.Equal(t, [3]int{1503, 188, 296}, alliance2.Lineup)
}

func TestTruncateAllianceTeams(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	alliance := Alliance{Id: 1, TeamIds: []int{148, 118, 125}, Lineup: [3]int{118, 148, 125}}
	assert.Nil(t, db.CreateAlliance(&alliance))
	assert.Nil(t, db.TruncateAlliances())
	alliance2, err := db.GetAllianceById(1)
	assert.Nil(t, err)
	assert.Nil(t, alliance2)
}

func TestGetAllAlliances(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	alliances, err := db.GetAllAlliances()
	assert.Nil(t, err)
	assert.Empty(t, alliances)

	BuildTestAlliances(db)
	alliances, err = db.GetAllAlliances()
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(alliances)) {
		assert.Equal(t, 1, alliances[0].Id)
		assert.Equal(t, []int{254, 469, 2848, 74, 3175}, alliances[0].TeamIds)
		assert.Equal(t, 2, alliances[1].Id)
		assert.Equal(t, []int{1718, 2451, 1619}, alliances[1].TeamIds)
	}
}

func TestGetOffFieldTeamIds(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()
	BuildTestAlliances(db)

	match := &Match{
		PlayoffRedAlliance:  1,
		PlayoffBlueAlliance: 2,
		Red1:                469,
		Red2:                254,
		Red3:                2848,
		Blue1:               1619,
		Blue2:               1718,
		Blue3:               2451,
	}

	redOffFieldTeams, blueOffFieldTeams, err := db.GetOffFieldTeamIds(match)
	assert.Nil(t, err)
	assert.Equal(t, []int{74, 3175}, redOffFieldTeams)
	assert.Equal(t, []int{}, blueOffFieldTeams)

	match.Red1 = 74
	match.Red2 = 3175
	redOffFieldTeams, blueOffFieldTeams, err = db.GetOffFieldTeamIds(match)
	assert.Nil(t, err)
	assert.Equal(t, []int{254, 469}, redOffFieldTeams)
	assert.Equal(t, []int{}, blueOffFieldTeams)

	match.PlayoffRedAlliance = 0
	match.PlayoffBlueAlliance = 0
	redOffFieldTeams, blueOffFieldTeams, err = db.GetOffFieldTeamIds(match)
	assert.Nil(t, err)
	assert.Equal(t, []int{}, redOffFieldTeams)
	assert.Equal(t, []int{}, blueOffFieldTeams)

	match = &Match{
		PlayoffRedAlliance:  2,
		PlayoffBlueAlliance: 1,
		Red1:                1718,
		Red2:                2451,
		Red3:                1619,
		Blue1:               3175,
		Blue2:               74,
		Blue3:               2848,
	}
	redOffFieldTeams, blueOffFieldTeams, err = db.GetOffFieldTeamIds(match)
	assert.Nil(t, err)
	assert.Equal(t, []int{}, redOffFieldTeams)
	assert.Equal(t, []int{254, 469}, blueOffFieldTeams)
}
