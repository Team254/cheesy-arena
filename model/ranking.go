// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package model

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/game"
)

type RankingDb struct {
	TeamId            int
	Rank              int
	RankingFieldsJson string
}

func (database *Database) CreateRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	return database.rankingMap.Insert(rankingDb)
}

func (database *Database) GetRankingForTeam(teamId int) (*game.Ranking, error) {
	rankingDb := new(RankingDb)
	err := database.rankingMap.Get(rankingDb, teamId)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return nil, nil
	}
	ranking, err := rankingDb.deserialize()
	if err != nil {
		return nil, err
	}
	return ranking, err
}

func (database *Database) SaveRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	_, err = database.rankingMap.Update(rankingDb)
	return err
}

func (database *Database) DeleteRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	_, err = database.rankingMap.Delete(rankingDb)
	return err
}

func (database *Database) TruncateRankings() error {
	return database.rankingMap.TruncateTables()
}

func (database *Database) GetAllRankings() ([]game.Ranking, error) {
	var rankingDbs []RankingDb
	err := database.rankingMap.Select(&rankingDbs, "SELECT * FROM rankings ORDER BY rank")
	if err != nil {
		return nil, err
	}
	var rankings []game.Ranking
	for _, rankingDb := range rankingDbs {
		ranking, err := rankingDb.deserialize()
		if err != nil {
			return nil, err
		}
		rankings = append(rankings, *ranking)
	}
	return rankings, err
}

// Deletes the existing rankings and inserts the given ones as a replacement, in a single transaction.
func (database *Database) ReplaceAllRankings(rankings game.Rankings) error {
	transaction, err := database.rankingMap.Begin()
	if err != nil {
		return err
	}

	_, err = transaction.Exec("DELETE FROM rankings")
	if err != nil {
		transaction.Rollback()
		return err
	}

	for _, ranking := range rankings {
		rankingDb, err := serializeRanking(ranking)
		if err != nil {
			transaction.Rollback()
			return err
		}
		err = transaction.Insert(rankingDb)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}

	return transaction.Commit()
}

// Converts the nested struct MatchResult to the DB version that has JSON fields.
func serializeRanking(ranking *game.Ranking) (*RankingDb, error) {
	rankingDb := RankingDb{TeamId: ranking.TeamId, Rank: ranking.Rank}
	if err := serializeHelper(&rankingDb.RankingFieldsJson, ranking.RankingFields); err != nil {
		return nil, err
	}
	return &rankingDb, nil
}

// Converts the DB Ranking with JSON fields to the nested struct version.
func (rankingDb *RankingDb) deserialize() (*game.Ranking, error) {
	ranking := game.Ranking{TeamId: rankingDb.TeamId, Rank: rankingDb.Rank}
	if err := json.Unmarshal([]byte(rankingDb.RankingFieldsJson), &ranking.RankingFields); err != nil {
		return nil, err
	}
	return &ranking, nil
}
