// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a schedule block at an event.

package model

import (
	"time"
)

type ScheduleBlock struct {
	Id              int
	MatchType       string
	StartTime       time.Time
	NumMatches      int
	MatchSpacingSec int
}

func (database *Database) CreateScheduleBlock(block *ScheduleBlock) error {
	return database.scheduleBlockMap.Insert(block)
}

func (database *Database) GetScheduleBlocksByMatchType(matchType string) ([]ScheduleBlock, error) {
	var blocks []ScheduleBlock
	err := database.scheduleBlockMap.Select(&blocks, "SELECT * FROM schedule_blocks WHERE matchtype = ? ORDER BY "+
		"starttime ", matchType)
	return blocks, err
}

func (database *Database) DeleteScheduleBlocksByMatchType(matchType string) error {
	_, err := database.scheduleBlockMap.Exec("DELETE FROM schedule_blocks WHERE matchtype = ?", matchType)
	return err
}

func (database *Database) TruncateScheduleBlocks() error {
	return database.scheduleBlockMap.TruncateTables()
}
