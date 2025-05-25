// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a judging slot.

package model

import (
	"sort"
	"time"
)

type JudgingSlot struct {
	Id                  int `db:"id"`
	Time                time.Time
	TeamId              int
	PreviousMatchNumber int
	PreviousMatchTime   time.Time
	NextMatchNumber     int
	NextMatchTime       time.Time
	JudgeNumber         int
}

func (database *Database) CreateJudgingSlot(judgingSlot *JudgingSlot) error {
	return database.judgingSlotTable.create(judgingSlot)
}

func (database *Database) TruncateJudgingSlots() error {
	return database.judgingSlotTable.truncate()
}

func (database *Database) GetAllJudgingSlots() ([]JudgingSlot, error) {
	judgingSlots, err := database.judgingSlotTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(
		judgingSlots,
		func(i, j int) bool {
			return judgingSlots[i].TeamId < judgingSlots[j].TeamId
		},
	)
	return judgingSlots, nil
}
