// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAssignTeam(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	team := Team{Id: 254}
	err = db.CreateTeam(&team)
	assert.Nil(t, err)
	err = db.CreateTeam(&Team{Id: 1114})
	assert.Nil(t, err)
	mainArena.Setup()

	err = mainArena.AssignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *mainArena.allianceStations["B1"].team)
	dsConn := mainArena.allianceStations["B1"].driverStationConnection
	assert.Equal(t, 254, dsConn.TeamId)
	assert.Equal(t, "B1", dsConn.AllianceStation)

	// Nothing should happen if the same team is assigned to the same station.
	err = mainArena.AssignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *mainArena.allianceStations["B1"].team)
	dsConn2 := mainArena.allianceStations["B1"].driverStationConnection
	assert.Equal(t, dsConn, dsConn2) // Pointer equality

	// Test reassignment to another team.
	err = mainArena.AssignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.NotEqual(t, team, *mainArena.allianceStations["B1"].team)
	assert.Equal(t, 1114, mainArena.allianceStations["B1"].driverStationConnection.TeamId)
	err = dsConn.conn.Close()
	assert.NotNil(t, err) // Connection should have already been closed.

	// Check assigning an unknown team.
	err = mainArena.AssignTeam(1503, "R1")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid team number")
	}

	// Check assigning zero as the team number.
	err = mainArena.AssignTeam(0, "R2")
	assert.Nil(t, err)
	assert.Nil(t, mainArena.allianceStations["R2"].team)
	assert.Nil(t, mainArena.allianceStations["R2"].driverStationConnection)

	// Check assigning to a non-existent station.
	err = mainArena.AssignTeam(254, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}
}

func TestArenaMatchFlow(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	err = db.CreateTeam(&Team{Id: 254})
	assert.Nil(t, err)
	mainArena.Setup()
	err = mainArena.AssignTeam(254, "B3")
	assert.Nil(t, err)

	// Check pre-match state and packet timing.
	assert.Equal(t, PRE_MATCH, mainArena.matchState)
	mainArena.Update()
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	lastPacketCount := mainArena.allianceStations["B3"].driverStationConnection.packetCount
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-10 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, lastPacketCount, mainArena.allianceStations["B3"].driverStationConnection.packetCount)
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, lastPacketCount+1, mainArena.allianceStations["B3"].driverStationConnection.packetCount)

	// Check match start, autonomous and transition to teleop.
	mainArena.StartMatch()
	mainArena.Update()
	assert.Equal(t, AUTO_PERIOD, mainArena.matchState)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.Update()
	assert.Equal(t, AUTO_PERIOD, mainArena.matchState)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.matchStartTime = time.Now().Add(-autoDurationSec * time.Second)
	mainArena.Update()
	assert.Equal(t, PAUSE_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.Update()
	assert.Equal(t, PAUSE_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.matchStartTime = time.Now().Add(-(autoDurationSec + pauseDurationSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)

	// Check e-stop and bypass.
	mainArena.allianceStations["B3"].emergencyStop = true
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.allianceStations["B3"].bypass = true
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.allianceStations["B3"].emergencyStop = false
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.allianceStations["B3"].bypass = false
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)

	// Check endgame and match end.
	mainArena.matchStartTime = time.Now().
		Add(-(autoDurationSec + pauseDurationSec + teleopDurationSec - endgameTimeLeftSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, ENDGAME_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.Update()
	assert.Equal(t, ENDGAME_PERIOD, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.matchStartTime = time.Now().Add(-(autoDurationSec + pauseDurationSec + teleopDurationSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, POST_MATCH, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
	mainArena.Update()
	assert.Equal(t, POST_MATCH, mainArena.matchState)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)

	mainArena.ResetMatch()
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, PRE_MATCH, mainArena.matchState)
	assert.Equal(t, true, mainArena.allianceStations["B3"].driverStationConnection.Auto)
	assert.Equal(t, false, mainArena.allianceStations["B3"].driverStationConnection.Enabled)
}

func TestArenaStateEnforcement(t *testing.T) {
	mainArena.Setup()

	err := mainArena.LoadMatch(new(Match))
	assert.Nil(t, err)
	err = mainArena.StartMatch()
	assert.Nil(t, err)
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	mainArena.matchState = AUTO_PERIOD
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	mainArena.matchState = PAUSE_PERIOD
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	mainArena.matchState = TELEOP_PERIOD
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	mainArena.matchState = ENDGAME_PERIOD
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	mainArena.matchState = POST_MATCH
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}

	err = mainArena.ResetMatch()
	assert.Nil(t, err)
	assert.Equal(t, PRE_MATCH, mainArena.matchState)
	assert.Nil(t, mainArena.currentMatch)
	err = mainArena.ResetMatch()
	assert.Nil(t, err)
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "no match is loaded")
	}
	err = mainArena.LoadMatch(new(Match))
	assert.Nil(t, err)
}
