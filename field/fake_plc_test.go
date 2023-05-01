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
	fieldEstop            bool
	redEstops             [3]bool
	blueEstops            [3]bool
	redEthernetConnected  [3]bool
	blueEthernetConnected [3]bool
	stackLights           [4]bool
	stackLightBuzzer      bool
	fieldResetLight       bool
	cycleState            bool
	chargeStationsLevel   [2]bool
	chargeStationLights   [2]bool
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

func (plc *FakePlc) GetFieldEstop() bool {
	return plc.fieldEstop
}

func (plc *FakePlc) GetTeamEstops() ([3]bool, [3]bool) {
	return plc.redEstops, plc.blueEstops
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

func (plc *FakePlc) GetChargeStationsLevel() (bool, bool) {
	return plc.chargeStationsLevel[0], plc.chargeStationsLevel[1]
}

func (plc *FakePlc) SetChargeStationLights(redState, blueState bool) {
	plc.chargeStationLights[0] = redState
	plc.chargeStationLights[1] = blueState
}
