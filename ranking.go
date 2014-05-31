// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package main

type Ranking struct {
	TeamId             int
	Rank               int
	QualificationScore int
	AssistPoints       int
	AutoPoints         int
	TrussCatchPoints   int
	GoalFoulPoints     int
	Random             float64
	Wins               int
	Losses             int
	Ties               int
	Disqualifications  int
	Played             int
}

func (database *Database) CreateRanking(ranking *Ranking) error {
	return database.rankingMap.Insert(ranking)
}

func (database *Database) GetRankingForTeam(teamId int) (*Ranking, error) {
	ranking := new(Ranking)
	err := database.rankingMap.Get(ranking, teamId)
	if err != nil && err.Error() == "sql: no rows in result set" {
		ranking = nil
		err = nil
	}
	return ranking, err
}

func (database *Database) SaveRanking(ranking *Ranking) error {
	_, err := database.rankingMap.Update(ranking)
	return err
}

func (database *Database) DeleteRanking(ranking *Ranking) error {
	_, err := database.rankingMap.Delete(ranking)
	return err
}

func (database *Database) TruncateRankings() error {
	return database.rankingMap.TruncateTables()
}

func (database *Database) GetAllRankings() ([]Ranking, error) {
	var rankings []Ranking
	err := database.rankingMap.Select(&rankings, "SELECT * FROM rankings ORDER BY rank")
	return rankings, err
}
