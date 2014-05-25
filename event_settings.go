// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore read/write methods for event-level configuration.

package main

type EventSettings struct {
	Id   int
	Name string
	Code string
}

const eventSettingsId = 0

func (database *Database) GetEventSettings() (*EventSettings, error) {
	eventSettings := new(EventSettings)
	err := database.eventSettingsMap.Get(eventSettings, eventSettingsId)
	if err != nil {
		// Database record doesn't exist yet; create it now.
		err = database.eventSettingsMap.Insert(eventSettings)
		if err != nil {
			return nil, err
		}
		return new(EventSettings), nil
	}
	return eventSettings, nil
}

func (database *Database) SaveEventSettings(eventSettings *EventSettings) error {
	eventSettings.Id = eventSettingsId
	_, err := database.eventSettingsMap.Update(eventSettings)
	return err
}
