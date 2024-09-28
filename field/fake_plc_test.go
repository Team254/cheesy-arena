// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains a fake implementation of the PLC interface for testing.

package field

import (
	"github.com/Team254/cheesy-arena/websocket"
)

type FakePlc struct {
	isEnabled                bool
	fieldEStop               bool
	redEStops                [3]bool
	blueEStops               [3]bool
	redAStops                [3]bool
	blueAStops               [3]bool
	redEthernetConnected     [3]bool
	blueEthernetConnected    [3]bool
	stackLights              [4]bool
	stackLightBuzzer         bool
	fieldResetLight          bool
	cycleState               bool
	redAmpButtons            [2]bool
	blueAmpButtons           [2]bool
	redNoteCounts            [2]int
	blueNoteCounts           [2]int
	speakerMotors            bool
	redSpeakerLight          bool
	blueSpeakerLight         bool
	redSubwooferCountdown    bool
	blueSubwooferCountdown   bool
	redAmpLights             [3]bool
	blueAmpLights            [3]bool
	postMatchSubwooferLights bool
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

func (plc *FakePlc) GetAmpButtons() (bool, bool, bool, bool) {
	return plc.redAmpButtons[0], plc.redAmpButtons[1], plc.blueAmpButtons[0], plc.blueAmpButtons[1]
}

func (plc *FakePlc) GetAmpSpeakerNoteCounts() (int, int, int, int) {
	return plc.redNoteCounts[0], plc.redNoteCounts[1], plc.blueNoteCounts[0], plc.blueNoteCounts[1]
}

func (plc *FakePlc) SetSpeakerMotors(state bool) {
	plc.speakerMotors = state
}

func (plc *FakePlc) SetSpeakerLights(redState, blueState bool) {
	plc.redSpeakerLight = redState
	plc.blueSpeakerLight = blueState
}

func (plc *FakePlc) SetSubwooferCountdown(redState, blueState bool) {
	plc.redSubwooferCountdown = redState
	plc.blueSubwooferCountdown = blueState
}

func (plc *FakePlc) SetAmpLights(redLow, redHigh, redCoop, blueLow, blueHigh, blueCoop bool) {
	plc.redAmpLights[0] = redLow
	plc.redAmpLights[1] = redHigh
	plc.redAmpLights[2] = redCoop
	plc.blueAmpLights[0] = blueLow
	plc.blueAmpLights[1] = blueHigh
	plc.blueAmpLights[2] = blueCoop
}

func (plc *FakePlc) SetPostMatchSubwooferLights(state bool) {
	plc.postMatchSubwooferLights = state
}
