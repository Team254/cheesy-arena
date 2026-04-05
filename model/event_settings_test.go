// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventSettingsReadWrite(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	eventSettings, err := db.GetEventSettings()
	assert.Nil(t, err)
	assert.Equal(
		t,
		EventSettings{
			Id:                          1,
			Name:                        "Untitled Event",
			PlayoffType:                 DoubleEliminationPlayoff,
			NumPlayoffAlliances:         8,
			SelectionRound2Order:        "L",
			SelectionRound3Order:        "",
			SelectionShowUnpickedTeams:  true,
			TbaDownloadEnabled:          true,
			ApChannel:                   36,
			SCCUpCommands:               "configure terminal\ninterface range gigabitEthernet 1/2-4\nno shutdown\nexit\nexit\nexit",
			SCCDownCommands:             "configure terminal\ninterface range gigabitEthernet 1/2-4\nshutdown\nexit\nexit\nexit",
			AutoDurationSec:             20,
			PauseDurationSec:            3,
			TransitionShiftDurationSec:  10,
			ShiftDurationSec:            25,
			EndgameDurationSec:          30,
			AutoBonusCoralThreshold:     1,
			CoralBonusPerLevelThreshold: 7,
			CoralBonusCoopEnabled:       true,
			BargeBonusPointThreshold:    16,
			IncludeAlgaeInBargeBonus:    false,
			CompanionAddress:            "",
			CompanionPort:               0,
		},
		*eventSettings,
	)
	assert.Equal(t, 140, game.GetTeleopDurationSec())

	eventSettings.Name = "Chezy Champs"
	eventSettings.NumPlayoffAlliances = 6
	eventSettings.SelectionRound2Order = "F"
	eventSettings.SelectionRound3Order = "L"
	err = db.UpdateEventSettings(eventSettings)
	assert.Nil(t, err)
	eventSettings2, err := db.GetEventSettings()
	assert.Nil(t, err)
	assert.Equal(t, eventSettings, eventSettings2)
}
