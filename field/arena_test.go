// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAssignTeam(t *testing.T) {
	arena := setupTestArena(t)

	team := model.Team{Id: 254}
	err := arena.Database.CreateTeam(&team)
	assert.Nil(t, err)
	err = arena.Database.CreateTeam(&model.Team{Id: 1114})
	assert.Nil(t, err)

	err = arena.assignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *arena.AllianceStations["B1"].Team)
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["B1"].DsConn = dummyDs

	// Nothing should happen if the same team is assigned to the same station.
	err = arena.assignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *arena.AllianceStations["B1"].Team)
	assert.NotNil(t, arena.AllianceStations["B1"])
	assert.Equal(t, dummyDs, arena.AllianceStations["B1"].DsConn) // Pointer equality

	// Test reassignment to another team.
	err = arena.assignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.NotEqual(t, team, *arena.AllianceStations["B1"].Team)
	assert.Nil(t, arena.AllianceStations["B1"].DsConn)

	// Check assigning zero as the team number.
	err = arena.assignTeam(0, "R2")
	assert.Nil(t, err)
	assert.Nil(t, arena.AllianceStations["R2"].Team)
	assert.Nil(t, arena.AllianceStations["R2"].DsConn)

	// Check assigning to a non-existent station.
	err = arena.assignTeam(254, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}
}

func TestArenaCheckCanStartMatch(t *testing.T) {
	arena := setupTestArena(t)

	// Check robot state constraints.
	err := arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match until all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match until all robots are connected or bypassed")
	}
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.checkCanStartMatch())

	// Check PLC constraints.
	arena.Plc.SetAddress("1.2.3.4")
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while PLC is not healthy")
	}
	arena.Plc.SetAddress("")
	assert.Nil(t, arena.checkCanStartMatch())
}

func TestArenaMatchFlow(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 254})
	err := arena.assignTeam(254, "B3")
	assert.Nil(t, err)
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["B3"].DsConn = dummyDs

	// Check pre-match state and packet timing.
	assert.Equal(t, PreMatch, arena.MatchState)
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	lastPacketCount := arena.AllianceStations["B3"].DsConn.packetCount
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-10 * time.Millisecond)
	arena.Update()
	assert.Equal(t, lastPacketCount, arena.AllianceStations["B3"].DsConn.packetCount)
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, lastPacketCount+1, arena.AllianceStations["B3"].DsConn.packetCount)

	// Check match start, autonomous and transition to teleop.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].DsConn.RobotLinked = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.Update()
	assert.Equal(t, WarmupPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, WarmupPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)

	// Check e-stop and bypass.
	arena.AllianceStations["B3"].Estop = true
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].Bypass = true
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].Estop = false
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].Bypass = false
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)

	// Check match end.
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec) *
		time.Second)
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)

	arena.AllianceStations["R1"].Bypass = true
	arena.ResetMatch()
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].Bypass)
}

func TestArenaStateEnforcement(t *testing.T) {
	arena := setupTestArena(t)

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true

	err := arena.LoadMatch(new(model.Match))
	assert.Nil(t, err)
	err = arena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot abort match when")
	}
	err = arena.StartMatch()
	assert.Nil(t, err)
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	arena.MatchState = AutoPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	arena.MatchState = PausePeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	arena.MatchState = TeleopPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot reset match while")
	}
	arena.MatchState = PostMatch
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot start match while")
	}
	err = arena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Cannot abort match when")
	}

	err = arena.ResetMatch()
	assert.Nil(t, err)
	assert.Equal(t, PreMatch, arena.MatchState)
	err = arena.ResetMatch()
	assert.Nil(t, err)
	err = arena.LoadMatch(new(model.Match))
	assert.Nil(t, err)
}

func TestMatchStartRobotLinkEnforcement(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106})
	match := model.Match{Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)

	err := arena.LoadMatch(&match)
	assert.Nil(t, err)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 101}
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 102}
	arena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 103}
	arena.AllianceStations["B1"].DsConn = &DriverStationConnection{TeamId: 104}
	arena.AllianceStations["B2"].DsConn = &DriverStationConnection{TeamId: 105}
	arena.AllianceStations["B3"].DsConn = &DriverStationConnection{TeamId: 106}
	for _, station := range arena.AllianceStations {
		station.DsConn.RobotLinked = true
	}
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.MatchState = PreMatch

	// Check with a single team e-stopped, not linked and bypassed.
	arena.AllianceStations["R1"].Estop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	arena.AllianceStations["R1"].Estop = false
	arena.AllianceStations["R1"].DsConn.RobotLinked = false
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.AllianceStations["R1"].Bypass = false
	arena.MatchState = PreMatch

	// Check with a team missing.
	err = arena.assignTeam(0, "R1")
	assert.Nil(t, err)
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.MatchState = PreMatch

	// Check with no teams present.
	arena.LoadMatch(new(model.Match))
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "until all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	arena.AllianceStations["B3"].Estop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	arena.AllianceStations["B3"].Estop = false
	err = arena.StartMatch()
	assert.Nil(t, err)
}

func TestLoadNextMatch(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 1114})
	practiceMatch1 := model.Match{Type: "practice", DisplayName: "1"}
	practiceMatch2 := model.Match{Type: "practice", DisplayName: "2", Status: model.RedWonMatch}
	practiceMatch3 := model.Match{Type: "practice", DisplayName: "3"}
	arena.Database.CreateMatch(&practiceMatch1)
	arena.Database.CreateMatch(&practiceMatch2)
	arena.Database.CreateMatch(&practiceMatch3)
	qualificationMatch1 := model.Match{Type: "qualification", DisplayName: "1", Status: model.BlueWonMatch}
	qualificationMatch2 := model.Match{Type: "qualification", DisplayName: "2"}
	arena.Database.CreateMatch(&qualificationMatch1)
	arena.Database.CreateMatch(&qualificationMatch2)

	// Test match should be followed by another, empty test match.
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	err := arena.SubstituteTeam(1114, "R1")
	assert.Nil(t, err)
	arena.CurrentMatch.Status = model.TieMatch
	err = arena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, 0, arena.CurrentMatch.Red1)
	assert.Equal(t, false, arena.CurrentMatch.IsComplete())

	// Other matches should be loaded by type until they're all complete.
	err = arena.LoadMatch(&practiceMatch2)
	assert.Nil(t, err)
	err = arena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch1.Id, arena.CurrentMatch.Id)
	practiceMatch1.Status = model.RedWonMatch
	arena.Database.UpdateMatch(&practiceMatch1)
	err = arena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch3.Id, arena.CurrentMatch.Id)
	practiceMatch3.Status = model.BlueWonMatch
	arena.Database.UpdateMatch(&practiceMatch3)
	err = arena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, "test", arena.CurrentMatch.Type)

	err = arena.LoadMatch(&qualificationMatch1)
	assert.Nil(t, err)
	err = arena.LoadNextMatch()
	assert.Nil(t, err)
	assert.Equal(t, qualificationMatch2.Id, arena.CurrentMatch.Id)
}

func TestSubstituteTeam(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106})
	arena.Database.CreateTeam(&model.Team{Id: 107})

	// Substitute teams into test match.
	err := arena.SubstituteTeam(101, "B1")
	assert.Nil(t, err)
	assert.Equal(t, 101, arena.CurrentMatch.Blue1)
	assert.Equal(t, 101, arena.AllianceStations["B1"].Team.Id)
	err = arena.assignTeam(104, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}

	// Substitute teams into practice match.
	match := model.Match{Type: "practice", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeam(107, "R1")
	assert.Nil(t, err)
	assert.Equal(t, 107, arena.CurrentMatch.Red1)
	assert.Equal(t, 107, arena.AllianceStations["R1"].Team.Id)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = arena.CurrentMatch.Id

	// Check that substitution is disallowed in qualification matches.
	match = model.Match{Type: "qualification", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeam(107, "R1")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Can't substitute teams for qualification matches.")
	}
	match = model.Match{Type: "elimination", Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	assert.Nil(t, arena.SubstituteTeam(107, "R1"))
}

func TestAstop(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 254})
	err := arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["R1"].DsConn = dummyDs
	arena.Database.CreateTeam(&model.Team{Id: 148})
	err = arena.assignTeam(148, "R2")
	assert.Nil(t, err)
	dummyDs = &DriverStationConnection{TeamId: 148}
	arena.AllianceStations["R2"].DsConn = dummyDs

	arena.AllianceStations["R1"].DsConn.RobotLinked = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)

	arena.handleEstop("R1", true)
	arena.handleEstop("R2", false)
	assert.Equal(t, true, arena.AllianceStations["R1"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R1"].Estop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Estop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.Enabled)

	arena.handleEstop("R1", true)
	arena.handleEstop("R2", true)
	assert.Equal(t, true, arena.AllianceStations["R1"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R1"].Estop)
	assert.Equal(t, true, arena.AllianceStations["R2"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Estop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	arena.handleEstop("R1", false)
	arena.handleEstop("R2", true)
	assert.Equal(t, true, arena.AllianceStations["R1"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R1"].Estop)
	assert.Equal(t, true, arena.AllianceStations["R2"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Estop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec) * time.Second)
	arena.handleEstop("R1", false)
	arena.handleEstop("R2", true)
	assert.Equal(t, false, arena.AllianceStations["R1"].Astop)
	assert.Equal(t, false, arena.AllianceStations["R1"].Estop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Astop)
	assert.Equal(t, true, arena.AllianceStations["R2"].Estop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	arena.handleEstop("R1", true)
	arena.handleEstop("R2", false)
	assert.Equal(t, false, arena.AllianceStations["R1"].Astop)
	assert.Equal(t, true, arena.AllianceStations["R1"].Estop)
	assert.Equal(t, false, arena.AllianceStations["R2"].Astop)
	assert.Equal(t, true, arena.AllianceStations["R2"].Estop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)
}

func TestArenaTimeout(t *testing.T) {
	arena := setupTestArena(t)

	// Test regular ending of timeout.
	timeoutDurationSec := 9
	assert.Nil(t, arena.StartTimeout(timeoutDurationSec))
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test early cancellation of timeout.
	timeoutDurationSec = 28
	assert.Nil(t, arena.StartTimeout(timeoutDurationSec))
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Nil(t, arena.AbortMatch())
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test that timeout can't be started during a match.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.NotNil(t, arena.StartTimeout(1))
	assert.NotEqual(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec) *
		time.Second)
	for arena.MatchState != PostMatch {
		arena.Update()
		assert.NotNil(t, arena.StartTimeout(1))
	}
}

func TestSaveTeamHasConnected(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106, City: "San Jose", HasConnected: true})
	match := model.Match{Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 101}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 102, RobotLinked: true}
	arena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 103}
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].DsConn = &DriverStationConnection{TeamId: 104}
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].DsConn = &DriverStationConnection{TeamId: 105, RobotLinked: true}
	arena.AllianceStations["B3"].DsConn = &DriverStationConnection{TeamId: 106, RobotLinked: true}
	arena.AllianceStations["B3"].Team.City = "Sand Hosay" // Change some other field to verify that it isn't saved.
	assert.Nil(t, arena.StartMatch())

	// Check that the connection status was saved for the teams that just linked for the first time.
	teams, _ := arena.Database.GetAllTeams()
	if assert.Equal(t, 6, len(teams)) {
		assert.False(t, teams[0].HasConnected)
		assert.True(t, teams[1].HasConnected)
		assert.False(t, teams[2].HasConnected)
		assert.False(t, teams[3].HasConnected)
		assert.True(t, teams[4].HasConnected)
		assert.True(t, teams[5].HasConnected)
		assert.Equal(t, "San Jose", teams[5].City)
	}
}
