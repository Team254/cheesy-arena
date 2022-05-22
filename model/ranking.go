// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package model

import (
	"github.com/Team254/cheesy-arena/game"
	"sort"
)

func (database *Database) CreateRanking(ranking *game.Ranking) error {
	return database.rankingTable.create(ranking)
}

func (database *Database) GetRankingForTeam(teamId int) (*game.Ranking, error) {
	return database.rankingTable.getById(teamId)
}

func (database *Database) UpdateRanking(ranking *game.Ranking) error {
	return database.rankingTable.update(ranking)
}

func (database *Database) DeleteRanking(teamId int) error {
	return database.rankingTable.delete(teamId)
}

func (database *Database) TruncateRankings() error {
	return database.rankingTable.truncate()
}

func (database *Database) GetAllRankings() (game.Rankings, error) {
	rankings, err := database.rankingTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Rank < rankings[j].Rank
	})
	return rankings, nil
}

// Deletes the existing rankings and inserts the given ones as a replacement.
func (database *Database) ReplaceAllRankings(rankings game.Rankings) error {
	if err := database.rankingTable.truncate(); err != nil {
		return err
	}

	for _, ranking := range rankings {
		if err := database.CreateRanking(&ranking); err != nil {
			return err
		}
	}
	return nil
}
