// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for controlling the arena and match play.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"log"
	"time"
)

const (
	arenaLoopPeriodMs     = 10
	dsPacketPeriodMs      = 250
	matchEndScoreDwellSec = 3
)

// Progression of match states.
const (
	PreMatch      = 0
	StartMatch    = 1
	AutoPeriod    = 2
	PausePeriod   = 3
	TeleopPeriod  = 4
	EndgamePeriod = 5
	PostMatch     = 6
)

type Arena struct {
	Database                       *model.Database
	EventSettings                  *model.EventSettings
	accessPoint                    *AccessPoint
	networkSwitch                  *NetworkSwitch
	Plc                            Plc
	TbaClient                      *partner.TbaClient
	StemTvClient                   *partner.StemTvClient
	StemTvClient2                  *partner.StemTvClient
	AllianceStations               map[string]*AllianceStation
	CurrentMatch                   *model.Match
	MatchState                     int
	lastMatchState                 int
	MatchStartTime                 time.Time
	LastMatchTimeSec               float64
	RedRealtimeScore               *RealtimeScore
	BlueRealtimeScore              *RealtimeScore
	lastDsPacketTime               time.Time
	FieldReset                     bool
	AudienceDisplayScreen          string
	SavedMatch                     *model.Match
	SavedMatchResult               *model.MatchResult
	AllianceStationDisplays        map[string]string
	AllianceStationDisplayScreen   string
	MuteMatchSounds                bool
	FieldTestMode                  string
	matchAborted                   bool
	matchStateNotifier             *Notifier
	MatchTimeNotifier              *Notifier
	RobotStatusNotifier            *Notifier
	MatchLoadTeamsNotifier         *Notifier
	ScoringStatusNotifier          *Notifier
	RealtimeScoreNotifier          *Notifier
	ScorePostedNotifier            *Notifier
	AudienceDisplayNotifier        *Notifier
	PlaySoundNotifier              *Notifier
	AllianceStationDisplayNotifier *Notifier
	AllianceSelectionNotifier      *Notifier
	LowerThirdNotifier             *Notifier
	ReloadDisplaysNotifier         *Notifier
}

type ArenaStatus struct {
	AllianceStations map[string]*AllianceStation
	MatchState       int
	CanStartMatch    bool
	PlcIsHealthy     bool
	FieldEstop       bool
}

type AllianceStation struct {
	DsConn *DriverStationConnection
	Estop  bool
	Bypass bool
	Team   *model.Team
}

// Creates the arena and sets it to its initial state.
func NewArena(dbPath string) (*Arena, error) {
	arena := new(Arena)

	var err error
	arena.Database, err = model.OpenDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	err = arena.LoadSettings()
	if err != nil {
		return nil, err
	}

	arena.AllianceStations = make(map[string]*AllianceStation)
	arena.AllianceStations["R1"] = new(AllianceStation)
	arena.AllianceStations["R2"] = new(AllianceStation)
	arena.AllianceStations["R3"] = new(AllianceStation)
	arena.AllianceStations["B1"] = new(AllianceStation)
	arena.AllianceStations["B2"] = new(AllianceStation)
	arena.AllianceStations["B3"] = new(AllianceStation)

	arena.matchStateNotifier = NewNotifier()
	arena.MatchTimeNotifier = NewNotifier()
	arena.RobotStatusNotifier = NewNotifier()
	arena.MatchLoadTeamsNotifier = NewNotifier()
	arena.ScoringStatusNotifier = NewNotifier()
	arena.RealtimeScoreNotifier = NewNotifier()
	arena.ScorePostedNotifier = NewNotifier()
	arena.AudienceDisplayNotifier = NewNotifier()
	arena.PlaySoundNotifier = NewNotifier()
	arena.AllianceStationDisplayNotifier = NewNotifier()
	arena.AllianceSelectionNotifier = NewNotifier()
	arena.LowerThirdNotifier = NewNotifier()
	arena.ReloadDisplaysNotifier = NewNotifier()

	// Load empty match as current.
	arena.MatchState = PreMatch
	arena.LoadTestMatch()
	arena.lastMatchState = -1
	arena.LastMatchTimeSec = 0

	// Initialize display parameters.
	arena.AudienceDisplayScreen = "blank"
	arena.SavedMatch = &model.Match{}
	arena.SavedMatchResult = model.NewMatchResult()
	arena.AllianceStationDisplays = make(map[string]string)
	arena.AllianceStationDisplayScreen = "match"

	return arena, nil
}

// Loads or reloads the event settings upon initial setup or change.
func (arena *Arena) LoadSettings() error {
	settings, err := arena.Database.GetEventSettings()
	if err != nil {
		return err
	}
	arena.EventSettings = settings

	// Initialize the components that depend on settings.
	arena.accessPoint = NewAccessPoint(settings.ApAddress, settings.ApUsername, settings.ApPassword)
	arena.networkSwitch = NewNetworkSwitch(settings.SwitchAddress, settings.SwitchPassword)
	arena.Plc.SetAddress(settings.PlcAddress)
	arena.TbaClient = partner.NewTbaClient(settings.TbaEventCode, settings.TbaSecretId, settings.TbaSecret)
	arena.StemTvClient = partner.NewStemTvClient(settings.StemTvEventCode)
	arena.StemTvClient2 = partner.NewStemTvClient("2017cc2")
	arena.StemTvClient2.BaseUrl = "http://52.20.77.69:5000"

	return nil
}

// Sets up the arena for the given match.
func (arena *Arena) LoadMatch(match *model.Match) error {
	if arena.MatchState != PreMatch {
		return fmt.Errorf("Cannot load match while there is a match still in progress or with results pending.")
	}

	arena.CurrentMatch = match
	err := arena.assignTeam(match.Red1, "R1")
	if err != nil {
		return err
	}
	err = arena.assignTeam(match.Red2, "R2")
	if err != nil {
		return err
	}
	err = arena.assignTeam(match.Red3, "R3")
	if err != nil {
		return err
	}
	err = arena.assignTeam(match.Blue1, "B1")
	if err != nil {
		return err
	}
	err = arena.assignTeam(match.Blue2, "B2")
	if err != nil {
		return err
	}
	err = arena.assignTeam(match.Blue3, "B3")
	if err != nil {
		return err
	}

	arena.setupNetwork()

	// Reset the realtime scores.
	arena.RedRealtimeScore = NewRealtimeScore()
	arena.BlueRealtimeScore = NewRealtimeScore()
	arena.Plc.ResetCounts()
	arena.FieldReset = false

	// Notify any listeners about the new match.
	arena.MatchLoadTeamsNotifier.Notify(nil)
	arena.RealtimeScoreNotifier.Notify(nil)
	arena.AllianceStationDisplayScreen = "match"
	arena.AllianceStationDisplayNotifier.Notify(nil)

	return nil
}

// Sets a new test match containing no teams as the current match.
func (arena *Arena) LoadTestMatch() error {
	return arena.LoadMatch(&model.Match{Type: "test"})
}

// Loads the first unplayed match of the current match type.
func (arena *Arena) LoadNextMatch() error {
	if arena.CurrentMatch.Type == "test" {
		return arena.LoadTestMatch()
	}

	matches, err := arena.Database.GetMatchesByType(arena.CurrentMatch.Type)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if match.Status != "complete" {
			err = arena.LoadMatch(&match)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// Assigns the given team to the given station, also substituting it into the match record.
func (arena *Arena) SubstituteTeam(teamId int, station string) error {
	if arena.CurrentMatch.Type == "qualification" {
		return fmt.Errorf("Can't substitute teams for qualification matches.")
	}
	err := arena.assignTeam(teamId, station)
	if err != nil {
		return err
	}
	switch station {
	case "R1":
		arena.CurrentMatch.Red1 = teamId
	case "R2":
		arena.CurrentMatch.Red2 = teamId
	case "R3":
		arena.CurrentMatch.Red3 = teamId
	case "B1":
		arena.CurrentMatch.Blue1 = teamId
	case "B2":
		arena.CurrentMatch.Blue2 = teamId
	case "B3":
		arena.CurrentMatch.Blue3 = teamId
	}
	arena.setupNetwork()
	arena.MatchLoadTeamsNotifier.Notify(nil)
	return nil
}

// Starts the match if all conditions are met.
func (arena *Arena) StartMatch() error {
	err := arena.checkCanStartMatch()
	if err == nil {
		// Save the match start time to the database for posterity.
		arena.CurrentMatch.StartedAt = time.Now()
		if arena.CurrentMatch.Type != "test" {
			arena.Database.SaveMatch(arena.CurrentMatch)
		}

		// Save the missed packet count to subtract it from the running count.
		for _, allianceStation := range arena.AllianceStations {
			if allianceStation.DsConn != nil {
				err = allianceStation.DsConn.signalMatchStart(arena.CurrentMatch)
				if err != nil {
					log.Println(err)
				}
			}
		}

		arena.MatchState = StartMatch
	}
	return err
}

// Kills the current match if it is underway.
func (arena *Arena) AbortMatch() error {
	if arena.MatchState == PreMatch || arena.MatchState == PostMatch {
		return fmt.Errorf("Cannot abort match when it is not in progress.")
	}
	arena.MatchState = PostMatch
	arena.matchAborted = true
	arena.AudienceDisplayScreen = "blank"
	arena.AudienceDisplayNotifier.Notify(nil)
	if !arena.MuteMatchSounds {
		arena.PlaySoundNotifier.Notify("match-abort")
	}
	return nil
}

// Clears out the match and resets the arena state unless there is a match underway.
func (arena *Arena) ResetMatch() error {
	if arena.MatchState != PostMatch && arena.MatchState != PreMatch {
		return fmt.Errorf("Cannot reset match while it is in progress.")
	}
	arena.MatchState = PreMatch
	arena.matchAborted = false
	arena.AllianceStations["R1"].Bypass = false
	arena.AllianceStations["R2"].Bypass = false
	arena.AllianceStations["R3"].Bypass = false
	arena.AllianceStations["B1"].Bypass = false
	arena.AllianceStations["B2"].Bypass = false
	arena.AllianceStations["B3"].Bypass = false
	arena.MuteMatchSounds = false
	return nil
}

// Returns the fractional number of seconds since the start of the match.
func (arena *Arena) MatchTimeSec() float64 {
	if arena.MatchState == PreMatch || arena.MatchState == StartMatch || arena.MatchState == PostMatch {
		return 0
	} else {
		return time.Since(arena.MatchStartTime).Seconds()
	}
}

// Performs a single iteration of checking inputs and timers and setting outputs accordingly to control the
// flow of a match.
func (arena *Arena) Update() {
	// Decide what state the robots need to be in, depending on where we are in the match.
	auto := false
	enabled := false
	sendDsPacket := false
	matchTimeSec := arena.MatchTimeSec()
	switch arena.MatchState {
	case PreMatch:
		auto = true
		enabled = false
	case StartMatch:
		arena.MatchState = AutoPeriod
		arena.MatchStartTime = time.Now()
		arena.LastMatchTimeSec = -1
		auto = true
		enabled = true
		sendDsPacket = true
		arena.AudienceDisplayScreen = "match"
		arena.AudienceDisplayNotifier.Notify(nil)
		if !arena.MuteMatchSounds {
			arena.PlaySoundNotifier.Notify("match-start")
		}
		arena.FieldTestMode = ""
		arena.Plc.ResetCounts()
	case AutoPeriod:
		auto = true
		enabled = true
		if matchTimeSec >= float64(game.MatchTiming.AutoDurationSec) {
			arena.MatchState = PausePeriod
			auto = false
			enabled = false
			sendDsPacket = true
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-end")
			}
		}
	case PausePeriod:
		auto = false
		enabled = false
		if matchTimeSec >= float64(game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec) {
			arena.MatchState = TeleopPeriod
			auto = false
			enabled = true
			sendDsPacket = true
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-resume")
			}
		}
	case TeleopPeriod:
		auto = false
		enabled = true
		if matchTimeSec >= float64(game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+
			game.MatchTiming.TeleopDurationSec-game.MatchTiming.EndgameTimeLeftSec) {
			arena.MatchState = EndgamePeriod
			sendDsPacket = false
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-endgame")
			}
		}
	case EndgamePeriod:
		auto = false
		enabled = true
		if matchTimeSec >= float64(game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+
			game.MatchTiming.TeleopDurationSec) {
			arena.MatchState = PostMatch
			auto = false
			enabled = false
			sendDsPacket = true
			go func() {
				// Leave the scores on the screen briefly at the end of the match.
				time.Sleep(time.Second * matchEndScoreDwellSec)
				arena.AudienceDisplayScreen = "blank"
				arena.AudienceDisplayNotifier.Notify(nil)
				arena.AllianceStationDisplayScreen = "logo"
				arena.AllianceStationDisplayNotifier.Notify(nil)
			}()
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-end")
			}
		}
	}

	// Send a notification if the match state has changed.
	if arena.MatchState != arena.lastMatchState {
		arena.matchStateNotifier.Notify(arena.MatchState)
	}
	arena.lastMatchState = arena.MatchState

	// Send a match tick notification if passing an integer second threshold.
	if int(matchTimeSec) != int(arena.LastMatchTimeSec) {
		arena.MatchTimeNotifier.Notify(int(matchTimeSec))
	}
	arena.LastMatchTimeSec = matchTimeSec

	// Send a packet if at a period transition point or if it's been long enough since the last one.
	if sendDsPacket || time.Since(arena.lastDsPacketTime).Seconds()*1000 >= dsPacketPeriodMs {
		arena.sendDsPacket(auto, enabled)
		arena.RobotStatusNotifier.Notify(nil)
	}

	// Handle field sensors/lights/motors.
	arena.handlePlcInput()
	arena.handlePlcOutput()
}

// Loops indefinitely to track and update the arena components.
func (arena *Arena) Run() {
	// Start other loops in goroutines.
	go arena.listenForDriverStations()
	go arena.listenForDsUdpPackets()
	go arena.monitorBandwidth()
	go arena.Plc.Run()

	for {
		arena.Update()
		time.Sleep(time.Millisecond * arenaLoopPeriodMs)
	}
}

// Calculates the red alliance score summary for the given realtime snapshot.
func (arena *Arena) RedScoreSummary() *game.ScoreSummary {
	return arena.RedRealtimeScore.CurrentScore.Summarize(arena.BlueRealtimeScore.CurrentScore.Fouls,
		arena.CurrentMatch.Type)
}

// Calculates the blue alliance score summary for the given realtime snapshot.
func (arena *Arena) BlueScoreSummary() *game.ScoreSummary {
	return arena.BlueRealtimeScore.CurrentScore.Summarize(arena.RedRealtimeScore.CurrentScore.Fouls,
		arena.CurrentMatch.Type)
}

func (arena *Arena) GetStatus() *ArenaStatus {
	return &ArenaStatus{arena.AllianceStations, arena.MatchState, arena.checkCanStartMatch() == nil,
		arena.Plc.IsHealthy, arena.Plc.GetFieldEstop()}
}

// Loads a team into an alliance station, cleaning up the previous team there if there is one.
func (arena *Arena) assignTeam(teamId int, station string) error {
	// Reject invalid station values.
	if _, ok := arena.AllianceStations[station]; !ok {
		return fmt.Errorf("Invalid alliance station '%s'.", station)
	}

	// Do nothing if the station is already assigned to the requested team.
	dsConn := arena.AllianceStations[station].DsConn
	if dsConn != nil && dsConn.TeamId == teamId {
		return nil
	}
	if dsConn != nil {
		dsConn.close()
		arena.AllianceStations[station].Team = nil
		arena.AllianceStations[station].DsConn = nil
	}

	// Leave the station empty if the team number is zero.
	if teamId == 0 {
		arena.AllianceStations[station].Team = nil
		return nil
	}

	// Load the team model. If it doesn't exist, enable anonymous operation.
	team, err := arena.Database.GetTeamById(teamId)
	if err != nil {
		return err
	}
	if team == nil {
		team = &model.Team{Id: teamId}
	}

	arena.AllianceStations[station].Team = team
	return nil
}

// Asynchronously reconfigures the networking hardware for the new set of teams.
func (arena *Arena) setupNetwork() {
	if arena.EventSettings.NetworkSecurityEnabled {
		go func() {
			err := arena.accessPoint.ConfigureTeamWifi(arena.AllianceStations["R1"].Team,
				arena.AllianceStations["R2"].Team, arena.AllianceStations["R3"].Team, arena.AllianceStations["B1"].Team,
				arena.AllianceStations["B2"].Team, arena.AllianceStations["B3"].Team)
			if err != nil {
				log.Printf("Failed to configure team WiFi: %s", err.Error())
			}
		}()
		go func() {
			err := arena.networkSwitch.ConfigureTeamEthernet(arena.AllianceStations["R1"].Team,
				arena.AllianceStations["R2"].Team, arena.AllianceStations["R3"].Team, arena.AllianceStations["B1"].Team,
				arena.AllianceStations["B2"].Team, arena.AllianceStations["B3"].Team)
			if err != nil {
				log.Printf("Failed to configure team Ethernet: %s", err.Error())
			}
		}()
	}
}

// Returns nil if the match can be started, and an error otherwise.
func (arena *Arena) checkCanStartMatch() error {
	if arena.MatchState != PreMatch {
		return fmt.Errorf("Cannot start match while there is a match still in progress or with results pending.")
	}
	for _, allianceStation := range arena.AllianceStations {
		if allianceStation.Estop {
			return fmt.Errorf("Cannot start match while an emergency stop is active.")
		}
		if !allianceStation.Bypass {
			if allianceStation.DsConn == nil || !allianceStation.DsConn.RobotLinked {
				return fmt.Errorf("Cannot start match until all robots are connected or bypassed.")
			}
		}
	}

	if arena.EventSettings.PlcAddress != "" {
		if !arena.Plc.IsHealthy {
			return fmt.Errorf("Cannot start match while PLC is not healthy.")
		}
		if arena.Plc.GetFieldEstop() {
			return fmt.Errorf("Cannot start match while field emergency stop is active.")
		}
	}

	return nil
}

func (arena *Arena) sendDsPacket(auto bool, enabled bool) {
	for _, allianceStation := range arena.AllianceStations {
		dsConn := allianceStation.DsConn
		if dsConn != nil {
			dsConn.Auto = auto
			dsConn.Enabled = enabled && !allianceStation.Estop && !allianceStation.Bypass
			dsConn.Estop = allianceStation.Estop
			err := dsConn.update(arena)
			if err != nil {
				log.Printf("Unable to send driver station packet for team %d.", allianceStation.Team.Id)
			}
		}
	}
	arena.lastDsPacketTime = time.Now()
}

// Returns the alliance station identifier for the given team, or the empty string if the team is not present
// in the current match.
func (arena *Arena) getAssignedAllianceStation(teamId int) string {
	for station, allianceStation := range arena.AllianceStations {
		if allianceStation.Team != nil && allianceStation.Team.Id == teamId {
			return station
		}
	}

	return ""
}

// Updates the score given new input information from the field PLC.
func (arena *Arena) handlePlcInput() {
	// Handle emergency stops.
	if arena.Plc.GetFieldEstop() && arena.MatchTimeSec() > 0 && !arena.matchAborted {
		arena.AbortMatch()
	}
	redEstops, blueEstops := arena.Plc.GetTeamEstops()
	arena.handleEstop("R1", redEstops[0])
	arena.handleEstop("R2", redEstops[1])
	arena.handleEstop("R3", redEstops[2])
	arena.handleEstop("B1", blueEstops[0])
	arena.handleEstop("B2", blueEstops[1])
	arena.handleEstop("B3", blueEstops[2])

	matchStartTime := arena.MatchStartTime
	currentTime := time.Now()
	if arena.MatchState == PreMatch {
		// Set a match start time in the future.
		matchStartTime = currentTime.Add(time.Second)
	}
	matchEndTime := game.GetMatchEndTime(matchStartTime)
	inGracePeriod := currentTime.Before(matchEndTime.Add(game.BoilerTeleopGracePeriodSec * time.Second))
	if arena.MatchState == PostMatch && (!inGracePeriod || arena.matchAborted) {
		// Don't do anything if we're past the end of the match, otherwise we may overwrite manual edits.
		return
	}

	redScore := &arena.RedRealtimeScore.CurrentScore
	oldRedScore := *redScore
	blueScore := &arena.BlueRealtimeScore.CurrentScore
	oldBlueScore := *blueScore

	// Handle balls.
	redLow, redHigh, blueLow, blueHigh := arena.Plc.GetBalls()
	arena.RedRealtimeScore.boiler.UpdateState(redLow, redHigh, matchStartTime, currentTime)
	redScore.AutoFuelLow = arena.RedRealtimeScore.boiler.AutoFuelLow
	redScore.AutoFuelHigh = arena.RedRealtimeScore.boiler.AutoFuelHigh
	redScore.FuelLow = arena.RedRealtimeScore.boiler.FuelLow
	redScore.FuelHigh = arena.RedRealtimeScore.boiler.FuelHigh
	arena.BlueRealtimeScore.boiler.UpdateState(blueLow, blueHigh, matchStartTime, currentTime)
	blueScore.AutoFuelLow = arena.BlueRealtimeScore.boiler.AutoFuelLow
	blueScore.AutoFuelHigh = arena.BlueRealtimeScore.boiler.AutoFuelHigh
	blueScore.FuelLow = arena.BlueRealtimeScore.boiler.FuelLow
	blueScore.FuelHigh = arena.BlueRealtimeScore.boiler.FuelHigh

	// Handle rotors.
	redRotor1, redOtherRotors, blueRotor1, blueOtherRotors := arena.Plc.GetRotors()
	arena.RedRealtimeScore.rotorSet.UpdateState(redRotor1, redOtherRotors, matchStartTime, currentTime)
	redScore.AutoRotors = arena.RedRealtimeScore.rotorSet.AutoRotors
	redScore.Rotors = arena.RedRealtimeScore.rotorSet.Rotors
	arena.BlueRealtimeScore.rotorSet.UpdateState(blueRotor1, blueOtherRotors, matchStartTime, currentTime)
	blueScore.AutoRotors = arena.BlueRealtimeScore.rotorSet.AutoRotors
	blueScore.Rotors = arena.BlueRealtimeScore.rotorSet.Rotors

	// Handle touchpads.
	redTouchpads, blueTouchpads := arena.Plc.GetTouchpads()
	for i := 0; i < 3; i++ {
		arena.RedRealtimeScore.touchpads[i].UpdateState(redTouchpads[i], matchStartTime, currentTime)
		arena.BlueRealtimeScore.touchpads[i].UpdateState(blueTouchpads[i], matchStartTime, currentTime)
	}
	redScore.Takeoffs = game.CountTouchpads(&arena.RedRealtimeScore.touchpads, currentTime)
	blueScore.Takeoffs = game.CountTouchpads(&arena.BlueRealtimeScore.touchpads, currentTime)

	if !oldRedScore.Equals(redScore) || !oldBlueScore.Equals(blueScore) {
		arena.RealtimeScoreNotifier.Notify(nil)
	}
}

// Writes light/motor commands to the field PLC.
func (arena *Arena) handlePlcOutput() {
	if arena.FieldTestMode != "" {
		// PLC output is being manually overridden.
		if arena.FieldTestMode == "flash" {
			blinkState := arena.Plc.GetCycleState(2, 0, 1)
			arena.Plc.SetTouchpadLights([3]bool{blinkState, blinkState, blinkState},
				[3]bool{blinkState, blinkState, blinkState})
		} else if arena.FieldTestMode == "cycle" {
			arena.Plc.SetTouchpadLights(
				[3]bool{arena.Plc.GetCycleState(3, 2, 1), arena.Plc.GetCycleState(3, 1, 1), arena.Plc.GetCycleState(3, 0, 1)},
				[3]bool{arena.Plc.GetCycleState(3, 0, 1), arena.Plc.GetCycleState(3, 1, 1), arena.Plc.GetCycleState(3, 2, 1)})
		} else if arena.FieldTestMode == "chase" {
			arena.Plc.SetTouchpadLights(
				[3]bool{arena.Plc.GetCycleState(12, 2, 2), arena.Plc.GetCycleState(12, 1, 2), arena.Plc.GetCycleState(12, 0, 2)},
				[3]bool{arena.Plc.GetCycleState(12, 3, 2), arena.Plc.GetCycleState(12, 4, 2), arena.Plc.GetCycleState(12, 5, 2)})
		} else if arena.FieldTestMode == "slowChase" {
			arena.Plc.SetTouchpadLights(
				[3]bool{arena.Plc.GetCycleState(6, 2, 8), arena.Plc.GetCycleState(6, 1, 8), arena.Plc.GetCycleState(6, 0, 8)},
				[3]bool{arena.Plc.GetCycleState(6, 3, 8), arena.Plc.GetCycleState(6, 4, 8), arena.Plc.GetCycleState(6, 5, 8)})
		}
		return
	}

	// Handle balls.
	matchEndTime := game.GetMatchEndTime(arena.MatchStartTime)
	inGracePeriod := time.Now().Before(matchEndTime.Add(game.BoilerTeleopGracePeriodSec * time.Second))
	if arena.MatchTimeSec() > 0 || arena.MatchState == PostMatch && !arena.matchAborted && inGracePeriod {
		arena.Plc.SetBoilerMotors(true)
	} else {
		arena.Plc.SetBoilerMotors(false)
	}

	// Handle rotors.
	redScore := &arena.RedRealtimeScore.CurrentScore
	blueScore := &arena.BlueRealtimeScore.CurrentScore
	if arena.MatchTimeSec() > 0 {
		arena.Plc.SetRotorMotors(redScore.AutoRotors+redScore.Rotors, blueScore.AutoRotors+blueScore.Rotors)
	} else {
		arena.Plc.SetRotorMotors(0, 0)
	}
	arena.Plc.SetRotorLights(redScore.AutoRotors, blueScore.AutoRotors)

	// Handle touchpads.
	var redTouchpads, blueTouchpads [3]bool
	currentTime := time.Now()
	blinkStopTime := matchEndTime.Add(-time.Duration(game.MatchTiming.EndgameTimeLeftSec-2) * time.Second)
	blinkState := arena.Plc.GetCycleState(2, 0, 1)
	if arena.MatchState == EndgamePeriod && currentTime.Before(blinkStopTime) {
		// Blink the touchpads at the endgame start point.
		for i := 0; i < 3; i++ {
			redTouchpads[i] = blinkState
			blueTouchpads[i] = blinkState
		}
	} else {
		for i := 0; i < 3; i++ {
			redState := arena.RedRealtimeScore.touchpads[i].GetState(currentTime)
			redTouchpads[i] = redState == game.Held || redState == game.Triggered && blinkState
			blueState := arena.BlueRealtimeScore.touchpads[i].GetState(currentTime)
			blueTouchpads[i] = blueState == game.Held || blueState == game.Triggered && blinkState
		}
	}
	arena.Plc.SetTouchpadLights(redTouchpads, blueTouchpads)
}

func (arena *Arena) handleEstop(station string, state bool) {
	allianceStation := arena.AllianceStations[station]
	if state {
		allianceStation.Estop = true
	} else if arena.MatchTimeSec() == 0 {
		// Don't reset the e-stop while a match is in progress.
		allianceStation.Estop = false
	}
}
