// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for an award.

package model

type Award struct {
	Id         int
	Type       AwardType
	AwardName  string
	TeamId     int
	PersonName string
}

type AwardType int

const (
	JudgedAward AwardType = iota
	FinalistAward
	WinnerAward
)

func (database *Database) CreateAward(award *Award) error {
	return database.awardMap.Insert(award)
}

func (database *Database) GetAwardById(id int) (*Award, error) {
	award := new(Award)
	err := database.awardMap.Get(award, id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		award = nil
		err = nil
	}
	return award, err
}

func (database *Database) SaveAward(award *Award) error {
	_, err := database.awardMap.Update(award)
	return err
}

func (database *Database) DeleteAward(award *Award) error {
	_, err := database.awardMap.Delete(award)
	return err
}

func (database *Database) TruncateAwards() error {
	return database.awardMap.TruncateTables()
}

func (database *Database) GetAllAwards() ([]Award, error) {
	var awards []Award
	err := database.awardMap.Select(&awards, "SELECT * FROM awards ORDER BY id")
	return awards, err
}

func (database *Database) GetAwardsByType(awardType AwardType) ([]Award, error) {
	var awards []Award
	err := database.awardMap.Select(&awards, "SELECT * FROM awards WHERE type = ? ORDER BY id", awardType)
	return awards, err
}
