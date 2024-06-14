// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore read/write methods for event-level configuration.

package model

import "github.com/Team254/cheesy-arena/game"

type PlayoffType int

const (
	DoubleEliminationPlayoff PlayoffType = iota
	SingleEliminationPlayoff
)

type EventSettings struct {
	Id                              int `db:"id"`
	Name                            string
	PlayoffType                     PlayoffType
	NumPlayoffAlliances             int
	SelectionRound2Order            string
	SelectionRound3Order            string
	SelectionShowUnpickedTeams      bool
	TbaDownloadEnabled              bool
	TbaPublishingEnabled            bool
	TbaEventCode                    string
	TbaSecretId                     string
	TbaSecret                       string
	NexusEnabled                    bool
	NetworkSecurityEnabled          bool
	ApAddress                       string
	ApPassword                      string
	ApChannel                       int
	SwitchAddress                   string
	SwitchPassword                  string
	PlcAddress                      string
	AdminPassword                   string
	TeamSignRed1Address             string
	TeamSignRed2Address             string
	TeamSignRed3Address             string
	TeamSignRedTimerAddress         string
	TeamSignBlue1Address            string
	TeamSignBlue2Address            string
	TeamSignBlue3Address            string
	TeamSignBlueTimerAddress        string
	WarmupDurationSec               int
	AutoDurationSec                 int
	PauseDurationSec                int
	TeleopDurationSec               int
	WarningRemainingDurationSec     int
	MelodyBonusThresholdWithoutCoop int
	MelodyBonusThresholdWithCoop    int
	AmplificationNoteLimit          int
	AmplificationDurationSec        int
}

func (database *Database) GetEventSettings() (*EventSettings, error) {
	allEventSettings, err := database.eventSettingsTable.getAll()
	if err != nil {
		return nil, err
	}
	if len(allEventSettings) == 1 {
		return &allEventSettings[0], nil
	}

	// Database record doesn't exist yet; create it now.
	eventSettings := EventSettings{
		Name:                            "Untitled Event",
		PlayoffType:                     DoubleEliminationPlayoff,
		NumPlayoffAlliances:             8,
		SelectionRound2Order:            "L",
		SelectionRound3Order:            "",
		SelectionShowUnpickedTeams:      false,
		TbaDownloadEnabled:              true,
		ApChannel:                       36,
		WarmupDurationSec:               game.MatchTiming.WarmupDurationSec,
		AutoDurationSec:                 game.MatchTiming.AutoDurationSec,
		PauseDurationSec:                game.MatchTiming.PauseDurationSec,
		TeleopDurationSec:               game.MatchTiming.TeleopDurationSec,
		WarningRemainingDurationSec:     game.MatchTiming.WarningRemainingDurationSec,
		MelodyBonusThresholdWithoutCoop: game.MelodyBonusThresholdWithoutCoop,
		MelodyBonusThresholdWithCoop:    game.MelodyBonusThresholdWithCoop,
		AmplificationNoteLimit:          game.AmplificationNoteLimit,
		AmplificationDurationSec:        game.AmplificationDurationSec,
	}

	if err := database.eventSettingsTable.create(&eventSettings); err != nil {
		return nil, err
	}
	return &eventSettings, nil
}

func (database *Database) UpdateEventSettings(eventSettings *EventSettings) error {
	return database.eventSettingsTable.update(eventSettings)
}
