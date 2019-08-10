// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for the text on a lower third slide.

package model

type LowerThird struct {
	Id           int
	TopText      string
	BottomText   string
	DisplayOrder int
	AwardId      int
}

func (database *Database) CreateLowerThird(lowerThird *LowerThird) error {
	return database.lowerThirdMap.Insert(lowerThird)
}

func (database *Database) GetLowerThirdById(id int) (*LowerThird, error) {
	lowerThird := new(LowerThird)
	err := database.lowerThirdMap.Get(lowerThird, id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		lowerThird = nil
		err = nil
	}
	return lowerThird, err
}

func (database *Database) SaveLowerThird(lowerThird *LowerThird) error {
	_, err := database.lowerThirdMap.Update(lowerThird)
	return err
}

func (database *Database) DeleteLowerThird(lowerThird *LowerThird) error {
	_, err := database.lowerThirdMap.Delete(lowerThird)
	return err
}

func (database *Database) TruncateLowerThirds() error {
	return database.lowerThirdMap.TruncateTables()
}

func (database *Database) GetAllLowerThirds() ([]LowerThird, error) {
	var lowerThirds []LowerThird
	err := database.lowerThirdMap.Select(&lowerThirds, "SELECT * FROM lower_thirds ORDER BY displayorder")
	return lowerThirds, err
}

func (database *Database) GetLowerThirdsByAwardId(awardId int) ([]LowerThird, error) {
	var lowerThirds []LowerThird
	err := database.lowerThirdMap.Select(&lowerThirds, "SELECT * FROM lower_thirds WHERE awardid = ? ORDER BY id",
		awardId)
	return lowerThirds, err
}

func (database *Database) GetNextLowerThirdDisplayOrder() int {
	var count int
	_ = database.lowerThirdMap.SelectOne(&count, "SELECT MAX(displayorder) + 1 FROM lower_thirds")
	return count
}
