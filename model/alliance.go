// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a playoff alliance.

package model

import "sort"

type Alliance struct {
	Id      int `db:"id,manual"`
	TeamIds []int
	Lineup  [3]int
}

func (database *Database) CreateAlliance(alliance *Alliance) error {
	return database.allianceTable.create(alliance)
}

func (database *Database) GetAllianceById(id int) (*Alliance, error) {
	return database.allianceTable.getById(id)
}

func (database *Database) UpdateAlliance(alliance *Alliance) error {
	return database.allianceTable.update(alliance)
}

func (database *Database) DeleteAlliance(id int) error {
	return database.allianceTable.delete(id)
}

func (database *Database) TruncateAlliances() error {
	return database.allianceTable.truncate()
}

func (database *Database) GetAllAlliances() ([]Alliance, error) {
	alliances, err := database.allianceTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(alliances, func(i, j int) bool {
		return alliances[i].Id < alliances[j].Id
	})
	return alliances, nil
}

// Updates the alliance, if necessary, to include whoever played in the match, in case there was a substitute.
func (database *Database) UpdateAllianceFromMatch(allianceId int, matchTeamIds [3]int) error {
	alliance, err := database.GetAllianceById(allianceId)
	if err != nil {
		return err
	}

	changed := false
	if matchTeamIds != alliance.Lineup {
		alliance.Lineup = matchTeamIds
		changed = true
	}

	for _, teamId := range matchTeamIds {
		found := false
		for _, allianceTeamId := range alliance.TeamIds {
			if teamId == allianceTeamId {
				found = true
				break
			}
		}
		if !found {
			alliance.TeamIds = append(alliance.TeamIds, teamId)
			changed = true
		}
	}

	if changed {
		return database.UpdateAlliance(alliance)
	}
	return nil
}

// Returns two arrays containing the IDs of any teams for the red and blue alliances, respectively, who are part of the
// playoff alliance but are not playing in the given match.
// If the given match isn't a playoff match, empty arrays are returned.
func (database *Database) GetOffFieldTeamIds(match *Match) ([]int, []int, error) {
	redOffFieldTeams, err := database.getOffFieldTeamIdsForAlliance(
		match.PlayoffRedAlliance, match.Red1, match.Red2, match.Red3,
	)
	if err != nil {
		return nil, nil, err
	}

	blueOffFieldTeams, err := database.getOffFieldTeamIdsForAlliance(
		match.PlayoffBlueAlliance, match.Blue1, match.Blue2, match.Blue3,
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

	alliance, err := database.GetAllianceById(allianceId)
	if alliance == nil || err != nil {
		return nil, err
	}
	offFieldTeamIds := []int{}
	for _, allianceTeamId := range alliance.TeamIds {
		if allianceTeamId != teamId1 && allianceTeamId != teamId2 && allianceTeamId != teamId3 {
			offFieldTeamIds = append(offFieldTeamIds, allianceTeamId)
		}
	}
	return offFieldTeamIds, nil
}
