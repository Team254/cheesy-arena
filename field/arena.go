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
	WarmupPeriod  = 2
	AutoPeriod    = 3
	PausePeriod   = 4
	TeleopPeriod  = 5
	EndgamePeriod = 6
	PostMatch     = 7
)

type Arena struct {
	Database                       *model.Database
	EventSettings                  *model.EventSettings
	accessPoint                    *AccessPoint
	networkSwitch                  *NetworkSwitch
	Plc                            Plc
	TbaClient                      *partner.TbaClient
	StemTvClient                   *partner.StemTvClient
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
	scale                          *game.Seesaw
	redSwitch                      *game.Seesaw
	blueSwitch                     *game.Seesaw
	redVault                       *game.Vault
	blueVault                      *game.Vault
}

type ArenaStatus struct {
	AllianceStations map[string]*AllianceStation
	MatchState       int
	CanStartMatch    bool
	PlcIsHealthy     bool
	FieldEstop       bool
	GameSpecificData string
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
	arena.accessPoint = NewAccessPoint(settings.ApAddress, settings.ApUsername, settings.ApPassword,
		settings.ApTeamChannel, settings.ApAdminChannel, settings.ApAdminWpaKey)
	arena.networkSwitch = NewNetworkSwitch(settings.SwitchAddress, settings.SwitchPassword)
	arena.Plc.SetAddress(settings.PlcAddress)
	arena.TbaClient = partner.NewTbaClient(settings.TbaEventCode, settings.TbaSecretId, settings.TbaSecret)
	arena.StemTvClient = partner.NewStemTvClient(settings.StemTvEventCode)

	err = arena.accessPoint.ConfigureAdminWifi()
	if err != nil {
		return nil
	}

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
	arena.scale = new(game.Seesaw)
	arena.redSwitch = new(game.Seesaw)
	arena.blueSwitch = new(game.Seesaw)
	arena.redVault = new(game.Vault)
	arena.blueVault = new(game.Vault)
	game.ResetPowerUps()

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

	if arena.CurrentMatch.Type != "test" {
		arena.Database.SaveMatch(arena.CurrentMatch)
	}
	return nil
}

// Starts the match if all conditions are met.
func (arena *Arena) StartMatch() error {
	err := arena.checkCanStartMatch()
	if err == nil {
		// Generate game-specific data or allow manual input for test matches.
		if arena.CurrentMatch.Type != "test" || !game.IsValidGameSpecificData(arena.CurrentMatch.GameSpecificData) {
			arena.CurrentMatch.GameSpecificData = game.GenerateGameSpecificData()
		}

		// Save the match start time and game-specifc data to the database for posterity.
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
		arena.MatchState = WarmupPeriod
		arena.MatchStartTime = time.Now()
		arena.LastMatchTimeSec = -1
		auto = true
		enabled = false
		arena.AudienceDisplayScreen = "match"
		arena.AudienceDisplayNotifier.Notify(nil)
		arena.FieldTestMode = ""
		arena.Plc.ResetCounts()
		arena.sendGameSpecificDataPacket()
		if !arena.MuteMatchSounds {
			arena.PlaySoundNotifier.Notify("match-warmup")
		}
	case WarmupPeriod:
		auto = true
		enabled = false
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec) {
			arena.MatchState = AutoPeriod
			auto = true
			enabled = true
			sendDsPacket = true
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-start")
			}
		}
	case AutoPeriod:
		auto = true
		enabled = true
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec) {
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
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+
			game.MatchTiming.PauseDurationSec) {
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
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+
			game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec-game.MatchTiming.EndgameTimeLeftSec) {
			arena.MatchState = EndgamePeriod
			sendDsPacket = false
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.Notify("match-endgame")
			}
		}
	case EndgamePeriod:
		auto = false
		enabled = true
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec+game.MatchTiming.AutoDurationSec+
			game.MatchTiming.PauseDurationSec+game.MatchTiming.TeleopDurationSec) {
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
	return arena.RedRealtimeScore.CurrentScore.Summarize(arena.BlueRealtimeScore.CurrentScore.Fouls)
}

// Calculates the blue alliance score summary for the given realtime snapshot.
func (arena *Arena) BlueScoreSummary() *game.ScoreSummary {
	return arena.BlueRealtimeScore.CurrentScore.Summarize(arena.RedRealtimeScore.CurrentScore.Fouls)
}

func (arena *Arena) GetStatus() *ArenaStatus {
	return &ArenaStatus{arena.AllianceStations, arena.MatchState, arena.checkCanStartMatch() == nil,
		arena.Plc.IsHealthy, arena.Plc.GetFieldEstop(), arena.CurrentMatch.GameSpecificData}
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

func (arena *Arena) sendGameSpecificDataPacket() {
	for _, allianceStation := range arena.AllianceStations {
		dsConn := allianceStation.DsConn
		if dsConn != nil {
			err := dsConn.sendGameSpecificDataPacket(arena.CurrentMatch.GameSpecificData)
			if err != nil {
				log.Printf("Error sending game-specific data packet to Team %d: %v", dsConn.TeamId, err)
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

	if arena.MatchState == PreMatch || arena.MatchState == PostMatch {
		// Don't do anything if we're outside the match, otherwise we may overwrite manual edits.
		return
	}
	matchStartTime := arena.MatchStartTime
	currentTime := time.Now()
	teleopStartTime := game.GetTeleopStartTime(matchStartTime)

	redScore := &arena.RedRealtimeScore.CurrentScore
	oldRedScore := *redScore
	blueScore := &arena.BlueRealtimeScore.CurrentScore
	oldBlueScore := *blueScore

	// Handle scale and switch ownership.
	scale, redSwitch, blueSwitch := arena.Plc.GetScaleAndSwitches()
	arena.scale.UpdateState(scale, currentTime)
	arena.redSwitch.UpdateState(redSwitch, currentTime)
	arena.blueSwitch.UpdateState(blueSwitch, currentTime)
	if arena.MatchState == AutoPeriod {
		redScore.AutoOwnershipPoints = 2 * int(arena.redSwitch.GetRedSeconds(matchStartTime, currentTime)+
			arena.scale.GetRedSeconds(matchStartTime, currentTime))
		blueScore.AutoOwnershipPoints = 2 * int(arena.blueSwitch.GetBlueSeconds(matchStartTime, currentTime)+
			arena.scale.GetBlueSeconds(matchStartTime, currentTime))
	} else {
		redScore.TeleopOwnershipPoints = int(arena.redSwitch.GetRedSeconds(teleopStartTime, currentTime) +
			arena.scale.GetRedSeconds(teleopStartTime, currentTime))
		blueScore.TeleopOwnershipPoints = int(arena.blueSwitch.GetBlueSeconds(teleopStartTime, currentTime) +
			arena.scale.GetBlueSeconds(teleopStartTime, currentTime))
	}

	// Handle vaults.
	redForceCubes, redLevitateCubes, redBoostCubes, blueForceCubes, blueLevitateCubes, blueBoostCubes :=
		arena.Plc.GetVaults()
	arena.redVault.UpdateCubes(redForceCubes, redLevitateCubes, redBoostCubes)
	arena.blueVault.UpdateCubes(blueForceCubes, blueLevitateCubes, blueBoostCubes)
	redForce, redLevitate, redBoost, blueForce, blueLevitate, blueBoost := arena.Plc.GetPowerUpButtons()
	arena.redVault.UpdateButtons(redForce, redLevitate, redBoost, currentTime)
	arena.blueVault.UpdateButtons(blueForce, blueLevitate, blueBoost, currentTime)
	redScore.VaultCubes = arena.redVault.GetNumCubes()
	redScore.Levitate = arena.redVault.LevitatePlayed
	blueScore.VaultCubes = arena.blueVault.GetNumCubes()
	blueScore.Levitate = arena.blueVault.LevitatePlayed

	if !oldRedScore.Equals(redScore) || !oldBlueScore.Equals(blueScore) {
		arena.RealtimeScoreNotifier.Notify(nil)
	}
}

// Writes light/motor commands to the field PLC.
func (arena *Arena) handlePlcOutput() {
	if arena.FieldTestMode != "" {
		// PLC output is being manually overridden.
		// TODO(patrick): Update for 2018.
		/*
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
		*/
	}
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
