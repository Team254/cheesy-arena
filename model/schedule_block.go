// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a schedule block at an event.

package model

import (
	"sort"
	"time"
)

type ScheduleBlock struct {
	Id              int `db:"id"`
	MatchType       MatchType
	StartTime       time.Time
	NumMatches      int
	MatchSpacingSec int
}

func (database *Database) CreateScheduleBlock(block *ScheduleBlock) error {
	return database.scheduleBlockTable.create(block)
}

func (database *Database) GetScheduleBlocksByMatchType(matchType MatchType) ([]ScheduleBlock, error) {
	scheduleBlocks, err := database.scheduleBlockTable.getAll()
	if err != nil {
		return nil, err
	}

	var matchingScheduleBlocks []ScheduleBlock
	for _, scheduleBlock := range scheduleBlocks {
		if scheduleBlock.MatchType == matchType {
			matchingScheduleBlocks = append(matchingScheduleBlocks, scheduleBlock)
		}
	}

	sort.Slice(matchingScheduleBlocks, func(i, j int) bool {
		return matchingScheduleBlocks[i].StartTime.Before(matchingScheduleBlocks[j].StartTime)
	})
	return matchingScheduleBlocks, nil
}

func (database *Database) DeleteScheduleBlocksByMatchType(matchType MatchType) error {
	scheduleBlocks, err := database.GetScheduleBlocksByMatchType(matchType)
	if err != nil {
		return err
	}

	for _, scheduleBlock := range scheduleBlocks {
		if err = database.scheduleBlockTable.delete(scheduleBlock.Id); err != nil {
			return err
		}
	}
	return nil
}

func (database *Database) TruncateScheduleBlocks() error {
	return database.scheduleBlockTable.truncate()
}
