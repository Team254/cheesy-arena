// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentTeam(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	team, err := db.GetTeamById(1114)
	assert.Nil(t, err)
	assert.Nil(t, team)
}

func TestTeamCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	team := Team{Id: 254, Name: "NASA", Nickname: "The Cheesy Poofs", City: "San Jose", StateProv: "CA",
		Country: "USA", RookieYear: 1999, RobotName: "Barrage"}
	db.CreateTeam(&team)
	team2, err := db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Equal(t, team, *team2)

	team.Name = "Updated name"
	db.UpdateTeam(&team)
	team2, err = db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Equal(t, team.Name, team2.Name)

	db.DeleteTeam(team.Id)
	team2, err = db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Nil(t, team2)
}

func TestTruncateTeams(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	team := Team{Id: 254, Name: "NASA", Nickname: "The Cheesy Poofs", City: "San Jose", StateProv: "CA",
		Country: "USA", RookieYear: 1999, RobotName: "Barrage"}
	db.CreateTeam(&team)
	db.TruncateTeams()
	team2, err := db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Nil(t, team2)
}

func TestGetAllTeams(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	teams, err := db.GetAllTeams()
	assert.Nil(t, err)
	assert.Empty(t, teams)

	numTeams := 20
	for i := 1; i <= numTeams; i++ {
		db.CreateTeam(&Team{Id: i, RookieYear: 2014})
	}
	teams, err = db.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, numTeams, len(teams))
	for i := 0; i < numTeams; i++ {
		assert.Equal(t, i+1, teams[i].Id)
	}
}
