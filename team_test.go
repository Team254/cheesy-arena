// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentTeam(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	team, err := db.GetTeamById(1114)
	assert.Nil(t, err)
	assert.Nil(t, team)
}

func TestTeamCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	team2, err := db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Equal(t, team, *team2)

	team.Name = "Updated name"
	db.SaveTeam(&team)
	team2, err = db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Equal(t, team.Name, team2.Name)

	db.DeleteTeam(&team)
	team2, err = db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Nil(t, team2)
}

func TestTruncateTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	db.TruncateTeams()
	team2, err := db.GetTeamById(254)
	assert.Nil(t, err)
	assert.Nil(t, team2)
}

func TestGetAllTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	teams, err := db.GetAllTeams()
	assert.Nil(t, err)
	assert.Empty(t, teams)

	numTeams := 20
	for i := 1; i <= numTeams; i++ {
		db.CreateTeam(&Team{i, "", "", "", "", "", 2014, ""})
	}
	teams, err = db.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, numTeams, len(teams))
	for i := 0; i < numTeams; i++ {
		assert.Equal(t, i+1, teams[i].Id)
	}
}
