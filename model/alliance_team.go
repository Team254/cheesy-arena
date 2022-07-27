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
	allianceTeams, err := database.allianceTeamTable.getAll()
	if err != nil {
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
	allianceTeams, err := database.allianceTeamTable.getAll()
	if err != nil {
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

// Returns two arrays containing the IDs of any teams for the red and blue alliances, respectively, who are part of the
// elimination alliance but are not playing in the given match.
// If the given match isn't an elimination match, empty arrays are returned.
func (database *Database) GetOffFieldTeamIds(match *Match) ([]int, []int, error) {
	redOffFieldTeams, err := database.getOffFieldTeamIdsForAlliance(
		match.ElimRedAlliance, match.Red1, match.Red2, match.Red3,
	)
	if err != nil {
		return nil, nil, err
	}

	blueOffFieldTeams, err := database.getOffFieldTeamIdsForAlliance(
		match.ElimBlueAlliance, match.Blue1, match.Blue2, match.Blue3,
	)
	if err != nil {
		return nil, nil, err
	}

	return redOffFieldTeams, blueOffFieldTeams, nil
}

func (database *Database) getOffFieldTeamIdsForAlliance(allianceId int, teamId1, teamId2, teamId3 int) ([]int, error) {
	if allianceId == 0 {
		return []int{}, nil
	}

	allianceTeams, err := database.GetTeamsByAlliance(allianceId)
	if err != nil {
		return nil, err
	}
	offFieldTeamIds := []int{}
	for _, allianceTeam := range allianceTeams {
		if allianceTeam.TeamId != teamId1 && allianceTeam.TeamId != teamId2 && allianceTeam.TeamId != teamId3 {
			offFieldTeamIds = append(offFieldTeamIds, allianceTeam.TeamId)
		}
	}
	return offFieldTeamIds, nil
}
