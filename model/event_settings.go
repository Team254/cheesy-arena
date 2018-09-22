// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore read/write methods for event-level configuration.

package model

type EventSettings struct {
	Id                      int
	Name                    string
	NumElimAlliances        int
	SelectionRound2Order    string
	SelectionRound3Order    string
	TBADownloadEnabled      bool
	TbaPublishingEnabled    bool
	TbaEventCode            string
	TbaSecretId             string
	TbaSecret               string
	NetworkSecurityEnabled  bool
	ApAddress               string
	ApUsername              string
	ApPassword              string
	ApTeamChannel           int
	ApAdminChannel          int
	ApAdminWpaKey           string
	SwitchAddress           string
	SwitchPassword          string
	PlcAddress              string
	AdminPassword           string
	ReaderPassword          string
	StemTvPublishingEnabled bool
	StemTvEventCode         string
	ScaleLedAddress         string
	RedSwitchLedAddress     string
	BlueSwitchLedAddress    string
	RedVaultLedAddress      string
	BlueVaultLedAddress     string
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
