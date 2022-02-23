// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore read/write methods for event-level configuration.

package model

import "github.com/Team254/cheesy-arena/game"

type EventSettings struct {
	Id                                            int `db:"id"`
	Name                                          string
	NumElimAlliances                              int
	SelectionRound2Order                          string
	SelectionRound3Order                          string
	TBADownloadEnabled                            bool
	TbaPublishingEnabled                          bool
	TbaEventCode                                  string
	TbaSecretId                                   string
	TbaSecret                                     string
	NetworkSecurityEnabled                        bool
	ApAddress                                     string
	ApUsername                                    string
	ApPassword                                    string
	ApTeamChannel                                 int
	ApAdminChannel                                int
	ApAdminWpaKey                                 string
	Ap2Address                                    string
	Ap2Username                                   string
	Ap2Password                                   string
	Ap2TeamChannel                                int
	SwitchAddress                                 string
	SwitchPassword                                string
	PlcAddress                                    string
	AdminPassword                                 string
	WarmupDurationSec                             int
	AutoDurationSec                               int
	PauseDurationSec                              int
	TeleopDurationSec                             int
	WarningRemainingDurationSec                   int
	QuintetThreshold                              int
	CargoBonusRankingPointThresholdWithoutQuintet int
	CargoBonusRankingPointThresholdWithQuintet    int
	HangarBonusRankingPointThreshold              int
}

func (database *Database) GetEventSettings() (*EventSettings, error) {
	var allEventSettings []EventSettings
	if err := database.eventSettingsTable.getAll(&allEventSettings); err != nil {
		return nil, err
	}
	if len(allEventSettings) == 1 {
		return &allEventSettings[0], nil
	}

	// Database record doesn't exist yet; create it now.
	eventSettings := EventSettings{
		Name:                        "Untitled Event",
		NumElimAlliances:            8,
		SelectionRound2Order:        "L",
		SelectionRound3Order:        "",
		TBADownloadEnabled:          true,
		ApTeamChannel:               157,
		ApAdminChannel:              0,
		ApAdminWpaKey:               "1234Five",
		Ap2TeamChannel:              0,
		WarmupDurationSec:           game.MatchTiming.WarmupDurationSec,
		AutoDurationSec:             game.MatchTiming.AutoDurationSec,
		PauseDurationSec:            game.MatchTiming.PauseDurationSec,
		TeleopDurationSec:           game.MatchTiming.TeleopDurationSec,
		WarningRemainingDurationSec: game.MatchTiming.WarningRemainingDurationSec,
		QuintetThreshold:            game.QuintetThreshold,
		CargoBonusRankingPointThresholdWithoutQuintet: game.CargoBonusRankingPointThresholdWithoutQuintet,
		CargoBonusRankingPointThresholdWithQuintet:    game.CargoBonusRankingPointThresholdWithQuintet,
		HangarBonusRankingPointThreshold:              game.HangarBonusRankingPointThreshold,
	}

	if err := database.eventSettingsTable.create(&eventSettings); err != nil {
		return nil, err
	}
	return &eventSettings, nil
}

func (database *Database) UpdateEventSettings(eventSettings *EventSettings) error {
	return database.eventSettingsTable.update(eventSettings)
}
