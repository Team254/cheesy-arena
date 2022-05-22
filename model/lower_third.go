// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for the text on a lower third slide.

package model

import (
	"sort"
)

type LowerThird struct {
	Id           int `db:"id"`
	TopText      string
	BottomText   string
	DisplayOrder int
	AwardId      int
}

func (database *Database) CreateLowerThird(lowerThird *LowerThird) error {
	return database.lowerThirdTable.create(lowerThird)
}

func (database *Database) GetLowerThirdById(id int) (*LowerThird, error) {
	return database.lowerThirdTable.getById(id)
}

func (database *Database) UpdateLowerThird(lowerThird *LowerThird) error {
	return database.lowerThirdTable.update(lowerThird)
}

func (database *Database) DeleteLowerThird(id int) error {
	return database.lowerThirdTable.delete(id)
}

func (database *Database) TruncateLowerThirds() error {
	return database.lowerThirdTable.truncate()
}

func (database *Database) GetAllLowerThirds() ([]LowerThird, error) {
	lowerThirds, err := database.lowerThirdTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(lowerThirds, func(i, j int) bool {
		return lowerThirds[i].DisplayOrder < lowerThirds[j].DisplayOrder
	})
	return lowerThirds, nil
}

func (database *Database) GetLowerThirdsByAwardId(awardId int) ([]LowerThird, error) {
	lowerThirds, err := database.GetAllLowerThirds()
	if err != nil {
		return nil, err
	}

	var matchingLowerThirds []LowerThird
	for _, lowerThird := range lowerThirds {
		if lowerThird.AwardId == awardId {
			matchingLowerThirds = append(matchingLowerThirds, lowerThird)
		}
	}
	return matchingLowerThirds, nil
}

func (database *Database) GetNextLowerThirdDisplayOrder() int {
	lowerThirds, err := database.GetAllLowerThirds()
	if err != nil {
		return 0
	}
	if len(lowerThirds) == 0 {
		return 1
	}
	return lowerThirds[len(lowerThirds)-1].DisplayOrder + 1
}
