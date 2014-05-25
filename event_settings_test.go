// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"testing"
)

func TestEventSettingsReadWrite(t *testing.T) {
	clearDb()
	defer clearDb()

	db, _ := OpenDatabase(testDbPath)
	defer db.Close()
	eventSettings, err := db.GetEventSettings()
	if err != nil {
		t.Error("Error:", err)
	}
	if *eventSettings != *new(EventSettings) {
		t.Errorf("Expected blank event settings, got %v", eventSettings)
	}

	eventSettings.Name = "Chezy Champs"
	eventSettings.Code = "cc"
	err = db.SaveEventSettings(eventSettings)
	if err != nil {
		t.Error("Error:", err)
	}
	eventSettings2, err := db.GetEventSettings()
	if err != nil {
		t.Error("Error:", err)
	}
	if *eventSettings2 != *eventSettings {
		t.Errorf("Expected '%v', got '%v'", eventSettings, eventSettings2)
	}
}
