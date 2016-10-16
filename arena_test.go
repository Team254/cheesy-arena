// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
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
	eventSettings, _ = db.GetEventSettings()
	team := Team{Id: 254}
	err = db.CreateTeam(&team)
	assert.Nil(t, err)
	err = db.CreateTeam(&Team{Id: 1114})
	assert.Nil(t, err)
	mainArena.Setup()

	err = mainArena.AssignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *mainArena.AllianceStations["B1"].Team)
	dummyDs := &DriverStationConnection{TeamId: 254}
	mainArena.AllianceStations["B1"].DsConn = dummyDs

	// Nothing should happen if the same team is assigned to the same station.
	err = mainArena.AssignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *mainArena.AllianceStations["B1"].Team)
	assert.NotNil(t, mainArena.AllianceStations["B1"])
	assert.Equal(t, dummyDs, mainArena.AllianceStations["B1"].DsConn) // Pointer equality

	// Test reassignment to another team.
	err = mainArena.AssignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.NotEqual(t, team, *mainArena.AllianceStations["B1"].Team)
	assert.Nil(t, mainArena.AllianceStations["B1"].DsConn)

	// Check assigning zero as the team number.
	err = mainArena.AssignTeam(0, "R2")
	assert.Nil(t, err)
	assert.Nil(t, mainArena.AllianceStations["R2"].Team)
	assert.Nil(t, mainArena.AllianceStations["R2"].DsConn)

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
	eventSettings, _ = db.GetEventSettings()
	mainArena = Arena{}
	mainArena.Setup()
	db.CreateTeam(&Team{Id: 254})
	err = mainArena.AssignTeam(254, "B3")
	dummyDs := &DriverStationConnection{TeamId: 254}
	mainArena.AllianceStations["B3"].DsConn = dummyDs
	assert.Nil(t, err)

	// Check pre-match state and packet timing.
	assert.Equal(t, PRE_MATCH, mainArena.MatchState)
	mainArena.Update()
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	lastPacketCount := mainArena.AllianceStations["B3"].DsConn.packetCount
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-10 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, lastPacketCount, mainArena.AllianceStations["B3"].DsConn.packetCount)
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, lastPacketCount+1, mainArena.AllianceStations["B3"].DsConn.packetCount)

	// Check match start, autonomous and transition to teleop.
	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].DsConn.RobotLinked = true
	err = mainArena.StartMatch()
	assert.Nil(t, err)
	mainArena.Update()
	assert.Equal(t, AUTO_PERIOD, mainArena.MatchState)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.Update()
	assert.Equal(t, AUTO_PERIOD, mainArena.MatchState)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.matchStartTime = time.Now().Add(-time.Duration(mainArena.matchTiming.AutoDurationSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, PAUSE_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.Update()
	assert.Equal(t, PAUSE_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.matchStartTime = time.Now().Add(-time.Duration(mainArena.matchTiming.AutoDurationSec+
		mainArena.matchTiming.PauseDurationSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)

	// Check e-stop and bypass.
	mainArena.AllianceStations["B3"].EmergencyStop = true
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.AllianceStations["B3"].Bypass = true
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.AllianceStations["B3"].EmergencyStop = false
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.AllianceStations["B3"].Bypass = false
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, TELEOP_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)

	// Check endgame and match end.
	mainArena.matchStartTime = time.Now().
		Add(-time.Duration(mainArena.matchTiming.AutoDurationSec+mainArena.matchTiming.PauseDurationSec+
			mainArena.matchTiming.TeleopDurationSec-mainArena.matchTiming.EndgameTimeLeftSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, ENDGAME_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.Update()
	assert.Equal(t, ENDGAME_PERIOD, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.matchStartTime = time.Now().Add(-time.Duration(mainArena.matchTiming.AutoDurationSec+
		mainArena.matchTiming.PauseDurationSec+mainArena.matchTiming.TeleopDurationSec) * time.Second)
	mainArena.Update()
	assert.Equal(t, POST_MATCH, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	mainArena.Update()
	assert.Equal(t, POST_MATCH, mainArena.MatchState)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)

	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.ResetMatch()
	mainArena.lastDsPacketTime = mainArena.lastDsPacketTime.Add(-300 * time.Millisecond)
	mainArena.Update()
	assert.Equal(t, PRE_MATCH, mainArena.MatchState)
	assert.Equal(t, true, mainArena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, mainArena.AllianceStations["B3"].DsConn.Enabled)
	assert.Equal(t, false, mainArena.AllianceStations["R1"].Bypass)
}

func TestArenaStateEnforcement(t *testing.T) {
	mainArena.Setup()
	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].Bypass = true

	err := mainArena.LoadMatch(new(Match))
	assert.Nil(t, err)
	err = mainArena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot abort match when")
	}
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
	mainArena.MatchState = AUTO_PERIOD
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
	mainArena.MatchState = PAUSE_PERIOD
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
	mainArena.MatchState = TELEOP_PERIOD
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
	mainArena.MatchState = ENDGAME_PERIOD
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
	err = mainArena.AbortMatch()
	assert.Nil(t, err)
	mainArena.MatchState = POST_MATCH
	err = mainArena.LoadMatch(new(Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = mainArena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot abort match when")
	}

	err = mainArena.ResetMatch()
	assert.Nil(t, err)
	assert.Equal(t, PRE_MATCH, mainArena.MatchState)
	err = mainArena.ResetMatch()
	assert.Nil(t, err)
	err = mainArena.LoadMatch(new(Match))
	assert.Nil(t, err)
}

func TestMatchStartRobotLinkEnforcement(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	db.CreateTeam(&Team{Id: 101})
	db.CreateTeam(&Team{Id: 102})
	db.CreateTeam(&Team{Id: 103})
	db.CreateTeam(&Team{Id: 104})
	db.CreateTeam(&Team{Id: 105})
	db.CreateTeam(&Team{Id: 106})
	match := Match{Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	mainArena.Setup()

	err = mainArena.LoadMatch(&match)
	assert.Nil(t, err)
	mainArena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 101}
	mainArena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 102}
	mainArena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 103}
	mainArena.AllianceStations["B1"].DsConn = &DriverStationConnection{TeamId: 104}
	mainArena.AllianceStations["B2"].DsConn = &DriverStationConnection{TeamId: 105}
	mainArena.AllianceStations["B3"].DsConn = &DriverStationConnection{TeamId: 106}
	for _, station := range mainArena.AllianceStations {
		station.DsConn.RobotLinked = true
	}
	err = mainArena.StartMatch()
	assert.Nil(t, err)
	mainArena.MatchState = PRE_MATCH

	// Check with a single team e-stopped, not linked and bypassed.
	mainArena.AllianceStations["R1"].EmergencyStop = true
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	mainArena.AllianceStations["R1"].EmergencyStop = false
	mainArena.AllianceStations["R1"].DsConn.RobotLinked = false
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	mainArena.AllianceStations["R1"].Bypass = true
	err = mainArena.StartMatch()
	assert.Nil(t, err)
	mainArena.AllianceStations["R1"].Bypass = false
	mainArena.MatchState = PRE_MATCH

	// Check with a team missing.
	err = mainArena.AssignTeam(0, "R1")
	assert.Nil(t, err)
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	mainArena.AllianceStations["R1"].Bypass = true
	err = mainArena.StartMatch()
	assert.Nil(t, err)
	mainArena.MatchState = PRE_MATCH

	// Check with no teams present.
	mainArena.LoadMatch(new(Match))
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	mainArena.AllianceStations["R1"].Bypass = true
	mainArena.AllianceStations["R2"].Bypass = true
	mainArena.AllianceStations["R3"].Bypass = true
	mainArena.AllianceStations["B1"].Bypass = true
	mainArena.AllianceStations["B2"].Bypass = true
	mainArena.AllianceStations["B3"].Bypass = true
	mainArena.AllianceStations["B3"].EmergencyStop = true
	err = mainArena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	mainArena.AllianceStations["B3"].EmergencyStop = false
	err = mainArena.StartMatch()
	assert.Nil(t, err)
}

func TestLoadNextMatch(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	mainArena.Setup()

	db.CreateTeam(&Team{Id: 1114})
	practiceMatch1 := Match{Type: "practice", DisplayName: "1"}
	practiceMatch2 := Match{Type: "practice", DisplayName: "2", Status: "complete"}
	practiceMatch3 := Match{Type: "practice", DisplayName: "3"}
	db.CreateMatch(&practiceMatch1)
	db.CreateMatch(&practiceMatch2)
	db.CreateMatch(&practiceMatch3)
	qualificationMatch1 := Match{Type: "qualification", DisplayName: "1", Status: "complete"}
	qualificationMatch2 := Match{Type: "qualification", DisplayName: "2"}
	db.CreateMatch(&qualificationMatch1)
	db.CreateMatch(&qualificationMatch2)

	// Test match should be followed by another, empty test match.
	assert.Equal(t, 0, mainArena.currentMatch.Id)
	err = mainArena.SubstituteTeam(1114, "R1")
	assert.Nil(t, err)
	mainArena.currentMatch.Status = "complete"
	err = mainArena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, 0, mainArena.currentMatch.Id)
	assert.Equal(t, 0, mainArena.currentMatch.Red1)
	assert.NotEqual(t, "complete", mainArena.currentMatch.Status)

	// Other matches should be loaded by type until they're all complete.
	err = mainArena.LoadMatch(&practiceMatch2)
	assert.Nil(t, err)
	err = mainArena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch1.Id, mainArena.currentMatch.Id)
	practiceMatch1.Status = "complete"
	db.SaveMatch(&practiceMatch1)
	err = mainArena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch3.Id, mainArena.currentMatch.Id)
	practiceMatch3.Status = "complete"
	db.SaveMatch(&practiceMatch3)
	err = mainArena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch3.Id, mainArena.currentMatch.Id)
	assert.Equal(t, "complete", practiceMatch3.Status)

	err = mainArena.LoadMatch(&qualificationMatch1)
	assert.Nil(t, err)
	err = mainArena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, qualificationMatch2.Id, mainArena.currentMatch.Id)
}

func TestSubstituteTeam(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()
	db.CreateTeam(&Team{Id: 101})
	db.CreateTeam(&Team{Id: 102})
	db.CreateTeam(&Team{Id: 103})
	db.CreateTeam(&Team{Id: 104})
	db.CreateTeam(&Team{Id: 105})
	db.CreateTeam(&Team{Id: 106})
	db.CreateTeam(&Team{Id: 107})

	// Substitute teams into test match.
	err = mainArena.SubstituteTeam(101, "B1")
	assert.Nil(t, err)
	assert.Equal(t, 101, mainArena.currentMatch.Blue1)
	assert.Equal(t, 101, mainArena.AllianceStations["B1"].Team.Id)
	err = mainArena.AssignTeam(104, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}

	// Substitute teams into practice match. Replacement should also be persisted in the DB.
	match := Match{Type: "practice", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	mainArena.LoadMatch(&match)
	err = mainArena.SubstituteTeam(107, "R1")
	assert.Nil(t, err)
	assert.Equal(t, 107, mainArena.currentMatch.Red1)
	assert.Equal(t, 107, mainArena.AllianceStations["R1"].Team.Id)
	CommitMatchScore(mainArena.currentMatch, &MatchResult{MatchId: mainArena.currentMatch.Id}, false)
	match2, _ := db.GetMatchById(match.Id)
	assert.Equal(t, 107, match2.Red1)

	// Check that substitution is disallowed in qualification matches.
	match = Match{Type: "qualification", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	mainArena.LoadMatch(&match)
	err = mainArena.SubstituteTeam(107, "R1")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Can't substitute teams for qualification matches.")
	}
	match = Match{Type: "elimination", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	db.CreateMatch(&match)
	mainArena.LoadMatch(&match)
	assert.Nil(t, mainArena.SubstituteTeam(107, "R1"))
}

func TestSetupNetwork(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	// Verify the setup ran by checking the log for the expected failure messages.
	eventSettings.NetworkSecurityEnabled = true
	aironetTelnetPort = 10023
	catalystTelnetPort = 10023
	mainArena.LoadMatch(&Match{Type: "test"})
	var writer bytes.Buffer
	log.SetOutput(&writer)
	time.Sleep(time.Millisecond * 10) // Allow some time for the asynchronous configuration to happen.
	assert.Contains(t, writer.String(), "Failed to configure team Ethernet")
	assert.Contains(t, writer.String(), "Failed to configure team WiFi")
}
