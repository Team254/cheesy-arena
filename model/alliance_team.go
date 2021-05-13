// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for an alliance-team association.

package model

import "sort"

type AllianceTeam struct {
	Id           int `db:"id"`
	AllianceId   int
	PickPosition int
	TeamId       int
}

func (database *Database) CreateAllianceTeam(allianceTeam *AllianceTeam) error {
	return database.allianceTeamTable.create(allianceTeam)
}

func (database *Database) GetTeamsByAlliance(allianceId int) ([]AllianceTeam, error) {
	var allianceTeams []AllianceTeam
	if err := database.allianceTeamTable.getAll(&allianceTeams); err != nil {
		return nil, err
	}
	sort.Slice(allianceTeams, func(i, j int) bool {
		return allianceTeams[i].PickPosition < allianceTeams[j].PickPosition
	})

	var matchingAllianceTeams []AllianceTeam
	for _, allianceTeam := range allianceTeams {
		if allianceTeam.AllianceId == allianceId {
			matchingAllianceTeams = append(matchingAllianceTeams, allianceTeam)
		}
	}
	return matchingAllianceTeams, nil
}

func (database *Database) UpdateAllianceTeam(allianceTeam *AllianceTeam) error {
	return database.allianceTeamTable.update(allianceTeam)
}

func (database *Database) DeleteAllianceTeam(id int) error {
	return database.allianceTeamTable.delete(id)
}

func (database *Database) TruncateAllianceTeams() error {
	return database.allianceTeamTable.truncate()
}

func (database *Database) GetAllAlliances() ([][]AllianceTeam, error) {
	var allianceTeams []AllianceTeam
	if err := database.allianceTeamTable.getAll(&allianceTeams); err != nil {
		return nil, err
	}
	sort.Slice(allianceTeams, func(i, j int) bool {
		if allianceTeams[i].AllianceId == allianceTeams[j].AllianceId {
			return allianceTeams[i].PickPosition < allianceTeams[j].PickPosition
		}
		return allianceTeams[i].AllianceId < allianceTeams[j].AllianceId
	})

	alliances := make([][]AllianceTeam, 0)
	// Format the sorted list of teams into a two-dimensional slice.
	currentAllianceId := -1
	for _, allianceTeam := range allianceTeams {
		if allianceTeam.AllianceId != currentAllianceId {
			currentAllianceId = allianceTeam.AllianceId
			alliances = append(alliances, make([]AllianceTeam, 0))
		}
		alliances[len(alliances)-1] = append(alliances[len(alliances)-1], allianceTeam)
	}
	return alliances, nil
}
