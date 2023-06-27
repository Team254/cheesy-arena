// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a scheduled break at an event.

package model

import (
	"sort"
	"time"
)

type ScheduledBreak struct {
	Id              int `db:"id"`
	MatchType       MatchType
	TypeOrderBefore int
	Time            time.Time
	DurationSec     int
	Description     string
}

func (database *Database) CreateScheduledBreak(scheduledBreak *ScheduledBreak) error {
	return database.scheduledBreakTable.create(scheduledBreak)
}

func (database *Database) GetScheduledBreaksByMatchType(matchType MatchType) ([]ScheduledBreak, error) {
	scheduledBreaks, err := database.scheduledBreakTable.getAll()
	if err != nil {
		return nil, err
	}

	var matchingScheduledBreaks []ScheduledBreak
	for _, scheduledBreak := range scheduledBreaks {
		if scheduledBreak.MatchType == matchType {
			matchingScheduledBreaks = append(matchingScheduledBreaks, scheduledBreak)
		}
	}

	sort.Slice(matchingScheduledBreaks, func(i, j int) bool {
		return matchingScheduledBreaks[i].TypeOrderBefore < matchingScheduledBreaks[j].TypeOrderBefore
	})
	return matchingScheduledBreaks, nil
}

func (database *Database) GetScheduledBreakByMatchTypeOrder(
	matchType MatchType,
	typeOrder int,
) (*ScheduledBreak, error) {
	scheduledBreaks, err := database.GetScheduledBreaksByMatchType(matchType)
	if err != nil {
		return nil, err
	}

	for _, scheduledBreak := range scheduledBreaks {
		if scheduledBreak.TypeOrderBefore == typeOrder {
			return &scheduledBreak, nil
		}
	}
	return nil, nil
}

func (database *Database) DeleteScheduledBreaksByMatchType(matchType MatchType) error {
	scheduledBreaks, err := database.GetScheduledBreaksByMatchType(matchType)
	if err != nil {
		return err
	}

	for _, scheduledBreak := range scheduledBreaks {
		if err = database.scheduledBreakTable.delete(scheduledBreak.Id); err != nil {
			return err
		}
	}
	return nil
}

func (database *Database) TruncateScheduledBreaks() error {
	return database.scheduledBreakTable.truncate()
}
