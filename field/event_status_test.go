// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCycleTime(t *testing.T) {
	arena := setupTestArena(t)

	assert.Equal(t, "", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Time{})
	assert.Equal(t, "", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Now().Add(-125 * time.Second))
	assert.Equal(t, "", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Now())
	assert.Equal(t, "2:05", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Now().Add(3456 * time.Second))
	assert.Equal(t, "57:36", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Now().Add(5 * time.Hour))
	assert.Equal(t, "4:02:24", arena.EventStatus.CycleTime)
	arena.updateCycleTime(time.Now().Add(123*time.Hour + 1256*time.Second))
	assert.Equal(t, "118:20:56", arena.EventStatus.CycleTime)
}

func TestEarlyLateMessage(t *testing.T) {
	arena := setupTestArena(t)

	arena.LoadTestMatch()
	assert.Equal(t, "", arena.getEarlyLateMessage())

	arena.Database.CreateMatch(&model.Match{Type: model.Qualification, TypeOrder: 1})
	arena.Database.CreateMatch(&model.Match{Type: model.Qualification, TypeOrder: 2})
	matches, _ := arena.Database.GetMatchesByType(model.Qualification, false)
	assert.Equal(t, 2, len(matches))

	setMatch(arena.Database, &matches[0], time.Now().Add(300*time.Second), time.Time{}, false)
	arena.CurrentMatch = &matches[0]
	arena.MatchState = PreMatch
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(60*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(-60*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(-120*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(-180*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running 3 minutes late", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(181*time.Second), time.Now(), false)
	arena.MatchState = AutoPeriod
	assert.Equal(t, "Event is running 3 minutes early", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(-300*time.Second), time.Now().Add(-601*time.Second), false)
	setMatch(arena.Database, &matches[1], time.Now().Add(481*time.Second), time.Time{}, false)
	arena.MatchState = PostMatch
	assert.Equal(t, "Event is running 5 minutes early", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(181*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running 3 minutes early", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(-60*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(-180*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running 3 minutes late", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[0], time.Now().Add(-300*time.Second), time.Now().Add(-601*time.Second), true)
	assert.Equal(t, "", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(900*time.Second), time.Time{}, false)
	arena.CurrentMatch = &matches[1]
	arena.MatchState = PreMatch
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(899*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running 5 minutes early", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(60*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(-120*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running on schedule", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(-180*time.Second), time.Time{}, false)
	assert.Equal(t, "Event is running 3 minutes late", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now().Add(-180*time.Second), time.Now().Add(-541*time.Second), false)
	arena.MatchState = TeleopPeriod
	assert.Equal(t, "Event is running 6 minutes early", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now(), time.Now().Add(481*time.Second), false)
	arena.MatchState = PostMatch
	assert.Equal(t, "Event is running 8 minutes late", arena.getEarlyLateMessage())

	setMatch(arena.Database, &matches[1], time.Now(), time.Now().Add(481*time.Second), true)
	assert.Equal(t, "", arena.getEarlyLateMessage())

	// Check other match types.
	arena.MatchState = PreMatch
	arena.CurrentMatch = &model.Match{Type: model.Practice, Time: time.Now().Add(-181 * time.Second)}
	assert.Equal(t, "Event is running 3 minutes late", arena.getEarlyLateMessage())

	arena.CurrentMatch = &model.Match{Type: model.Playoff, Time: time.Now().Add(-181 * time.Second)}
	assert.Equal(t, "Event is running 3 minutes late", arena.getEarlyLateMessage())
}

func setMatch(database *model.Database, match *model.Match, matchTime time.Time, startedAt time.Time, isComplete bool) {
	match.Time = matchTime
	match.StartedAt = startedAt
	if isComplete {
		match.Status = game.TieMatch
	} else {
		match.Status = game.MatchScheduled
	}
	_ = database.UpdateMatch(match)
}
