// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"testing"
)

func TestGetNonexistentTeam(t *testing.T) {
	clearDb()
	defer clearDb()

	db, _ := OpenDatabase(testDbPath)
	defer db.Close()
	team, _ := db.GetTeamById(1114)
	if team != nil {
		t.Errorf("Expected '%v' to be nil", team)
	}
}

func TestTeamCrud(t *testing.T) {
	clearDb()
	defer clearDb()

	db, _ := OpenDatabase(testDbPath)
	defer db.Close()
	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	team2, _ := db.GetTeamById(254)
	if team != *team2 {
		t.Errorf("Expected '%v', got '%v'", team, team2)
	}

	team.Name = "Updated name"
	db.SaveTeam(&team)
	team2, _ = db.GetTeamById(254)
	if team.Name != team2.Name {
		t.Errorf("Expected '%v', got '%v'", team.Name, team2.Name)
	}

	db.DeleteTeam(&team)
	team2, _ = db.GetTeamById(254)
	if team2 != nil {
		t.Errorf("Expected '%v' to be nil", team2)
	}
}

func TestTruncateTeams(t *testing.T) {
	clearDb()
	defer clearDb()

	db, _ := OpenDatabase(testDbPath)
	defer db.Close()
	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	db.TruncateTeams()
	team2, _ := db.GetTeamById(254)
	if team2 != nil {
		t.Errorf("Expected '%v' to be nil", team2)
	}
}
