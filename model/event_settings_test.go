// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
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
			WarmupDurationSec:           0,
			AutoDurationSec:             15,
			PauseDurationSec:            3,
			TeleopDurationSec:           135,
			WarningRemainingDurationSec: 20,
			AutoBonusCoralThreshold:     1,
			CoralBonusPerLevelThreshold: 7,
			CoralBonusCoopEnabled:       true,
			BargeBonusPointThreshold:    16,
			IncludeAlgaeInBargeBonus:    false,
			CompanionAddress:            "",
			CompanionPort:               51234,
		},
		*eventSettings,
	)

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
