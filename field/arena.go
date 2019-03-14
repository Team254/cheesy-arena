// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for controlling the arena and match play.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/led"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/plc"
	"github.com/Team254/cheesy-arena/vaultled"
	"log"
	"math/rand"
	"time"
)

const (
	arenaLoopPeriodMs     = 10
	dsPacketPeriodMs      = 250
	matchEndScoreDwellSec = 3
	postTimeoutSec        = 4
)

// Progression of match states.
type MatchState int

const (
	PreMatch MatchState = iota
	StartMatch
	WarmupPeriod
	AutoPeriod
	PausePeriod
	TeleopPeriod
	EndgamePeriod
	PostMatch
	TimeoutActive
	PostTimeout
)

type Arena struct {
	Database         *model.Database
	EventSettings    *model.EventSettings
	accessPoint      network.AccessPoint
	networkSwitch    *network.Switch
	Plc              plc.Plc
	TbaClient        *partner.TbaClient
	AllianceStations map[string]*AllianceStation
	Displays         map[string]*Display
	ArenaNotifiers
	MatchState
	lastMatchState             MatchState
	CurrentMatch               *model.Match
	MatchStartTime             time.Time
	LastMatchTimeSec           float64
	RedRealtimeScore           *RealtimeScore
	BlueRealtimeScore          *RealtimeScore
	lastDsPacketTime           time.Time
	FieldVolunteers            bool
	FieldReset                 bool
	AudienceDisplayMode        string
	SavedMatch                 *model.Match
	SavedMatchResult           *model.MatchResult
	AllianceStationDisplayMode string
	AllianceSelectionAlliances [][]model.AllianceTeam
	LowerThird                 *model.LowerThird
	MuteMatchSounds            bool
	matchAborted               bool
	Scale                      *game.Seesaw
	RedSwitch                  *game.Seesaw
	BlueSwitch                 *game.Seesaw
	RedVault                   *game.Vault
	BlueVault                  *game.Vault
	ScaleLeds                  led.Controller
	RedSwitchLeds              led.Controller
	BlueSwitchLeds             led.Controller
	RedVaultLeds               vaultled.Controller
	BlueVaultLeds              vaultled.Controller
	warmupLedMode              led.Mode
	lastRedAllianceReady       bool
	lastBlueAllianceReady      bool
}

type AllianceStation struct {
	DsConn *DriverStationConnection
	Astop  bool
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

	arena.Displays = make(map[string]*Display)

	arena.configureNotifiers()

	// Load empty match as current.
	arena.MatchState = PreMatch
	arena.LoadTestMatch()
	arena.LastMatchTimeSec = 0
	arena.lastMatchState = -1

	// Initialize display parameters.
	arena.AudienceDisplayMode = "blank"
	arena.SavedMatch = &model.Match{}
	arena.SavedMatchResult = model.NewMatchResult()
	arena.AllianceStationDisplayMode = "match"

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
	arena.accessPoint.SetSettings(settings.ApAddress, settings.ApUsername, settings.ApPassword,
		settings.ApTeamChannel, settings.ApAdminChannel, settings.ApAdminWpaKey, settings.NetworkSecurityEnabled)
	arena.networkSwitch = network.NewSwitch(settings.SwitchAddress, settings.SwitchPassword)
	arena.Plc.SetAddress(settings.PlcAddress)
	arena.TbaClient = partner.NewTbaClient(settings.TbaEventCode, settings.TbaSecretId, settings.TbaSecret)

	if arena.EventSettings.NetworkSecurityEnabled && arena.MatchState == PreMatch {
		if err = arena.accessPoint.ConfigureAdminWifi(); err != nil {
			log.Printf("Failed to configure admin WiFi: %s", err.Error())
		}
	}

	// Initialize LEDs.
	if err = arena.ScaleLeds.SetAddress(settings.ScaleLedAddress); err != nil {
		return err
	}
	if err = arena.RedSwitchLeds.SetAddress(settings.RedSwitchLedAddress); err != nil {
		return err
	}
	if err = arena.BlueSwitchLeds.SetAddress(settings.BlueSwitchLedAddress); err != nil {
		return err
	}
	if err = arena.RedVaultLeds.SetAddress(settings.RedVaultLedAddress); err != nil {
		return err
	}
	if err = arena.BlueVaultLeds.SetAddress(settings.BlueVaultLedAddress); err != nil {
		return err
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
	arena.FieldVolunteers = false
	arena.FieldReset = false
	arena.Scale = &game.Seesaw{Kind: game.NeitherAlliance}
	arena.RedSwitch = &game.Seesaw{Kind: game.RedAlliance}
	arena.BlueSwitch = &game.Seesaw{Kind: game.BlueAlliance}
	arena.RedVault = &game.Vault{Alliance: game.RedAlliance}
	arena.BlueVault = &game.Vault{Alliance: game.BlueAlliance}
	game.ResetPowerUps()

	// Set a consistent initial value for field element sidedness.
	arena.Scale.NearIsRed = true
	arena.RedSwitch.NearIsRed = true
	arena.BlueSwitch.NearIsRed = true
	arena.ScaleLeds.SetSidedness(true)
	arena.RedSwitchLeds.SetSidedness(true)
	arena.BlueSwitchLeds.SetSidedness(true)

	// Notify any listeners about the new match.
	arena.MatchLoadNotifier.Notify()
	arena.RealtimeScoreNotifier.Notify()
	arena.AllianceStationDisplayMode = "match"
	arena.AllianceStationDisplayModeNotifier.Notify()

	// Set the initial state of the lights.
	arena.ScaleLeds.SetMode(led.GreenMode, led.GreenMode)
	arena.RedSwitchLeds.SetMode(led.RedMode, led.RedMode)
	arena.BlueSwitchLeds.SetMode(led.BlueMode, led.BlueMode)
	arena.RedVaultLeds.SetAllModes(vaultled.OffMode)
	arena.BlueVaultLeds.SetAllModes(vaultled.OffMode)
	arena.lastRedAllianceReady = false
	arena.lastBlueAllianceReady = false

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
			if err = arena.LoadMatch(&match); err != nil {
				return err
			}
			return nil
		}
	}

	// There are no matches left in the series; just load a test match.
	return arena.LoadTestMatch()
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
	arena.MatchLoadNotifier.Notify()

	if arena.CurrentMatch.Type != "test" {
		arena.Database.SaveMatch(arena.CurrentMatch)
	}
	return nil
}

// StartMatch starts the match if all conditions are met.
func (arena *Arena) StartMatch() error {
	err := arena.checkCanStartMatch()
	if err == nil {
		// Generate game-specific data or allow manual input for test matches.
		if arena.CurrentMatch.Type != "test" || !game.IsValidGameSpecificData(arena.CurrentMatch.GameSpecificData) {
			arena.CurrentMatch.GameSpecificData = game.GenerateGameSpecificData()
		}

		// Configure the field elements with the game-specific data.
		switchNearIsRed := arena.CurrentMatch.GameSpecificData[0] == 'L'
		scaleNearIsRed := arena.CurrentMatch.GameSpecificData[1] == 'L'
		arena.Scale.NearIsRed = scaleNearIsRed
		arena.RedSwitch.NearIsRed = switchNearIsRed
		arena.BlueSwitch.NearIsRed = switchNearIsRed
		arena.ScaleLeds.SetSidedness(scaleNearIsRed)
		arena.RedSwitchLeds.SetSidedness(switchNearIsRed)
		arena.BlueSwitchLeds.SetSidedness(switchNearIsRed)

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

			// Save the teams that have successfully connected to the field.
			if allianceStation.Team != nil && !allianceStation.Team.HasConnected && allianceStation.DsConn != nil &&
				allianceStation.DsConn.RobotLinked {
				allianceStation.Team.HasConnected = true
				arena.Database.SaveTeam(allianceStation.Team)
			}
		}

		arena.MatchState = StartMatch
	}
	return err
}

// Kills the current match or timeout if it is underway.
func (arena *Arena) AbortMatch() error {
	if arena.MatchState == PreMatch || arena.MatchState == PostMatch || arena.MatchState == PostTimeout {
		return fmt.Errorf("Cannot abort match when it is not in progress.")
	}

	if arena.MatchState == TimeoutActive {
		// Handle by advancing the timeout clock to the end and letting the regular logic deal with it.
		arena.MatchStartTime = time.Now().Add(-time.Second * time.Duration(game.MatchTiming.TimeoutDurationSec))
		return nil
	}

	if !arena.MuteMatchSounds && arena.MatchState != WarmupPeriod {
		arena.PlaySoundNotifier.NotifyWithMessage("match-abort")
	}
	arena.MatchState = PostMatch
	arena.matchAborted = true
	arena.AudienceDisplayMode = "blank"
	arena.AudienceDisplayModeNotifier.Notify()
	arena.AllianceStationDisplayMode = "logo"
	arena.AllianceStationDisplayModeNotifier.Notify()
	return nil
}

// ResetMatch clears out the match and resets the arena state unless there is a match underway.
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

// StartTimeout starts a timeout of the given duration.
func (arena *Arena) StartTimeout(durationSec int) error {
	if arena.MatchState != PreMatch {
		return fmt.Errorf("Cannot start timeout while there is a match still in progress or with results pending.")
	}

	game.MatchTiming.TimeoutDurationSec = durationSec
	arena.MatchTimingNotifier.Notify()
	arena.MatchState = TimeoutActive
	arena.MatchStartTime = time.Now()
	arena.LastMatchTimeSec = -1
	arena.AudienceDisplayMode = "timeout"
	arena.AudienceDisplayModeNotifier.Notify()

	return nil
}

// MatchTimeSec returns the fractional number of seconds since the start of the match.
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
		arena.AudienceDisplayMode = "match"
		arena.AudienceDisplayModeNotifier.Notify()
		arena.AllianceStationDisplayMode = "match"
		arena.AllianceStationDisplayModeNotifier.Notify()
		arena.sendGameSpecificDataPacket()
		if !arena.MuteMatchSounds {
			arena.PlaySoundNotifier.NotifyWithMessage("match-warmup")
		}
		// Pick an LED warmup mode at random to keep things interesting.
		allWarmupModes := []led.Mode{led.WarmupMode, led.Warmup2Mode, led.Warmup3Mode, led.Warmup4Mode}
		arena.warmupLedMode = allWarmupModes[rand.Intn(len(allWarmupModes))]
	case WarmupPeriod:
		auto = true
		enabled = false
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec) {
			arena.MatchState = AutoPeriod
			auto = true
			enabled = true
			sendDsPacket = true
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.NotifyWithMessage("match-start")
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
				arena.PlaySoundNotifier.NotifyWithMessage("match-end")
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
				arena.PlaySoundNotifier.NotifyWithMessage("match-resume")
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
				arena.PlaySoundNotifier.NotifyWithMessage("match-endgame")
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
				arena.AudienceDisplayMode = "blank"
				arena.AudienceDisplayModeNotifier.Notify()
				arena.AllianceStationDisplayMode = "logo"
				arena.AllianceStationDisplayModeNotifier.Notify()
			}()
			if !arena.MuteMatchSounds {
				arena.PlaySoundNotifier.NotifyWithMessage("match-end")
			}
		}
	case TimeoutActive:
		if matchTimeSec >= float64(game.MatchTiming.TimeoutDurationSec) {
			arena.MatchState = PostTimeout
			arena.PlaySoundNotifier.NotifyWithMessage("match-end")
			go func() {
				// Leave the timer on the screen briefly at the end of the timeout period.
				time.Sleep(time.Second * matchEndScoreDwellSec)
				arena.AudienceDisplayMode = "blank"
				arena.AudienceDisplayModeNotifier.Notify()
			}()
		}
	case PostTimeout:
		if matchTimeSec >= float64(game.MatchTiming.TimeoutDurationSec+postTimeoutSec) {
			arena.MatchState = PreMatch
		}
	}

	// Send a match tick notification if passing an integer second threshold or if the match state changed.
	if int(matchTimeSec) != int(arena.LastMatchTimeSec) || arena.MatchState != arena.lastMatchState {
		arena.MatchTimeNotifier.Notify()
	}
	arena.LastMatchTimeSec = matchTimeSec
	arena.lastMatchState = arena.MatchState

	// Send a packet if at a period transition point or if it's been long enough since the last one.
	if sendDsPacket || time.Since(arena.lastDsPacketTime).Seconds()*1000 >= dsPacketPeriodMs {
		arena.sendDsPacket(auto, enabled)
		arena.ArenaStatusNotifier.Notify()
	}

	// Handle field sensors/lights/motors.
	arena.handlePlcInput()
	arena.handleLeds()
}

// Loops indefinitely to track and update the arena components.
func (arena *Arena) Run() {
	// Start other loops in goroutines.
	go arena.listenForDriverStations()
	go arena.listenForDsUdpPackets()
	go arena.accessPoint.Run()
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
		err := arena.accessPoint.ConfigureTeamWifi(arena.AllianceStations["R1"].Team,
			arena.AllianceStations["R2"].Team, arena.AllianceStations["R3"].Team, arena.AllianceStations["B1"].Team,
			arena.AllianceStations["B2"].Team, arena.AllianceStations["B3"].Team)
		if err != nil {
			log.Printf("Failed to configure team WiFi: %s", err.Error())
		}
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

	err := arena.checkAllianceStationsReady("R1", "R2", "R3", "B1", "B2", "B3")
	if err != nil {
		return err
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

func (arena *Arena) checkAllianceStationsReady(stations ...string) error {
	for _, station := range stations {
		allianceStation := arena.AllianceStations[station]
		if allianceStation.Estop {
			return fmt.Errorf("Cannot start match while an emergency stop is active.")
		}
		if !allianceStation.Bypass {
			if allianceStation.DsConn == nil || !allianceStation.DsConn.RobotLinked {
				return fmt.Errorf("Cannot start match until all robots are connected or bypassed.")
			}
		}
	}

	return nil
}

func (arena *Arena) sendDsPacket(auto bool, enabled bool) {
	for _, allianceStation := range arena.AllianceStations {
		dsConn := allianceStation.DsConn
		if dsConn != nil {
			dsConn.Auto = auto
			dsConn.Enabled = enabled && !allianceStation.Estop && !allianceStation.Astop && !allianceStation.Bypass
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

	if arena.MatchState == PreMatch || arena.MatchState == PostMatch || arena.MatchState == TimeoutActive ||
		arena.MatchState == PostTimeout {
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
	ownershipChanged := arena.Scale.UpdateState(scale, currentTime)
	ownershipChanged = arena.RedSwitch.UpdateState(redSwitch, currentTime) || ownershipChanged
	ownershipChanged = arena.BlueSwitch.UpdateState(blueSwitch, currentTime) || ownershipChanged
	if arena.MatchState == AutoPeriod {
		redScore.AutoScaleOwnershipSec, _ = arena.Scale.GetRedSeconds(matchStartTime, currentTime)
		redScore.AutoSwitchOwnershipSec, _ = arena.RedSwitch.GetRedSeconds(matchStartTime, currentTime)
		blueScore.AutoScaleOwnershipSec, _ = arena.Scale.GetBlueSeconds(matchStartTime, currentTime)
		blueScore.AutoSwitchOwnershipSec, _ = arena.BlueSwitch.GetBlueSeconds(matchStartTime, currentTime)
		redScore.AutoEndSwitchOwnership = arena.RedSwitch.GetOwnedBy() == game.RedAlliance
		blueScore.AutoEndSwitchOwnership = arena.BlueSwitch.GetOwnedBy() == game.BlueAlliance
	} else {
		redScore.TeleopScaleOwnershipSec, redScore.TeleopScaleBoostSec =
			arena.Scale.GetRedSeconds(teleopStartTime, currentTime)
		redScore.TeleopSwitchOwnershipSec, redScore.TeleopSwitchBoostSec =
			arena.RedSwitch.GetRedSeconds(teleopStartTime, currentTime)
		blueScore.TeleopScaleOwnershipSec, blueScore.TeleopScaleBoostSec =
			arena.Scale.GetBlueSeconds(teleopStartTime, currentTime)
		blueScore.TeleopSwitchOwnershipSec, blueScore.TeleopSwitchBoostSec =
			arena.BlueSwitch.GetBlueSeconds(teleopStartTime, currentTime)
	}

	// Handle vaults.
	redForceDistance, redLevitateDistance, redBoostDistance, blueForceDistance, blueLevitateDistance,
		blueBoostDistance := arena.Plc.GetVaults()
	arena.RedVault.UpdateCubes(redForceDistance, redLevitateDistance, redBoostDistance)
	arena.BlueVault.UpdateCubes(blueForceDistance, blueLevitateDistance, blueBoostDistance)
	redForce, redLevitate, redBoost, blueForce, blueLevitate, blueBoost := arena.Plc.GetPowerUpButtons()
	arena.RedVault.UpdateButtons(redForce, redLevitate, redBoost, currentTime)
	arena.BlueVault.UpdateButtons(blueForce, blueLevitate, blueBoost, currentTime)
	redScore.ForceCubes, redScore.ForceCubesPlayed = arena.RedVault.ForceCubes, arena.RedVault.ForceCubesPlayed
	redScore.LevitateCubes, redScore.LevitatePlayed = arena.RedVault.LevitateCubes, arena.RedVault.LevitatePlayed
	redScore.BoostCubes, redScore.BoostCubesPlayed = arena.RedVault.BoostCubes, arena.RedVault.BoostCubesPlayed
	blueScore.ForceCubes, blueScore.ForceCubesPlayed = arena.BlueVault.ForceCubes, arena.BlueVault.ForceCubesPlayed
	blueScore.LevitateCubes, blueScore.LevitatePlayed = arena.BlueVault.LevitateCubes, arena.BlueVault.LevitatePlayed
	blueScore.BoostCubes, blueScore.BoostCubesPlayed = arena.BlueVault.BoostCubes, arena.BlueVault.BoostCubesPlayed

	// Check if a power up has been newly played and trigger the accompanying sound effect if so.
	newRedPowerUp := arena.RedVault.CheckForNewlyPlayedPowerUp()
	if newRedPowerUp != "" && !arena.MuteMatchSounds {
		arena.PlaySoundNotifier.NotifyWithMessage("match-" + newRedPowerUp)
	}
	newBluePowerUp := arena.BlueVault.CheckForNewlyPlayedPowerUp()
	if newBluePowerUp != "" && !arena.MuteMatchSounds {
		arena.PlaySoundNotifier.NotifyWithMessage("match-" + newBluePowerUp)
	}

	if !oldRedScore.Equals(redScore) || !oldBlueScore.Equals(blueScore) || ownershipChanged {
		arena.RealtimeScoreNotifier.Notify()
	}
}

func (arena *Arena) handleLeds() {
	switch arena.MatchState {
	case PreMatch:
		fallthrough
	case TimeoutActive:
		fallthrough
	case PostTimeout:
		// Set the stack light state -- blinking green if ready, or solid alliance color(s) if not.
		redAllianceReady := arena.checkAllianceStationsReady("R1", "R2", "R3") == nil
		blueAllianceReady := arena.checkAllianceStationsReady("B1", "B2", "B3") == nil
		greenStackLight := redAllianceReady && blueAllianceReady && arena.Plc.GetCycleState(2, 0, 2)
		arena.Plc.SetStackLights(!redAllianceReady, !blueAllianceReady, greenStackLight)
		arena.Plc.SetStackBuzzer(redAllianceReady && blueAllianceReady)

		// Turn off scale and each alliance switch if all teams become ready.
		if redAllianceReady && blueAllianceReady && !(arena.lastRedAllianceReady && arena.lastBlueAllianceReady) {
			arena.ScaleLeds.SetMode(led.OffMode, led.OffMode)
		} else if !(redAllianceReady && blueAllianceReady) && arena.lastRedAllianceReady &&
			arena.lastBlueAllianceReady {
			arena.ScaleLeds.SetMode(led.GreenMode, led.GreenMode)
		}
		if redAllianceReady && !arena.lastRedAllianceReady {
			arena.RedSwitchLeds.SetMode(led.OffMode, led.OffMode)
		} else if !redAllianceReady && arena.lastRedAllianceReady {
			arena.RedSwitchLeds.SetMode(led.RedMode, led.RedMode)
		}
		arena.lastRedAllianceReady = redAllianceReady
		if blueAllianceReady && !arena.lastBlueAllianceReady {
			arena.BlueSwitchLeds.SetMode(led.OffMode, led.OffMode)
		} else if !blueAllianceReady && arena.lastBlueAllianceReady {
			arena.BlueSwitchLeds.SetMode(led.BlueMode, led.BlueMode)
		}
		arena.lastBlueAllianceReady = blueAllianceReady

	case WarmupPeriod:
		arena.Plc.SetStackLights(false, false, true)
		arena.ScaleLeds.SetMode(arena.warmupLedMode, arena.warmupLedMode)
		arena.RedSwitchLeds.SetMode(arena.warmupLedMode, arena.warmupLedMode)
		arena.BlueSwitchLeds.SetMode(arena.warmupLedMode, arena.warmupLedMode)
	case AutoPeriod:
		fallthrough
	case TeleopPeriod:
		fallthrough
	case EndgamePeriod:
		handleSeesawTeleopLeds(arena.Scale, &arena.ScaleLeds)
		handleSeesawTeleopLeds(arena.RedSwitch, &arena.RedSwitchLeds)
		handleSeesawTeleopLeds(arena.BlueSwitch, &arena.BlueSwitchLeds)
		handleVaultTeleopLeds(arena.RedVault, &arena.RedVaultLeds)
		handleVaultTeleopLeds(arena.BlueVault, &arena.BlueVaultLeds)
	case PausePeriod:
		arena.ScaleLeds.SetMode(led.OffMode, led.OffMode)
		arena.RedSwitchLeds.SetMode(led.OffMode, led.OffMode)
		arena.BlueSwitchLeds.SetMode(led.OffMode, led.OffMode)
	case PostMatch:
		arena.Plc.SetStackLights(false, false, false)
		mode := led.FadeSingleMode
		if arena.FieldReset {
			mode = led.GreenMode
		} else if arena.FieldVolunteers {
			mode = led.PurpleMode
		}
		arena.ScaleLeds.SetMode(mode, mode)
		arena.RedSwitchLeds.SetMode(mode, mode)
		arena.BlueSwitchLeds.SetMode(mode, mode)
		arena.RedVaultLeds.SetAllModes(vaultled.OffMode)
		arena.BlueVaultLeds.SetAllModes(vaultled.OffMode)
	}

	arena.ScaleLeds.Update()
	arena.RedSwitchLeds.Update()
	arena.BlueSwitchLeds.Update()
	arena.RedVaultLeds.Update()
	arena.BlueVaultLeds.Update()
}

func handleSeesawTeleopLeds(seesaw *game.Seesaw, leds *led.Controller) {
	// Assume the simplest mode to start and consider others in order of increasing complexity.
	redMode := led.NotOwnedMode
	blueMode := led.NotOwnedMode

	// Upgrade the mode to ownership based on the physical state of the switch or scale.
	if seesaw.GetOwnedBy() == game.RedAlliance && seesaw.Kind != game.BlueAlliance {
		redMode = led.OwnedMode
	} else if seesaw.GetOwnedBy() == game.BlueAlliance && seesaw.Kind != game.RedAlliance {
		blueMode = led.OwnedMode
	}

	// Upgrade the mode if there is an applicable power up.
	powerUp := game.GetActivePowerUp(time.Now())
	if powerUp != nil && (seesaw.Kind == game.NeitherAlliance && powerUp.Level >= 2 ||
		seesaw.Kind == powerUp.Alliance && (powerUp.Level == 1 || powerUp.Level == 3)) {
		if powerUp.Effect == game.Boost {
			if powerUp.Alliance == game.RedAlliance {
				redMode = led.BoostMode
			} else {
				blueMode = led.BoostMode
			}
		} else {
			if powerUp.Alliance == game.RedAlliance {
				redMode = led.ForceMode
			} else {
				blueMode = led.ForceMode
			}
		}
	}

	if seesaw.NearIsRed {
		leds.SetMode(redMode, blueMode)
	} else {
		leds.SetMode(blueMode, redMode)
	}
}

func handleVaultTeleopLeds(vault *game.Vault, leds *vaultled.Controller) {
	playedMode := vaultled.RedPlayedMode
	if vault.Alliance == game.BlueAlliance {
		playedMode = vaultled.BluePlayedMode
	}
	cubesModeMap := map[int]vaultled.Mode{0: vaultled.OffMode, 1: vaultled.OneCubeMode, 2: vaultled.TwoCubeMode,
		3: vaultled.ThreeCubeMode}

	if vault.ForcePowerUp != nil {
		leds.SetForceMode(playedMode)
	} else {
		leds.SetForceMode(cubesModeMap[vault.ForceCubes])
	}

	if vault.LevitatePlayed {
		leds.SetLevitateMode(playedMode)
	} else {
		leds.SetLevitateMode(cubesModeMap[vault.LevitateCubes])
	}

	if vault.BoostPowerUp != nil {
		leds.SetBoostMode(playedMode)
	} else {
		leds.SetBoostMode(cubesModeMap[vault.BoostCubes])
	}
}

func (arena *Arena) handleEstop(station string, state bool) {
	allianceStation := arena.AllianceStations[station]
	if state {
		if arena.MatchState == AutoPeriod {
			allianceStation.Astop = true
		} else {
			allianceStation.Estop = true
		}
	} else {
		if arena.MatchState != AutoPeriod {
			allianceStation.Astop = false
		}
		if arena.MatchTimeSec() == 0 {
			// Don't reset the e-stop while a match is in progress.
			allianceStation.Estop = false
		}
	}
}
