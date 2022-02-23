// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for controlling the arena and match play.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/plc"
	"log"
	"math/rand"
	"time"
)

const (
	arenaLoopPeriodMs        = 10
	dsPacketPeriodMs         = 250
	periodicTaskPeriodSec    = 30
	matchEndScoreDwellSec    = 3
	postTimeoutSec           = 4
	preLoadNextMatchDelaySec = 5
	earlyLateThresholdMin    = 2.5
	MaxMatchGapMin           = 20
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
	PostMatch
	TimeoutActive
	PostTimeout
)

type Arena struct {
	Database         *model.Database
	EventSettings    *model.EventSettings
	accessPoint      network.AccessPoint
	accessPoint2     network.AccessPoint
	networkSwitch    *network.Switch
	Plc              plc.Plc
	TbaClient        *partner.TbaClient
	AllianceStations map[string]*AllianceStation
	Displays         map[string]*Display
	ScoringPanelRegistry
	ArenaNotifiers
	MatchState
	lastMatchState             MatchState
	CurrentMatch               *model.Match
	MatchStartTime             time.Time
	LastMatchTimeSec           float64
	RedRealtimeScore           *RealtimeScore
	BlueRealtimeScore          *RealtimeScore
	lastDsPacketTime           time.Time
	lastPeriodicTaskTime       time.Time
	EventStatus                EventStatus
	FieldVolunteers            bool
	FieldReset                 bool
	AudienceDisplayMode        string
	SavedMatch                 *model.Match
	SavedMatchResult           *model.MatchResult
	SavedRankings              game.Rankings
	AllianceStationDisplayMode string
	AllianceSelectionAlliances [][]model.AllianceTeam
	LowerThird                 *model.LowerThird
	ShowLowerThird             bool
	MuteMatchSounds            bool
	matchAborted               bool
	soundsPlayed               map[*game.MatchSound]struct{}
}

type AllianceStation struct {
	DsConn   *DriverStationConnection
	Ethernet bool
	Astop    bool
	Estop    bool
	Bypass   bool
	Team     *model.Team
}

// Creates the arena and sets it to its initial state.
func NewArena(dbPath string) (*Arena, error) {
	arena := new(Arena)
	arena.configureNotifiers()

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

	arena.ScoringPanelRegistry.initialize()

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
	arena.accessPoint2.SetSettings(settings.Ap2Address, settings.Ap2Username, settings.Ap2Password,
		settings.Ap2TeamChannel, 0, "", settings.NetworkSecurityEnabled)
	arena.networkSwitch = network.NewSwitch(settings.SwitchAddress, settings.SwitchPassword)
	arena.Plc.SetAddress(settings.PlcAddress)
	arena.TbaClient = partner.NewTbaClient(settings.TbaEventCode, settings.TbaSecretId, settings.TbaSecret)

	if arena.EventSettings.NetworkSecurityEnabled && arena.MatchState == PreMatch {
		if err = arena.accessPoint.ConfigureAdminWifi(); err != nil {
			log.Printf("Failed to configure admin WiFi: %s", err.Error())
		}
		if err = arena.accessPoint2.ConfigureAdminWifi(); err != nil {
			log.Printf("Failed to configure admin WiFi: %s", err.Error())
		}
	}

	game.MatchTiming.WarmupDurationSec = settings.WarmupDurationSec
	game.MatchTiming.AutoDurationSec = settings.AutoDurationSec
	game.MatchTiming.PauseDurationSec = settings.PauseDurationSec
	game.MatchTiming.TeleopDurationSec = settings.TeleopDurationSec
	game.MatchTiming.WarningRemainingDurationSec = settings.WarningRemainingDurationSec
	game.UpdateMatchSounds()
	arena.MatchTimingNotifier.Notify()

	game.QuintetThreshold = settings.QuintetThreshold
	game.CargoBonusRankingPointThresholdWithoutQuintet = settings.CargoBonusRankingPointThresholdWithoutQuintet
	game.CargoBonusRankingPointThresholdWithQuintet = settings.CargoBonusRankingPointThresholdWithQuintet
	game.HangarBonusRankingPointThreshold = settings.HangarBonusRankingPointThreshold

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

	arena.setupNetwork([6]*model.Team{arena.AllianceStations["R1"].Team, arena.AllianceStations["R2"].Team,
		arena.AllianceStations["R3"].Team, arena.AllianceStations["B1"].Team, arena.AllianceStations["B2"].Team,
		arena.AllianceStations["B3"].Team})

	// Reset the arena state and realtime scores.
	arena.soundsPlayed = make(map[*game.MatchSound]struct{})
	arena.RedRealtimeScore = NewRealtimeScore()
	arena.BlueRealtimeScore = NewRealtimeScore()
	arena.FieldVolunteers = false
	arena.FieldReset = false
	arena.ScoringPanelRegistry.resetScoreCommitted()
	arena.Plc.ResetMatch()

	// Notify any listeners about the new match.
	arena.MatchLoadNotifier.Notify()
	arena.RealtimeScoreNotifier.Notify()
	arena.AllianceStationDisplayMode = "match"
	arena.AllianceStationDisplayModeNotifier.Notify()

	return nil
}

// Sets a new test match containing no teams as the current match.
func (arena *Arena) LoadTestMatch() error {
	return arena.LoadMatch(&model.Match{Type: "test"})
}

// Loads the first unplayed match of the current match type.
func (arena *Arena) LoadNextMatch() error {
	nextMatch, err := arena.getNextMatch(false)
	if err != nil {
		return err
	}
	if nextMatch == nil {
		return arena.LoadTestMatch()
	}
	return arena.LoadMatch(nextMatch)
}

// Assigns the given team to the given station, also substituting it into the match record.
func (arena *Arena) SubstituteTeam(teamId int, station string) error {
	if !arena.CurrentMatch.ShouldAllowSubstitution() {
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
	arena.setupNetwork([6]*model.Team{arena.AllianceStations["R1"].Team, arena.AllianceStations["R2"].Team,
		arena.AllianceStations["R3"].Team, arena.AllianceStations["B1"].Team, arena.AllianceStations["B2"].Team,
		arena.AllianceStations["B3"].Team})
	arena.MatchLoadNotifier.Notify()

	if arena.CurrentMatch.Type != "test" {
		arena.Database.UpdateMatch(arena.CurrentMatch)
	}
	return nil
}

// Starts the match if all conditions are met.
func (arena *Arena) StartMatch() error {
	err := arena.checkCanStartMatch()
	if err == nil {
		// Save the match start time and game-specifc data to the database for posterity.
		arena.CurrentMatch.StartedAt = time.Now()
		if arena.CurrentMatch.Type != "test" {
			arena.Database.UpdateMatch(arena.CurrentMatch)
		}
		arena.updateCycleTime(arena.CurrentMatch.StartedAt)

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
				arena.Database.UpdateTeam(allianceStation.Team)
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

	if arena.MatchState != WarmupPeriod {
		arena.playSound("abort")
	}
	arena.MatchState = PostMatch
	arena.matchAborted = true
	arena.AudienceDisplayMode = "blank"
	arena.AudienceDisplayModeNotifier.Notify()
	arena.AllianceStationDisplayMode = "logo"
	arena.AllianceStationDisplayModeNotifier.Notify()
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

// Starts a timeout of the given duration.
func (arena *Arena) StartTimeout(durationSec int) error {
	if arena.MatchState != PreMatch {
		return fmt.Errorf("Cannot start timeout while there is a match still in progress or with results pending.")
	}

	game.MatchTiming.TimeoutDurationSec = durationSec
	arena.MatchTimingNotifier.Notify()
	arena.MatchState = TimeoutActive
	arena.MatchStartTime = time.Now()
	arena.LastMatchTimeSec = -1
	arena.AllianceStationDisplayMode = "timeout"
	arena.AllianceStationDisplayModeNotifier.Notify()

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
		arena.MatchStartTime = time.Now()
		arena.LastMatchTimeSec = -1
		auto = true
		arena.AudienceDisplayMode = "match"
		arena.AudienceDisplayModeNotifier.Notify()
		arena.AllianceStationDisplayMode = "match"
		arena.AllianceStationDisplayModeNotifier.Notify()
		if game.MatchTiming.WarmupDurationSec > 0 {
			arena.MatchState = WarmupPeriod
			enabled = false
			sendDsPacket = false
		} else {
			arena.MatchState = AutoPeriod
			enabled = true
			sendDsPacket = true
		}
		arena.Plc.ResetMatch()
		arena.Plc.SetHubMotors(true, rand.Intn(2) == 1)
	case WarmupPeriod:
		auto = true
		enabled = false
		if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec) {
			arena.MatchState = AutoPeriod
			auto = true
			enabled = true
			sendDsPacket = true
		}
	case AutoPeriod:
		auto = true
		enabled = true
		if matchTimeSec >= game.GetDurationToAutoEnd().Seconds() {
			auto = false
			sendDsPacket = true
			if game.MatchTiming.PauseDurationSec > 0 {
				arena.MatchState = PausePeriod
				enabled = false
			} else {
				arena.MatchState = TeleopPeriod
				enabled = true
			}
		}
	case PausePeriod:
		auto = false
		enabled = false
		if matchTimeSec >= game.GetDurationToTeleopStart().Seconds() {
			arena.MatchState = TeleopPeriod
			auto = false
			enabled = true
			sendDsPacket = true
		}
	case TeleopPeriod:
		auto = false
		enabled = true
		if matchTimeSec >= game.GetDurationToTeleopEnd().Seconds() {
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
			go func() {
				// Configure the network in advance for the next match after a delay.
				time.Sleep(time.Second * preLoadNextMatchDelaySec)
				arena.preLoadNextMatch()
			}()
		}
	case TimeoutActive:
		if matchTimeSec >= float64(game.MatchTiming.TimeoutDurationSec) {
			arena.MatchState = PostTimeout
			arena.playSound("end")
			go func() {
				// Leave the timer on the screen briefly at the end of the timeout period.
				time.Sleep(time.Second * matchEndScoreDwellSec)
				arena.AudienceDisplayMode = "blank"
				arena.AudienceDisplayModeNotifier.Notify()
				arena.AllianceStationDisplayMode = "logo"
				arena.AllianceStationDisplayModeNotifier.Notify()
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

	// Send a packet if at a period transition point or if it's been long enough since the last one.
	if sendDsPacket || time.Since(arena.lastDsPacketTime).Seconds()*1000 >= dsPacketPeriodMs {
		arena.sendDsPacket(auto, enabled)
		arena.ArenaStatusNotifier.Notify()
	}

	arena.handleSounds(matchTimeSec)

	// Handle field sensors/lights/actuators.
	arena.handlePlcInput()
	arena.handlePlcOutput()

	arena.LastMatchTimeSec = matchTimeSec
	arena.lastMatchState = arena.MatchState
}

// Loops indefinitely to track and update the arena components.
func (arena *Arena) Run() {
	// Start other loops in goroutines.
	go arena.listenForDriverStations()
	go arena.listenForDsUdpPackets()
	go arena.accessPoint.Run()
	go arena.accessPoint2.Run()
	go arena.Plc.Run()

	for {
		arena.Update()
		if time.Since(arena.lastPeriodicTaskTime).Seconds() >= periodicTaskPeriodSec {
			arena.lastPeriodicTaskTime = time.Now()
			go arena.runPeriodicTasks()
		}
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

// Returns the next match of the same type that is currently loaded, or nil if there are no more matches.
func (arena *Arena) getNextMatch(excludeCurrent bool) (*model.Match, error) {
	if arena.CurrentMatch.Type == "test" {
		return nil, nil
	}

	matches, err := arena.Database.GetMatchesByType(arena.CurrentMatch.Type)
	if err != nil {
		return nil, err
	}
	for _, match := range matches {
		if !match.IsComplete() && !(excludeCurrent && match.Id == arena.CurrentMatch.Id) {
			return &match, nil
		}
	}

	// There are no matches left of the same type.
	return nil, nil
}

// Configures the field network for the next match in advance of the current match being scored and committed.
func (arena *Arena) preLoadNextMatch() {
	if arena.MatchState != PostMatch {
		// The next match has already been loaded; no need to do anything.
		return
	}

	nextMatch, err := arena.getNextMatch(true)
	if err != nil {
		log.Printf("Failed to pre-load next match: %s", err.Error())
	}
	if nextMatch == nil {
		return
	}

	var teams [6]*model.Team
	for i, teamId := range []int{nextMatch.Red1, nextMatch.Red2, nextMatch.Red3, nextMatch.Blue1, nextMatch.Blue2,
		nextMatch.Blue3} {
		if teamId == 0 {
			continue
		}
		if teams[i], err = arena.Database.GetTeamById(teamId); err != nil {
			log.Printf("Failed to get model for Team %d while pre-loading next match: %s", teamId, err.Error())
		}
	}
	arena.setupNetwork(teams)
}

// Asynchronously reconfigures the networking hardware for the new set of teams.
func (arena *Arena) setupNetwork(teams [6]*model.Team) {
	if arena.EventSettings.NetworkSecurityEnabled {
		if arena.EventSettings.Ap2TeamChannel == 0 {
			// Only one AP is being used.
			if err := arena.accessPoint.ConfigureTeamWifi(teams); err != nil {
				log.Printf("Failed to configure team WiFi: %s", err.Error())
			}
		} else {
			// Two APs are being used. Configure the first for the red teams and the second for the blue teams.
			if err := arena.accessPoint.ConfigureTeamWifi([6]*model.Team{teams[0], teams[1], teams[2], nil, nil,
				nil}); err != nil {
				log.Printf("Failed to configure red alliance WiFi: %s", err.Error())
			}
			if err := arena.accessPoint2.ConfigureTeamWifi([6]*model.Team{nil, nil, nil, teams[3], teams[4],
				teams[5]}); err != nil {
				log.Printf("Failed to configure blue alliance WiFi: %s", err.Error())
			}
		}
		go func() {
			if err := arena.networkSwitch.ConfigureTeamEthernet(teams); err != nil {
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

	if arena.Plc.IsEnabled() {
		if !arena.Plc.IsHealthy {
			return fmt.Errorf("Cannot start match while PLC is not healthy.")
		}
		if arena.Plc.GetFieldEstop() {
			return fmt.Errorf("Cannot start match while field emergency stop is active.")
		}
		for name, status := range arena.Plc.GetArmorBlockStatuses() {
			if !status {
				return fmt.Errorf("Cannot start match while PLC ArmorBlock '%s' is not connected.", name)
			}
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
	redEthernets, blueEthernets := arena.Plc.GetEthernetConnected()
	arena.AllianceStations["R1"].Ethernet = redEthernets[0]
	arena.AllianceStations["R2"].Ethernet = redEthernets[1]
	arena.AllianceStations["R3"].Ethernet = redEthernets[2]
	arena.AllianceStations["B1"].Ethernet = blueEthernets[0]
	arena.AllianceStations["B2"].Ethernet = blueEthernets[1]
	arena.AllianceStations["B3"].Ethernet = blueEthernets[2]

	if arena.MatchState == PreMatch || arena.MatchState == PostMatch || arena.MatchState == TimeoutActive ||
		arena.MatchState == PostTimeout {
		// Don't do anything if we're outside the match, otherwise we may overwrite manual edits.
		return
	}

	redScore := &arena.RedRealtimeScore.CurrentScore
	oldRedScore := *redScore
	blueScore := &arena.BlueRealtimeScore.CurrentScore
	oldBlueScore := *blueScore
	matchStartTime := arena.MatchStartTime
	currentTime := time.Now()

	if arena.Plc.IsEnabled() {
		// Handle hub.
		redLowerHubCounts, redUpperHubCounts, blueLowerHubCounts, blueUpperHubCounts := arena.Plc.GetHubCounts()
		redHub := &arena.RedRealtimeScore.hub
		redHub.UpdateState(redLowerHubCounts, redUpperHubCounts, matchStartTime, currentTime)
		redScore.AutoCargoLower = redHub.AutoCargoLower
		redScore.AutoCargoUpper = redHub.AutoCargoUpper
		redScore.TeleopCargoLower = redHub.TeleopCargoLower
		redScore.TeleopCargoUpper = redHub.TeleopCargoUpper
		blueHub := &arena.RedRealtimeScore.hub
		blueHub.UpdateState(blueLowerHubCounts, blueUpperHubCounts, matchStartTime, currentTime)
		blueScore.AutoCargoLower = blueHub.AutoCargoLower
		blueScore.AutoCargoUpper = blueHub.AutoCargoUpper
		blueScore.TeleopCargoLower = blueHub.TeleopCargoLower
		blueScore.TeleopCargoUpper = blueHub.TeleopCargoUpper
	}

	if !oldRedScore.Equals(redScore) || !oldBlueScore.Equals(blueScore) {
		arena.RealtimeScoreNotifier.Notify()
	}
}

// Updates the PLC's coils based on its inputs and the current scoring state.
func (arena *Arena) handlePlcOutput() {
	switch arena.MatchState {
	case PreMatch:
		if arena.lastMatchState != PreMatch {
			arena.Plc.SetFieldResetLight(true)
		}
		fallthrough
	case TimeoutActive:
		fallthrough
	case PostTimeout:
		// Set the stack light state -- solid alliance color(s) if robots are not connected, solid orange if scores are
		// not input, or blinking green if ready.
		redAllianceReady := arena.checkAllianceStationsReady("R1", "R2", "R3") == nil
		blueAllianceReady := arena.checkAllianceStationsReady("B1", "B2", "B3") == nil
		greenStackLight := redAllianceReady && blueAllianceReady && arena.Plc.GetCycleState(2, 0, 2)
		arena.Plc.SetStackLights(!redAllianceReady, !blueAllianceReady, false, greenStackLight)
		arena.Plc.SetStackBuzzer(redAllianceReady && blueAllianceReady)

		// Turn off lights if all teams become ready.
		if redAllianceReady && blueAllianceReady {
			arena.Plc.SetFieldResetLight(false)
		}
	case PostMatch:
		if arena.FieldReset {
			arena.Plc.SetFieldResetLight(true)
		}
		scoreReady := arena.RedRealtimeScore.FoulsCommitted && arena.BlueRealtimeScore.FoulsCommitted &&
			arena.alliancePostMatchScoreReady("red") && arena.alliancePostMatchScoreReady("blue")
		arena.Plc.SetStackLights(false, false, !scoreReady, false)

		if arena.lastMatchState != PostMatch {
			go func() {
				time.Sleep(time.Second * game.HubTeleopGracePeriodSec)
				arena.Plc.SetHubMotors(false, false)
			}()
		}
	case AutoPeriod:
		fallthrough
	case PausePeriod:
		fallthrough
	case TeleopPeriod:
		arena.Plc.SetStackLights(false, false, false, true)
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

func (arena *Arena) handleSounds(matchTimeSec float64) {
	if arena.MatchState == PreMatch || arena.MatchState == TimeoutActive || arena.MatchState == PostTimeout {
		// Only apply this logic during a match.
		return
	}

	for _, sound := range game.MatchSounds {
		if sound.MatchTimeSec < 0 {
			// Skip sounds with negative timestamps; they are meant to only be triggered explicitly.
			continue
		}
		if _, ok := arena.soundsPlayed[sound]; !ok {
			if matchTimeSec > sound.MatchTimeSec {
				arena.playSound(sound.Name)
				arena.soundsPlayed[sound] = struct{}{}
			}
		}
	}
}

func (arena *Arena) playSound(name string) {
	if !arena.MuteMatchSounds {
		arena.PlaySoundNotifier.NotifyWithMessage(name)
	}
}

func (arena *Arena) alliancePostMatchScoreReady(alliance string) bool {
	numPanels := arena.ScoringPanelRegistry.GetNumPanels(alliance)
	return numPanels > 0 && arena.ScoringPanelRegistry.GetNumScoreCommitted(alliance) >= numPanels
}

// Performs any actions that need to run at the interval specified by periodicTaskPeriodSec.
func (arena *Arena) runPeriodicTasks() {
	arena.updateEarlyLateMessage()
	arena.purgeDisconnectedDisplays()
}
