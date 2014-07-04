// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for controlling the arena and match play.

package main

import (
	"fmt"
	"time"
)

// Loop and match timing constants.
const arenaLoopPeriodMs = 1
const dsPacketPeriodMs = 250
const autoDurationSec = 10
const pauseDurationSec = 1
const teleopDurationSec = 140
const endgameTimeLeftSec = 30

// Progression of match states.
const (
	PRE_MATCH = iota
	START_MATCH
	AUTO_PERIOD
	PAUSE_PERIOD
	TELEOP_PERIOD
	ENDGAME_PERIOD
	POST_MATCH
)

type AllianceStation struct {
	team                    *Team
	driverStationConnection *DriverStationConnection
	emergencyStop           bool
	bypass                  bool
}

type Arena struct {
	allianceStations map[string]*AllianceStation
	currentMatch     *Match
	matchState       int
	matchStartTime   time.Time
	lastDsPacketTime time.Time
}

var mainArena Arena // Named thusly to avoid polluting the global namespace with something more generic.

// Sets the arena to its initial state.
func (arena *Arena) Setup() {
	arena.allianceStations = make(map[string]*AllianceStation)
	arena.allianceStations["R1"] = new(AllianceStation)
	arena.allianceStations["R2"] = new(AllianceStation)
	arena.allianceStations["R3"] = new(AllianceStation)
	arena.allianceStations["B1"] = new(AllianceStation)
	arena.allianceStations["B2"] = new(AllianceStation)
	arena.allianceStations["B3"] = new(AllianceStation)

	// Load empty match as current.
	arena.matchState = PRE_MATCH
	arena.LoadMatch(new(Match))
}

// Loads a team into an alliance station, cleaning up the previous team there if there is one.
func (arena *Arena) AssignTeam(teamId int, station string) error {
	// Reject invalid station values.
	if _, ok := arena.allianceStations[station]; !ok {
		return fmt.Errorf("Invalid alliance station '%s'.", station)
	}
	// Do nothing if the station is already assigned to the requested team.
	dsConn := arena.allianceStations[station].driverStationConnection
	if dsConn != nil && dsConn.TeamId == teamId {
		return nil
	}
	if dsConn != nil {
		err := dsConn.Close()
		if err != nil {
			return err
		}
		arena.allianceStations[station].team = nil
		arena.allianceStations[station].driverStationConnection = nil
	}

	// Leave the station empty if the team number is zero.
	if teamId == 0 {
		return nil
	}

	// Load the team model. Raise an error if a team doesn't exist.
	team, err := db.GetTeamById(teamId)
	if err != nil {
		return err
	}
	if team == nil {
		return fmt.Errorf("Invalid team number '%d'.", teamId)
	}

	arena.allianceStations[station].team = team
	arena.allianceStations[station].driverStationConnection, err = NewDriverStationConnection(team.Id, station)
	if err != nil {
		return err
	}
	return nil
}

// Sets up the arena for the given match.
func (arena *Arena) LoadMatch(match *Match) error {
	if arena.matchState != PRE_MATCH {
		return fmt.Errorf("Cannot load match while there is a match still in progress or with results pending.")
	}

	arena.currentMatch = match
	err := arena.AssignTeam(match.Red1, "R1")
	if err != nil {
		return err
	}
	err = arena.AssignTeam(match.Red2, "R2")
	if err != nil {
		return err
	}
	err = arena.AssignTeam(match.Red3, "R3")
	if err != nil {
		return err
	}
	err = arena.AssignTeam(match.Blue1, "B1")
	if err != nil {
		return err
	}
	err = arena.AssignTeam(match.Blue2, "B2")
	if err != nil {
		return err
	}
	err = arena.AssignTeam(match.Blue3, "B3")
	if err != nil {
		return err
	}
	return nil
}

// Starts the match if all conditions are met.
func (arena *Arena) StartMatch() error {
	if arena.matchState != PRE_MATCH {
		return fmt.Errorf("Cannot start match while there is a match still in progress or with results pending.")
	}
	if arena.currentMatch == nil {
		return fmt.Errorf("Cannot start match when no match is loaded.")
	}
	arena.matchState = START_MATCH
	return nil
}

// Clears out the match and resets the arena state unless there is a match underway.
func (arena *Arena) ResetMatch() error {
	if arena.matchState != POST_MATCH && arena.matchState != PRE_MATCH {
		return fmt.Errorf("Cannot reset match while it is in progress.")
	}
	arena.matchState = PRE_MATCH
	arena.currentMatch = nil
	return nil
}

// Performs a single iteration of checking inputs and timers and setting outputs accordingly to control the
// flow of a match.
func (arena *Arena) Update() {
	// Decide what state the robots need to be in, depending on where we are in the match.
	auto := false
	enabled := false
	sendDsPacket := false
	matchTimeSec := arena.MatchTimeSec()
	switch arena.matchState {
	case PRE_MATCH:
		auto = true
		enabled = false
	case START_MATCH:
		arena.matchState = AUTO_PERIOD
		arena.matchStartTime = time.Now()
		auto = true
		enabled = true
		sendDsPacket = true
	case AUTO_PERIOD:
		auto = true
		enabled = true
		if matchTimeSec >= autoDurationSec {
			arena.matchState = PAUSE_PERIOD
			auto = false
			enabled = false
			sendDsPacket = true
		}
	case PAUSE_PERIOD:
		auto = false
		enabled = false
		if matchTimeSec >= autoDurationSec+pauseDurationSec {
			arena.matchState = TELEOP_PERIOD
			auto = false
			enabled = true
			sendDsPacket = true
		}
	case TELEOP_PERIOD:
		auto = false
		enabled = true
		if matchTimeSec >= autoDurationSec+pauseDurationSec+teleopDurationSec-endgameTimeLeftSec {
			arena.matchState = ENDGAME_PERIOD
			sendDsPacket = false
		}
	case ENDGAME_PERIOD:
		auto = false
		enabled = true
		if matchTimeSec >= autoDurationSec+pauseDurationSec+teleopDurationSec {
			arena.matchState = POST_MATCH
			auto = false
			enabled = false
			sendDsPacket = true
		}
	}

	// Send a packet if at a period transition point or if it's been long enough since the last one.
	if sendDsPacket || time.Since(arena.lastDsPacketTime).Seconds()*1000 >= dsPacketPeriodMs {
		arena.sendDsPacket(auto, enabled)
	}
}

// Loops indefinitely to track and update the arena components.
func (arena *Arena) Run() {
	for {
		arena.Update()
		time.Sleep(time.Millisecond * arenaLoopPeriodMs)
	}
}

func (arena *Arena) sendDsPacket(auto bool, enabled bool) {
	for _, allianceStation := range arena.allianceStations {
		dsConn := allianceStation.driverStationConnection
		if dsConn != nil {
			dsConn.Auto = auto
			dsConn.Enabled = enabled && !allianceStation.emergencyStop && !allianceStation.bypass
			err := dsConn.Update()
			if err != nil {
				// TODO(pat): Handle errors.
			}
		}
	}
	arena.lastDsPacketTime = time.Now()
}

// Returns the fractional number of seconds since the start of the match.
func (arena *Arena) MatchTimeSec() float64 {
	if arena.matchState == PRE_MATCH || arena.matchState == POST_MATCH {
		return 0
	} else {
		return time.Since(arena.matchStartTime).Seconds()
	}
}
