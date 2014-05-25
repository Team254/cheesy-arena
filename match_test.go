// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"testing"
	"time"
)

func TestGetNonexistentMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	match, err := db.GetMatchById(1114)
	if err != nil {
		t.Error("Error:", err)
	}
	if match != nil {
		t.Errorf("Expected '%v' to be nil", match)
	}
}

func TestMatchCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	match := Match{254, "qualification", "254", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	match2, err := db.GetMatchById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if match != *match2 {
		t.Errorf("Expected '%v', got '%v'", match, match2)
	}

	match.Status = "started"
	db.SaveMatch(&match)
	match2, err = db.GetMatchById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if match.Status != match2.Status {
		t.Errorf("Expected '%v', got '%v'", match.Status, match2.Status)
	}

	db.DeleteMatch(&match)
	match2, err = db.GetMatchById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if match2 != nil {
		t.Errorf("Expected '%v' to be nil", match2)
	}
}

func TestTruncateMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	if err != nil {
		t.Error("Error:", err)
	}
	defer db.Close()

	match := Match{254, "qualification", "254", time.Now().UTC(), 1, false, 2, false, 3, false, 4, false, 5,
		false, 6, false, "", time.Now().UTC()}
	db.CreateMatch(&match)
	db.TruncateMatches()
	match2, err := db.GetMatchById(254)
	if err != nil {
		t.Error("Error:", err)
	}
	if match2 != nil {
		t.Errorf("Expected '%v' to be nil", match2)
	}
}
