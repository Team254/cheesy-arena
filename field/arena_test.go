// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/playoff"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
		assert.Contains(t, err.Error(), "cannot start match until all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match until all robots are connected or bypassed")
	}
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.checkCanStartMatch())

	// Check PLC constraints.
	arena.Plc.SetAddress("1.2.3.4")
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while PLC is not healthy")
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

	// Check E-stop and bypass.
	arena.AllianceStations["B3"].EStop = true
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
	arena.AllianceStations["B3"].EStop = false
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
		assert.Contains(t, err.Error(), "cannot abort match when")
	}
	err = arena.StartMatch()
	assert.Nil(t, err)
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = AutoPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = PausePeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = TeleopPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = PostMatch
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match while")
	}
	err = arena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot abort match when")
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

	// Check with a single team E-stopped, A-stopped, not linked, and bypassed.
	arena.AllianceStations["R1"].EStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	arena.AllianceStations["R1"].EStop = false
	arena.AllianceStations["R1"].aStopReset = false
	arena.AllianceStations["R1"].AStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "if an autonomous stop has not been reset since the previous match")
	}
	arena.AllianceStations["R1"].aStopReset = true
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
	arena.AllianceStations["B3"].EStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "while an emergency stop is active")
	}
	arena.AllianceStations["B3"].EStop = false
	err = arena.StartMatch()
	assert.Nil(t, err)
}

func TestLoadNextMatch(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 1114})
	practiceMatch1 := model.Match{Type: model.Practice, TypeOrder: 1}
	practiceMatch2 := model.Match{Type: model.Practice, TypeOrder: 2, Status: game.RedWonMatch}
	practiceMatch3 := model.Match{Type: model.Practice, TypeOrder: 3}
	arena.Database.CreateMatch(&practiceMatch1)
	arena.Database.CreateMatch(&practiceMatch2)
	arena.Database.CreateMatch(&practiceMatch3)
	qualificationMatch1 := model.Match{Type: model.Qualification, TypeOrder: 1, Status: game.BlueWonMatch}
	qualificationMatch2 := model.Match{Type: model.Qualification, TypeOrder: 2}
	arena.Database.CreateMatch(&qualificationMatch1)
	arena.Database.CreateMatch(&qualificationMatch2)

	// Test match should be followed by another, empty test match.
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	err := arena.SubstituteTeams(1114, 0, 0, 0, 0, 0)
	assert.Nil(t, err)
	arena.CurrentMatch.Status = game.TieMatch
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, 0, arena.CurrentMatch.Red1)
	assert.Equal(t, false, arena.CurrentMatch.IsComplete())

	// Other matches should be loaded by type until they're all complete.
	err = arena.LoadMatch(&practiceMatch2)
	assert.Nil(t, err)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch1.Id, arena.CurrentMatch.Id)
	practiceMatch1.Status = game.RedWonMatch
	arena.Database.UpdateMatch(&practiceMatch1)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch3.Id, arena.CurrentMatch.Id)
	practiceMatch3.Status = game.BlueWonMatch
	arena.Database.UpdateMatch(&practiceMatch3)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, model.Test, arena.CurrentMatch.Type)

	err = arena.LoadMatch(&qualificationMatch1)
	assert.Nil(t, err)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, qualificationMatch2.Id, arena.CurrentMatch.Id)
}

func TestSubstituteTeam(t *testing.T) {
	arena := setupTestArena(t)
	tournament.CreateTestAlliances(arena.Database, 2)
	arena.PlayoffTournament, _ = playoff.NewPlayoffTournament(
		arena.EventSettings.PlayoffType, arena.EventSettings.NumPlayoffAlliances,
	)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106})
	arena.Database.CreateTeam(&model.Team{Id: 107})

	// Substitute teams into test match.
	err := arena.SubstituteTeams(0, 0, 0, 101, 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, 101, arena.CurrentMatch.Blue1)
	assert.Equal(t, 101, arena.AllianceStations["B1"].Team.Id)
	err = arena.assignTeam(104, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}

	// Substitute teams into practice match.
	match := model.Match{Type: model.Practice, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeams(107, 102, 103, 104, 105, 106)
	assert.Nil(t, err)
	assert.Equal(t, 107, arena.CurrentMatch.Red1)
	assert.Equal(t, 107, arena.AllianceStations["R1"].Team.Id)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = arena.CurrentMatch.Id

	// Check that substitution is disallowed in qualification matches.
	match = model.Match{Type: model.Qualification, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeams(107, 102, 103, 104, 105, 106)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Can't substitute teams for qualification matches.")
	}
	match = model.Match{Type: model.Playoff, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	assert.Nil(t, arena.SubstituteTeams(107, 102, 103, 104, 105, 106))

	// Check that loading a nonexistent team fails.
	err = arena.SubstituteTeams(101, 102, 103, 104, 105, 108)
	if assert.NotNil(t, err) {
		assert.Equal(t, err.Error(), "Team 108 is not present at the event.")
	}
}

func TestLoadTeamsFromNexus(t *testing.T) {
	arena := setupTestArena(t)

	for i := 1; i <= 12; i++ {
		arena.Database.CreateTeam(&model.Team{Id: 100 + i})
	}
	match := model.Match{
		Type:        model.Practice,
		Red1:        101,
		Red2:        102,
		Red3:        103,
		Blue1:       104,
		Blue2:       105,
		Blue3:       106,
		TbaMatchKey: model.TbaMatchKey{CompLevel: "p", SetNumber: 0, MatchNumber: 1},
	}
	arena.Database.CreateMatch(&match)

	assertTeams := func(red1, red2, red3, blue1, blue2, blue int) {
		assert.Equal(t, red1, arena.CurrentMatch.Red1)
		assert.Equal(t, red2, arena.CurrentMatch.Red2)
		assert.Equal(t, red3, arena.CurrentMatch.Red3)
		assert.Equal(t, blue1, arena.CurrentMatch.Blue1)
		assert.Equal(t, blue2, arena.CurrentMatch.Blue2)
		assert.Equal(t, blue, arena.CurrentMatch.Blue3)
		assert.Equal(t, red1, arena.AllianceStations["R1"].Team.Id)
		assert.Equal(t, red2, arena.AllianceStations["R2"].Team.Id)
		assert.Equal(t, red3, arena.AllianceStations["R3"].Team.Id)
		assert.Equal(t, blue1, arena.AllianceStations["B1"].Team.Id)
		assert.Equal(t, blue2, arena.AllianceStations["B2"].Team.Id)
		assert.Equal(t, blue, arena.AllianceStations["B3"].Team.Id)
	}

	// Sanity check that the match loads correctly without Nexus enabled.
	assert.Nil(t, arena.LoadMatch(&match))
	assertTeams(101, 102, 103, 104, 105, 106)

	// Mock the Nexus server.
	nexusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/api/v1/event/my_event_code/match/p1/lineup") {
			w.Write([]byte("{\"red\":[\"112\",\"111\",\"110\"],\"blue\":[\"109\",\"108\",\"107\"]}"))
		} else {
			http.Error(w, "Match not found", 404)
		}
	}))
	defer nexusServer.Close()
	arena.NexusClient = partner.NewNexusClient("my_event_code")
	arena.NexusClient.BaseUrl = nexusServer.URL
	arena.EventSettings.NexusEnabled = true

	// Check that the correct teams are loaded from Nexus.
	assert.Nil(t, arena.LoadMatch(&match))
	assertTeams(112, 111, 110, 109, 108, 107)

	// Check with a match that Nexus doesn't know about.
	match = model.Match{
		Type:        model.Practice,
		Red1:        106,
		Red2:        105,
		Red3:        104,
		Blue1:       103,
		Blue2:       102,
		Blue3:       101,
		TbaMatchKey: model.TbaMatchKey{CompLevel: "p", SetNumber: 0, MatchNumber: 2},
	}
	arena.Database.CreateMatch(&match)
	assert.Nil(t, arena.LoadMatch(&match))
	assertTeams(106, 105, 104, 103, 102, 101)
}

func TestArenaTimeout(t *testing.T) {
	arena := setupTestArena(t)

	// Test regular ending of timeout.
	timeoutDurationSec := 9
	assert.Nil(t, arena.StartTimeout("Break 1", timeoutDurationSec))
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, "Break 1", arena.breakDescription)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test early cancellation of timeout.
	timeoutDurationSec = 28
	assert.Nil(t, arena.StartTimeout("Break 2", timeoutDurationSec))
	assert.Equal(t, "Break 2", arena.breakDescription)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
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
	assert.NotNil(t, arena.StartTimeout("Timeout", 1))
	assert.NotEqual(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec) *
		time.Second)
	for arena.MatchState != PostMatch {
		arena.Update()
		assert.NotNil(t, arena.StartTimeout("Timeout", 1))
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

func TestPlcEStopAStop(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	arena.Plc = &plc

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
	arena.AllianceStations["R1"].aStopReset = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R2"].aStopReset = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["R3"].aStopReset = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B1"].aStopReset = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B2"].aStopReset = true
	arena.AllianceStations["B3"].Bypass = true
	arena.AllianceStations["B3"].aStopReset = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)

	// Press the R1 A-stop.
	plc.redAStops[0] = true
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.Enabled)

	// Unpress the R1 A-stop and press the R2 E-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = true
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.AStop)

	// Unpress the R2 E-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Transition into the teleop period without any stops.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec) * time.Second,
	)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec,
		) * time.Second,
	)
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Press the R1 E-stop and the R2 A-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = true
	plc.redAStops[1] = true
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Ensure the other stations A-stops are working as well.
	plc.redAStops[2] = true
	plc.redEStops[2] = false
	plc.blueAStops[0] = true
	plc.blueEStops[0] = false
	plc.blueAStops[1] = true
	plc.blueEStops[1] = false
	plc.blueAStops[2] = true
	plc.blueEStops[2] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R3"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B3"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].EStop)

	// Ensure the other stations E-stops are working as well.
	plc.redAStops[2] = false
	plc.redEStops[2] = true
	plc.blueAStops[0] = false
	plc.blueEStops[0] = true
	plc.blueAStops[1] = false
	plc.blueEStops[1] = true
	plc.blueAStops[2] = false
	plc.blueEStops[2] = true
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R3"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B1"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B3"].EStop)

	// Ensure unpressed E-stops are cleared at the end of the match.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+
				game.MatchTiming.TeleopDurationSec,
		) * time.Second,
	)
	arena.Update()
	plc.blueEStops[2] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].EStop)
}

func TestPlcEStopAStopWithPlcDisabled(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = false
	arena.Plc = &plc

	arena.Database.CreateTeam(&model.Team{Id: 254})
	err := arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 1323}

	arena.AllianceStations["R1"].DsConn.RobotLinked = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)

	plc.redEStops[0] = true
	plc.redAStops[1] = true
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.Enabled)
}

func TestPlcFieldEStop(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	arena.Plc = &plc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	plc.fieldEStop = true
	arena.Update()
	assert.True(t, arena.matchAborted)
	assert.Equal(t, PostMatch, arena.MatchState)
}

func TestPlcFieldEStopWithPlcDisabled(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = false
	arena.Plc = &plc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	plc.fieldEStop = true
	arena.Update()
	assert.False(t, arena.matchAborted)
	assert.Equal(t, AutoPeriod, arena.MatchState)
}

func TestPlcMatchCycleEvergreen(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	arena.Plc = &plc

	arena.Update()
	assert.Equal(t, [4]bool{true, true, false, false}, plc.stackLights)

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.Update()
	assert.Equal(t, [4]bool{true, true, false, false}, plc.stackLights)

	arena.AllianceStations["R3"].Bypass = true
	arena.Update()
	assert.Equal(t, [4]bool{false, true, false, false}, plc.stackLights)
	assert.Equal(t, false, plc.stackLightBuzzer)

	// All teams are ready.
	arena.AllianceStations["B3"].Bypass = true
	plc.cycleState = true
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
	assert.Equal(t, true, plc.stackLightBuzzer)

	// Green light when blink cycle is off.
	plc.cycleState = false
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, false}, plc.stackLights)

	// Start the match.
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
	assert.Equal(t, false, plc.stackLightBuzzer)

	// End the match.
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec+
		game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec) *
		time.Second)
	arena.Update()
	arena.Update()
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, [4]bool{false, false, true, false}, plc.stackLights)
	assert.Equal(t, false, plc.fieldResetLight)

	// Ready the score.
	arena.RedRealtimeScore.FoulsCommitted = true
	arena.BlueRealtimeScore.FoulsCommitted = true
	redWs := &websocket.Websocket{}
	arena.ScoringPanelRegistry.RegisterPanel("red", redWs)
	arena.ScoringPanelRegistry.SetScoreCommitted("red", redWs)
	arena.Update()
	assert.Equal(t, [4]bool{false, false, true, false}, plc.stackLights)
	blueWs := &websocket.Websocket{}
	arena.ScoringPanelRegistry.RegisterPanel("blue", blueWs)
	arena.ScoringPanelRegistry.SetScoreCommitted("blue", blueWs)
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, false}, plc.stackLights)

	arena.FieldReset = true
	arena.Update()
	assert.Equal(t, true, plc.fieldResetLight)
}

func TestPlcMatchCycleGameSpecific(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	arena.Plc = &plc

	// Check that no inputs or outputs are active before the match starts.
	assert.Equal(t, PreMatch, arena.MatchState)
	plc.redNoteCounts = [2]int{1, 2}
	plc.blueNoteCounts = [2]int{3, 4}
	plc.redAmpButtons = [2]bool{true, true}
	plc.blueAmpButtons = [2]bool{true, true}
	arena.Update()
	redAmpSpeaker := &arena.RedRealtimeScore.CurrentScore.AmpSpeaker
	blueAmpSpeaker := &arena.BlueRealtimeScore.CurrentScore.AmpSpeaker
	assert.Equal(t, game.AmpSpeaker{}, *redAmpSpeaker)
	assert.Equal(t, game.AmpSpeaker{}, *blueAmpSpeaker)
	assert.Equal(t, [3]bool{false, false, false}, plc.redAmpLights)
	assert.Equal(t, [3]bool{false, false, false}, plc.blueAmpLights)
	assert.Equal(t, false, plc.redSubwooferCountdown)
	assert.Equal(t, false, plc.blueSubwooferCountdown)
	assert.Equal(t, false, plc.speakerMotors)
	assert.Equal(t, false, plc.postMatchSubwooferLights)
	plc.redNoteCounts = [2]int{0, 0}
	plc.blueNoteCounts = [2]int{0, 0}
	plc.redAmpButtons = [2]bool{false, false}
	plc.blueAmpButtons = [2]bool{false, false}

	// Start the match.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	arena.Update()
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Duration(game.MatchTiming.WarmupDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, plc.speakerMotors)

	// Check the autonomous period.
	plc.redNoteCounts = [2]int{1, 2}
	plc.blueNoteCounts = [2]int{3, 4}
	plc.redAmpButtons = [2]bool{true, true}
	plc.blueAmpButtons = [2]bool{true, true}
	arena.Update()
	assert.Equal(t, 1, redAmpSpeaker.AutoAmpNotes)
	assert.Equal(t, 2, redAmpSpeaker.AutoSpeakerNotes)
	assert.Equal(t, 3, blueAmpSpeaker.AutoAmpNotes)
	assert.Equal(t, 4, blueAmpSpeaker.AutoSpeakerNotes)
	assert.Equal(t, [3]bool{true, false, false}, plc.redAmpLights)
	assert.Equal(t, [3]bool{true, true, false}, plc.blueAmpLights)

	// Check the pause period.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec) * time.Second,
	)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	assert.Equal(t, true, plc.speakerMotors)
	assert.Equal(t, [3]bool{true, false, false}, plc.redAmpLights)
	assert.Equal(t, [3]bool{true, true, false}, plc.blueAmpLights)
	plc.redAmpButtons = [2]bool{false, false}
	plc.blueAmpButtons = [2]bool{false, false}

	// Check the teleop period.
	durationToTeleopStart := time.Duration(
		game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec,
	) * time.Second
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 5000*time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, true, plc.speakerMotors)

	// Trigger coop for the red alliance.
	plc.redAmpButtons = [2]bool{false, true}
	arena.Update()
	assert.Equal(t, true, redAmpSpeaker.CoopActivated)
	assert.Equal(t, [3]bool{false, false, true}, plc.redAmpLights)
	assert.Equal(t, false, blueAmpSpeaker.CoopActivated)
	assert.Equal(t, [3]bool{true, true, false}, plc.blueAmpLights)

	// Trigger coop for the blue alliance.
	plc.blueAmpButtons = [2]bool{false, true}
	arena.Update()
	assert.Equal(t, true, blueAmpSpeaker.CoopActivated)
	assert.Equal(t, [3]bool{true, false, true}, plc.blueAmpLights)

	// Do some scoring and trigger an amplified cycle.
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 10000*time.Millisecond)
	plc.redAmpButtons = [2]bool{false, false}
	plc.blueAmpButtons = [2]bool{false, false}
	plc.blueNoteCounts = [2]int{4, 6}
	arena.Update()
	assert.Equal(t, 3, blueAmpSpeaker.AutoAmpNotes)
	assert.Equal(t, 4, blueAmpSpeaker.AutoSpeakerNotes)
	assert.Equal(t, 1, blueAmpSpeaker.TeleopAmpNotes)
	assert.Equal(t, 2, blueAmpSpeaker.TeleopUnamplifiedSpeakerNotes)
	assert.Equal(t, 0, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	assert.Equal(t, false, plc.blueSpeakerLight)
	assert.Equal(t, false, plc.blueSubwooferCountdown)
	plc.blueAmpButtons = [2]bool{true, false}
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 10200*time.Millisecond)
	arena.Update()
	assert.Equal(t, false, plc.redSpeakerLight)
	assert.Equal(t, false, plc.redSubwooferCountdown)
	assert.Equal(t, true, plc.blueSpeakerLight)
	assert.Equal(t, true, plc.blueSubwooferCountdown)
	plc.blueAmpButtons = [2]bool{false, false}
	plc.blueNoteCounts = [2]int{4, 9}
	arena.Update()
	assert.Equal(t, 2, blueAmpSpeaker.TeleopUnamplifiedSpeakerNotes)
	assert.Equal(t, 3, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	blueAmpSpeaker.LastAmplifiedTime = time.Now().Add(-11 * time.Second)
	arena.Update()
	assert.Equal(t, false, plc.blueSpeakerLight)
	assert.Equal(t, false, plc.blueSubwooferCountdown)
	plc.blueNoteCounts = [2]int{4, 11}
	arena.Update()
	assert.Equal(t, 3, blueAmpSpeaker.TeleopUnamplifiedSpeakerNotes)
	assert.Equal(t, 4, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)

	// Trigger an amplified cycle that ends early.
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 20000*time.Millisecond)
	plc.redNoteCounts = [2]int{3, 2}
	plc.redAmpButtons = [2]bool{true, false}
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 20001*time.Millisecond)
	arena.Update()
	assert.Equal(t, true, plc.redSpeakerLight)
	assert.Equal(t, true, plc.redSubwooferCountdown)
	assert.Equal(t, false, plc.blueSpeakerLight)
	assert.Equal(t, false, plc.blueSubwooferCountdown)
	arena.MatchStartTime = time.Now().Add(-durationToTeleopStart - 23000*time.Millisecond)
	plc.redNoteCounts = [2]int{3, 6}
	plc.redAmpButtons = [2]bool{false, false}
	arena.Update()
	assert.Equal(t, 4, redAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	assert.Equal(t, false, plc.redSpeakerLight)
	assert.Equal(t, false, plc.redSubwooferCountdown)
	assert.Equal(t, false, plc.blueSpeakerLight)
	assert.Equal(t, false, plc.blueSubwooferCountdown)

	// Trigger an amplified cycle that ends after the match is over.
	durationToTeleopEnd := time.Duration(
		game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+
			game.MatchTiming.TeleopDurationSec,
	) * time.Second
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd + 1000*time.Millisecond)
	plc.blueNoteCounts = [2]int{6, 11}
	plc.blueAmpButtons = [2]bool{true, false}
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd + 999*time.Millisecond)
	arena.Update()
	assert.Equal(t, true, plc.blueSpeakerLight)
	assert.Equal(t, true, plc.blueSubwooferCountdown)
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd + 1*time.Millisecond)
	plc.redNoteCounts = [2]int{5, 6}
	plc.blueNoteCounts = [2]int{6, 12}
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, 5, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	assert.Equal(t, [3]bool{true, true, true}, plc.redAmpLights)
	assert.Equal(t, [3]bool{false, true, true}, plc.blueAmpLights)
	assert.Equal(t, true, plc.blueSpeakerLight)
	assert.Equal(t, true, plc.blueSubwooferCountdown)
	assert.Equal(t, false, plc.postMatchSubwooferLights)
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd - 1*time.Millisecond)
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, [3]bool{false, false, false}, plc.redAmpLights)
	assert.Equal(t, [3]bool{false, false, false}, plc.blueAmpLights)
	assert.Equal(t, false, plc.blueSpeakerLight)
	assert.Equal(t, false, plc.blueSubwooferCountdown)
	assert.Equal(t, true, plc.speakerMotors)
	assert.Equal(t, true, plc.postMatchSubwooferLights)
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd - 4999*time.Millisecond)
	plc.blueNoteCounts = [2]int{6, 13}
	arena.Update()
	assert.Equal(t, 6, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	assert.Equal(t, true, plc.speakerMotors)
	assert.Equal(t, true, plc.postMatchSubwooferLights)
	arena.MatchStartTime = time.Now().Add(-durationToTeleopEnd - 5001*time.Millisecond)
	plc.blueNoteCounts = [2]int{6, 20}
	arena.Update()
	assert.Equal(t, 3, blueAmpSpeaker.TeleopUnamplifiedSpeakerNotes)
	assert.Equal(t, 6, blueAmpSpeaker.TeleopAmplifiedSpeakerNotes)
	assert.Equal(t, [3]bool{false, false, false}, plc.redAmpLights)
	assert.Equal(t, [3]bool{false, false, false}, plc.blueAmpLights)
	assert.Equal(t, false, plc.speakerMotors)
	assert.Equal(t, false, plc.postMatchSubwooferLights)
}
