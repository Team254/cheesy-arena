// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore read/write methods for event-level configuration.

package model

import "github.com/Team254/cheesy-arena/game"

type EventSettings struct {
	Id                          int
	Name                        string
	NumElimAlliances            int
	SelectionRound2Order        string
	SelectionRound3Order        string
	TBADownloadEnabled          bool
	TbaPublishingEnabled        bool
	TbaEventCode                string
	TbaSecretId                 string
	TbaSecret                   string
	NetworkSecurityEnabled      bool
	ApAddress                   string
	ApUsername                  string
	ApPassword                  string
	ApTeamChannel               int
	ApAdminChannel              int
	ApAdminWpaKey               string
	Ap2Address                  string
	Ap2Username                 string
	Ap2Password                 string
	Ap2TeamChannel              int
	SwitchAddress               string
	SwitchPassword              string
	PlcAddress                  string
	AdminPassword               string
	WarmupDurationSec           int
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	Stage1Capacity              int
	Stage2Capacity              int
	Stage3Capacity              int
}

const eventSettingsId = 0

func (database *Database) GetEventSettings() (*EventSettings, error) {
	eventSettings := new(EventSettings)
	err := database.eventSettingsMap.Get(eventSettings, eventSettingsId)
	if err != nil {
		// Database record doesn't exist yet; create it now.
		eventSettings.Name = "Untitled Event"
		eventSettings.NumElimAlliances = 8
		eventSettings.SelectionRound2Order = "L"
		eventSettings.SelectionRound3Order = ""
		eventSettings.TBADownloadEnabled = true
		eventSettings.ApTeamChannel = 157
		eventSettings.ApAdminChannel = 0
		eventSettings.ApAdminWpaKey = "1234Five"
		eventSettings.Ap2TeamChannel = 0
		eventSettings.WarmupDurationSec = game.MatchTiming.WarmupDurationSec
		eventSettings.AutoDurationSec = game.MatchTiming.AutoDurationSec
		eventSettings.PauseDurationSec = game.MatchTiming.PauseDurationSec
		eventSettings.TeleopDurationSec = game.MatchTiming.TeleopDurationSec
		eventSettings.WarningRemainingDurationSec = game.MatchTiming.WarningRemainingDurationSec
		eventSettings.Stage1Capacity = game.StageCapacities[game.Stage1]
		eventSettings.Stage2Capacity = game.StageCapacities[game.Stage2]
		eventSettings.Stage3Capacity = game.StageCapacities[game.Stage3]

		err = database.eventSettingsMap.Insert(eventSettings)
		if err != nil {
			return nil, err
		}
	}
	return eventSettings, nil
}

func (database *Database) SaveEventSettings(eventSettings *EventSettings) error {
	eventSettings.Id = eventSettingsId
	_, err := database.eventSettingsMap.Update(eventSettings)
	return err
}
