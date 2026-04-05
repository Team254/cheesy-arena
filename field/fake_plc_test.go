// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains a fake implementation of the PLC interface for testing.

package field

import (
	"github.com/Team254/cheesy-arena/websocket"
)

type FakePlc struct {
	isEnabled             bool
	fieldEStop            bool
	redEStops             [3]bool
	blueEStops            [3]bool
	redAStops             [3]bool
	blueAStops            [3]bool
	redEthernetConnected  [3]bool
	blueEthernetConnected [3]bool
	stackLights           [4]bool
	stackLightBuzzer      bool
	fieldResetLight       bool
	awardsModeLight       bool
	cycleState            bool
	redHubCount           int
	blueHubCount          int
	redHubMotor           bool
	blueHubMotor          bool
	redHubLight           bool
	blueHubLight          bool
}

func (plc *FakePlc) SetAddress(address string) {
}

func (plc *FakePlc) IsEnabled() bool {
	return plc.isEnabled
}

func (plc *FakePlc) IsHealthy() bool {
	return true
}

func (plc *FakePlc) IoChangeNotifier() *websocket.Notifier {
	return nil
}

func (plc *FakePlc) Run() {
}

func (plc *FakePlc) GetArmorBlockStatuses() map[string]bool {
	return map[string]bool{}
}

func (plc *FakePlc) GetFieldEStop() bool {
	return plc.fieldEStop
}

func (plc *FakePlc) GetTeamEStops() ([3]bool, [3]bool) {
	return plc.redEStops, plc.blueEStops
}

func (plc *FakePlc) GetTeamAStops() ([3]bool, [3]bool) {
	return plc.redAStops, plc.blueAStops
}

func (plc *FakePlc) GetEthernetConnected() ([3]bool, [3]bool) {
	return plc.redEthernetConnected, plc.blueEthernetConnected
}

func (plc *FakePlc) ResetMatch() {
}

func (plc *FakePlc) SetStackLights(red, blue, orange, green bool) {
	plc.stackLights[0] = red
	plc.stackLights[1] = blue
	plc.stackLights[2] = orange
	plc.stackLights[3] = green
}

func (plc *FakePlc) SetStackBuzzer(state bool) {
	plc.stackLightBuzzer = state
}

func (plc *FakePlc) SetFieldResetLight(state bool) {
	plc.fieldResetLight = state
}

func (plc *FakePlc) SetAwardsModeLight(state bool) {
	plc.awardsModeLight = state
}

func (plc *FakePlc) GetHubCounts() (int, int) {
	return plc.redHubCount, plc.blueHubCount
}

func (plc *FakePlc) SetHubMotors(red, blue bool) {
	plc.redHubMotor = red
	plc.blueHubMotor = blue
}

func (plc *FakePlc) SetHubLights(red, blue bool) {
	plc.redHubLight = red
	plc.blueHubLight = blue
}

func (plc *FakePlc) GetCycleState(max, index, duration int) bool {
	return plc.cycleState
}

func (plc *FakePlc) GetInputNames() []string {
	return []string{}
}

func (plc *FakePlc) GetRegisterNames() []string {
	return []string{}
}

func (plc *FakePlc) GetCoilNames() []string {
	return []string{}
}

func (plc *FakePlc) SetCoilOverride(index int, state bool) {
	// Not needed for testing, just to satisfy the interface.
}

func (plc *FakePlc) ClearCoilOverride(index int) {
	// Not needed for testing, just to satisfy the interface.
}
