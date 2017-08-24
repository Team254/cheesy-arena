// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for an alliance-team association.

package model

type AllianceTeam struct {
	Id           int
	AllianceId   int
	PickPosition int
	TeamId       int
}

func (database *Database) CreateAllianceTeam(allianceTeam *AllianceTeam) error {
	return database.allianceTeamMap.Insert(allianceTeam)
}

func (database *Database) GetTeamsByAlliance(allianceId int) ([]AllianceTeam, error) {
	var allianceTeams []AllianceTeam
	err := database.allianceTeamMap.Select(&allianceTeams,
		"SELECT * FROM alliance_teams WHERE allianceid = ? ORDER BY pickposition", allianceId)
	return allianceTeams, err
}

func (database *Database) SaveAllianceTeam(allianceTeam *AllianceTeam) error {
	_, err := database.allianceTeamMap.Update(allianceTeam)
	return err
}

func (database *Database) DeleteAllianceTeam(allianceTeam *AllianceTeam) error {
	_, err := database.allianceTeamMap.Delete(allianceTeam)
	return err
}

func (database *Database) TruncateAllianceTeams() error {
	return database.allianceTeamMap.TruncateTables()
}

func (database *Database) GetAllAlliances() ([][]AllianceTeam, error) {
	alliances := make([][]AllianceTeam, 0)
	var allianceTeams []AllianceTeam
	err := database.allianceTeamMap.Select(&allianceTeams,
		"SELECT * FROM alliance_teams ORDER BY allianceid, pickposition")
	if err == nil {
		// Format the sorted list of teams into a two-dimensional slice.
		currentAllianceId := -1
		for _, allianceTeam := range allianceTeams {
			if allianceTeam.AllianceId != currentAllianceId {
				currentAllianceId = allianceTeam.AllianceId
				alliances = append(alliances, make([]AllianceTeam, 0))
			}
			alliances[len(alliances)-1] = append(alliances[len(alliances)-1], allianceTeam)
		}
	}
	return alliances, err
}
