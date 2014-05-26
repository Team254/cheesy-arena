// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"testing"
)

func TestGetNonexistentTeam(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	team, err := db.GetTeamById(1114)
	if err != nil {
		t.Error("Error:", err)
	}
	if team != nil {
		t.Errorf("Expected '%v' to be nil", team)
	}
}

func TestTeamCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	team2, err := db.GetTeamById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if team != *team2 {
		t.Errorf("Expected '%v', got '%v'", team, team2)
	}

	team.Name = "Updated name"
	db.SaveTeam(&team)
	team2, err = db.GetTeamById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if team.Name != team2.Name {
		t.Errorf("Expected '%v', got '%v'", team.Name, team2.Name)
	}

	db.DeleteTeam(&team)
	team2, err = db.GetTeamById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if team2 != nil {
		t.Errorf("Expected '%v' to be nil", team2)
	}
}

func TestTruncateTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	team := Team{254, "NASA Ames Research Center", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)
	db.TruncateTeams()
	team2, err := db.GetTeamById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if team2 != nil {
		t.Errorf("Expected '%v' to be nil", team2)
	}
}

func TestGetAllTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	teams, err := db.GetAllTeams()
	if err != nil {
		t.Error("Error:", err)
	}
	if len(teams) != 0 {
		t.Errorf("Expected %d teams, got %d", 0, len(teams))
	}

	numTeams := 20
	for i := 1; i <= numTeams; i++ {
		db.CreateTeam(&Team{i, "", "", "", "", "", 2014, ""})
	}
	teams, err = db.GetAllTeams()
	if err != nil {
		t.Error("Error:", err)
	}
	if len(teams) != numTeams {
		t.Errorf("Expected %d teams, got %d", numTeams, len(teams))
	}
	for i := 0; i < numTeams; i++ {
		if teams[i].Id != i+1 {
			t.Errorf("Expected team %d, got %d", i+1, teams[i].Id)
		}
	}
}
