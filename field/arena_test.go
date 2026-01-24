// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Tests for the 2026 REBUILT game logic in Arena.

package field

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/stretchr/testify/assert"
)

// Minimal Mock PLC implementation satisfying plc.Plc interface
type MockPlc struct {
	RedFuelVal  int
	BlueFuelVal int
	HubRed      bool
	HubBlue     bool
}

func (m *MockPlc) SetAddress(string)                        {}
func (m *MockPlc) IsEnabled() bool                          { return true }
func (m *MockPlc) IsHealthy() bool                          { return true }
func (m *MockPlc) IoChangeNotifier() *websocket.Notifier    { return nil }
func (m *MockPlc) Run()                                     {}
func (m *MockPlc) GetArmorBlockStatuses() map[string]bool   { return nil }
func (m *MockPlc) GetFieldEStop() bool                      { return false }
func (m *MockPlc) GetTeamEStops() ([3]bool, [3]bool)        { return [3]bool{}, [3]bool{} }
func (m *MockPlc) GetTeamAStops() ([3]bool, [3]bool)        { return [3]bool{}, [3]bool{} }
func (m *MockPlc) GetEthernetConnected() ([3]bool, [3]bool) { return [3]bool{}, [3]bool{} }
func (m *MockPlc) ResetMatch()                              {}
func (m *MockPlc) SetStackLights(bool, bool, bool, bool)    {}
func (m *MockPlc) SetStackBuzzer(bool)                      {}
func (m *MockPlc) SetFieldResetLight(bool)                  {}
func (m *MockPlc) GetCycleState(int, int, int) bool         { return false }
func (m *MockPlc) GetInputNames() []string                  { return nil }
func (m *MockPlc) GetRegisterNames() []string               { return nil }
func (m *MockPlc) GetCoilNames() []string                   { return nil }
func (m *MockPlc) GetFuelCounts() (int, int)                { return m.RedFuelVal, m.BlueFuelVal }
func (m *MockPlc) SetHubLights(red bool, blue bool)         { m.HubRed = red; m.HubBlue = blue }

// Helper to create a basic arena for testing
func createTestArena() *Arena {
	arena := &Arena{
		RedRealtimeScore:  NewRealtimeScore(),
		BlueRealtimeScore: NewRealtimeScore(),
		CurrentMatch:      &model.Match{},
		Plc:               &MockPlc{},
	}
	// Default timing settings
	game.MatchTiming.WarmupDurationSec = 0
	game.MatchTiming.AutoDurationSec = 15
	game.MatchTiming.PauseDurationSec = 3
	game.MatchTiming.TeleopDurationSec = 140

	// Initialize alliance stations to avoid panic in sendDsPacket
	arena.AllianceStations = make(map[string]*AllianceStation)
	for _, s := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		arena.AllianceStations[s] = &AllianceStation{}
	}
	return arena
}

func TestRedWonAutoFuel(t *testing.T) {
	arena := createTestArena()

	// Case 1: Red Wins (Fuel: 10 vs 5)
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 10
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 5
	assert.True(t, arena.redWonAutoFuel(), "Red should win when having more fuel")

	// Case 2: Blue Wins (Fuel: 5 vs 10)
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 5
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 10
	assert.False(t, arena.redWonAutoFuel(), "Blue should win when having more fuel")

	// Case 3: Tie - Red Random Win
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 10
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 10
	arena.autoTieBreakerRedWin = true
	assert.True(t, arena.redWonAutoFuel(), "Red should win tie if autoTieBreakerRedWin is true")

	// Case 4: Tie - Blue Random Win
	arena.autoTieBreakerRedWin = false
	assert.False(t, arena.redWonAutoFuel(), "Blue should win tie if autoTieBreakerRedWin is false")
}

func TestUpdateGameSpecificMessage(t *testing.T) {
	arena := createTestArena()

	// If Red Wins Auto -> Blue Advantage -> Message "B"
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 20
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 10
	arena.updateGameSpecificMessage()
	assert.Equal(t, "B", arena.RedRealtimeScore.GameSpecificMessage)
	assert.Equal(t, "B", arena.BlueRealtimeScore.GameSpecificMessage)

	// If Blue Wins Auto -> Red Advantage -> Message "R"
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 10
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 20
	arena.updateGameSpecificMessage()
	assert.Equal(t, "R", arena.RedRealtimeScore.GameSpecificMessage)
	assert.Equal(t, "R", arena.BlueRealtimeScore.GameSpecificMessage)
}

func TestUpdateHubStatus_RedWinsAuto(t *testing.T) {
	// Scenario: Red Wins Auto (Message "B")
	// Shift 1: Blue Active, Red Inactive
	arena := createTestArena()
	arena.MatchState = TeleopPeriod
	arena.RedRealtimeScore.CurrentScore.AutoFuelCount = 20
	arena.BlueRealtimeScore.CurrentScore.AutoFuelCount = 10

	// Simulate Match Time (Teleop starts at 18s)
	// Auto(15) + Pause(3) = 18s

	// 1. Transition (140s -> 130s left) => Both Active
	// MatchTime: 18 + 5 = 23s (135s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 23)
	arena.updateHubStatus()
	assert.True(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Transition: Red should be Active")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Transition: Blue should be Active")

	// 2. Shift 1 (130s -> 105s left) => Blue Active, Red Inactive
	// MatchTime: 18 + 20 = 38s (120s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 38)
	arena.updateHubStatus()
	assert.False(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Shift 1 (Red Won): Red Inactive")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Shift 1 (Red Won): Blue Active")

	// 3. Shift 2 (105s -> 80s left) => Red Active, Blue Inactive
	// MatchTime: 18 + 50 = 68s (90s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 68)
	arena.updateHubStatus()
	assert.True(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Shift 2 (Red Won): Red Active")
	assert.False(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Shift 2 (Red Won): Blue Inactive")

	// 4. Shift 3 (80s -> 55s left) => Blue Active, Red Inactive
	// MatchTime: 18 + 70 = 88s (70s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 88)
	arena.updateHubStatus()
	assert.False(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Shift 3 (Red Won): Red Inactive")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Shift 3 (Red Won): Blue Active")

	// 5. Shift 4 (55s -> 30s left) => Red Active, Blue Inactive
	// MatchTime: 18 + 100 = 118s (40s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 118)
	arena.updateHubStatus()
	assert.True(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Shift 4 (Red Won): Red Active")
	assert.False(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Shift 4 (Red Won): Blue Inactive")

	// 6. Endgame (30s -> 0s left) => Both Active
	// MatchTime: 18 + 130 = 148s (10s left)
	arena.MatchStartTime = time.Now().Add(-time.Second * 148)
	arena.updateHubStatus()
	assert.True(t, arena.RedRealtimeScore.CurrentScore.HubActive, "Endgame: Red should be Active")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.HubActive, "Endgame: Blue should be Active")
}

func TestPlcFuelScoring(t *testing.T) {
	arena := createTestArena()
	mockPlc := arena.Plc.(*MockPlc)
	arena.MatchState = TeleopPeriod
	arena.RedRealtimeScore.CurrentScore.HubActive = true // Allow scoring

	// 1. Initial State
	arena.handlePlcInputOutput()
	assert.Equal(t, 0, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)

	// 2. Increase Fuel by 5
	mockPlc.RedFuelVal = 5
	arena.handlePlcInputOutput()
	assert.Equal(t, 5, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)
	assert.Equal(t, 5, arena.lastRedPlcFuel)

	// 3. Increase Fuel by 2 (Total 7)
	mockPlc.RedFuelVal = 7
	arena.handlePlcInputOutput()
	assert.Equal(t, 7, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)

	// 4. PLC Reset (Value drops to 0) -> Score should NOT change/decrease
	mockPlc.RedFuelVal = 0
	arena.handlePlcInputOutput()
	assert.Equal(t, 7, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)
	assert.Equal(t, 0, arena.lastRedPlcFuel) // Last value tracks reset

	// 5. Increase from 0 to 3 -> Score should increase by 3 (Total 10)
	mockPlc.RedFuelVal = 3
	arena.handlePlcInputOutput()
	assert.Equal(t, 10, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)
}

func TestPlcFuelScoring_InactiveHub(t *testing.T) {
	arena := createTestArena()
	mockPlc := arena.Plc.(*MockPlc)
	arena.MatchState = TeleopPeriod

	// Set Hub to Inactive
	arena.RedRealtimeScore.CurrentScore.HubActive = false

	// Try to score 5 fuel
	mockPlc.RedFuelVal = 5
	arena.handlePlcInputOutput()

	// Should NOT add to score
	assert.Equal(t, 0, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)

	// BUT should update last value (so these 5 aren't counted later when active)
	assert.Equal(t, 5, arena.lastRedPlcFuel)

	// Set Hub to Active
	arena.RedRealtimeScore.CurrentScore.HubActive = true

	// Score 2 more (Total 7)
	mockPlc.RedFuelVal = 7
	arena.handlePlcInputOutput()

	// Should only add the new 2
	assert.Equal(t, 2, arena.RedRealtimeScore.CurrentScore.TeleopFuelCount)
}
