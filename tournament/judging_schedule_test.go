// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package tournament

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestBuildJudgingSchedule(t *testing.T) {
	rand.Seed(0)
	database := setupTestDb(t)

	// Test error when judging slots already exist.
	slot := model.JudgingSlot{Time: time.Now(), TeamId: 254, JudgeNumber: 1}
	assert.Nil(t, database.CreateJudgingSlot(&slot))
	params := JudgingScheduleParams{
		NumJudges:              3,
		DurationMinutes:        23,
		PreviousSpacingMinutes: 17,
		NextSpacingMinutes:     14,
	}
	err := BuildJudgingSchedule(database, params)
	assert.Contains(t, err.Error(), "existing judging slots found")

	assert.Nil(t, database.TruncateJudgingSlots())

	// Test error when no teams present.
	err = BuildJudgingSchedule(database, params)
	assert.Contains(t, err.Error(), "no teams present")

	// Generate teams to test against.
	for i := 1; i <= 24; i++ {
		assert.Nil(t, database.CreateTeam(&model.Team{Id: i}))
	}
	teams, err := database.GetAllTeams()
	assert.Nil(t, err)

	// Test error when no qualification matches found.
	err = BuildJudgingSchedule(database, params)
	assert.Contains(t, err.Error(), "no qualification matches found")

	// Generate qualification schedule to test against.
	scheduleBlocks := []model.ScheduleBlock{
		{
			MatchType:       model.Qualification,
			StartTime:       time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC),
			NumMatches:      12,
			MatchSpacingSec: 600,
		},
		{
			MatchType:       model.Qualification,
			StartTime:       time.Date(2025, 4, 1, 13, 0, 0, 0, time.UTC),
			NumMatches:      12,
			MatchSpacingSec: 600,
		},
	}
	for _, block := range scheduleBlocks {
		assert.Nil(t, database.CreateScheduleBlock(&block))
	}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, model.Qualification)
	assert.Nil(t, err)
	for _, match := range matches {
		assert.Nil(t, database.CreateMatch(&match))
	}

	err = BuildJudgingSchedule(database, params)
	assert.Nil(t, err)
	slots, err := database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.Equal(t, 24, len(slots))
	judgeTeamCounts := make(map[int]int)
	for _, slot := range slots {
		assert.NotEqual(t, 0, slot.TeamId)
		assert.NotEqual(t, 0, slot.JudgeNumber)
		judgeTeamCounts[slot.JudgeNumber]++

		// Check that the slot is not too close to the previous or next matches.
		if slot.PreviousMatchNumber > 0 {
			spacing := slot.Time.Sub(slot.PreviousMatchTime).Minutes()
			assert.GreaterOrEqual(t, spacing, float64(params.PreviousSpacingMinutes))
		}
		if slot.NextMatchNumber > 0 {
			spacing := slot.NextMatchTime.Sub(slot.Time).Minutes() - float64(params.DurationMinutes)
			assert.GreaterOrEqual(t, spacing, float64(params.NextSpacingMinutes))
		}

		// Check that the slot is not scheduled during the break.
		breakStartTime := scheduleBlocks[0].StartTime.Add(
			time.Duration(scheduleBlocks[0].NumMatches*scheduleBlocks[0].MatchSpacingSec) * time.Second,
		)
		if slot.Time.Before(scheduleBlocks[1].StartTime) {
			assert.True(t, slot.Time.Before(breakStartTime))
		}
		if slot.Time.After(breakStartTime) {
			assert.True(t, slot.Time.After(scheduleBlocks[1].StartTime))
		}
	}
	if assert.Equal(t, 3, len(judgeTeamCounts)) {
		assert.Equal(t, 8, judgeTeamCounts[1])
		assert.Equal(t, 8, judgeTeamCounts[2])
		assert.Equal(t, 8, judgeTeamCounts[3])
	}
}

func TestBuildJudgingScheduleMissingTeamMatches(t *testing.T) {
	rand.Seed(0)
	database := setupTestDb(t)

	for i := 1; i <= 6; i++ {
		assert.Nil(t, database.CreateTeam(&model.Team{Id: i}))
	}
	teams, err := database.GetAllTeams()
	assert.Nil(t, err)

	scheduleBlocks := []model.ScheduleBlock{
		{
			MatchType:       model.Qualification,
			StartTime:       time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC),
			NumMatches:      2,
			MatchSpacingSec: 600,
		},
	}
	for _, block := range scheduleBlocks {
		assert.Nil(t, database.CreateScheduleBlock(&block))
	}

	// Only include the first three teams in matches to force missing team matches.
	for i := 0; i < 2; i++ {
		match := model.Match{
			Type:      model.Qualification,
			TypeOrder: i + 1,
			Time:      scheduleBlocks[0].StartTime.Add(time.Duration(i*scheduleBlocks[0].MatchSpacingSec) * time.Second),
			Red1:      teams[0].Id,
			Red2:      teams[1].Id,
			Red3:      teams[2].Id,
			Blue1:     teams[0].Id,
			Blue2:     teams[1].Id,
			Blue3:     teams[2].Id,
		}
		assert.Nil(t, database.CreateMatch(&match))
	}

	params := JudgingScheduleParams{
		NumJudges:              2,
		DurationMinutes:        20,
		PreviousSpacingMinutes: 15,
		NextSpacingMinutes:     10,
	}
	err = BuildJudgingSchedule(database, params)
	assert.Contains(t, err.Error(), "no qualification matches for team")
}

func TestGetNextSlotForTeamAfterLastMatch(t *testing.T) {
	params := JudgingScheduleParams{
		NumJudges:              1,
		DurationMinutes:        20,
		PreviousSpacingMinutes: 30,
		NextSpacingMinutes:     10,
	}
	team := model.Team{Id: 1}
	matches := []model.Match{
		{
			Type:      model.Qualification,
			TypeOrder: 1,
			Time:      time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC),
			Red1:      team.Id,
		},
		{
			Type:      model.Qualification,
			TypeOrder: 2,
			Time:      time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC),
			Red1:      team.Id,
		},
	}
	candidateTime := time.Date(2025, 4, 1, 11, 0, 0, 0, time.UTC)

	slot, err := getNextSlotForTeam(team, candidateTime, matches, params)
	assert.Nil(t, err)
	assert.Equal(t, candidateTime, slot.Time)
}

func TestBuildJudgingScheduleAllowsEndDuringBreak(t *testing.T) {
	rand.Seed(0)
	database := setupTestDb(t)

	assert.Nil(t, database.CreateTeam(&model.Team{Id: 1}))

	scheduleBlocks := []model.ScheduleBlock{
		{
			MatchType:       model.Qualification,
			StartTime:       time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC),
			NumMatches:      2,
			MatchSpacingSec: 600,
		},
		{
			MatchType:       model.Qualification,
			StartTime:       time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC),
			NumMatches:      1,
			MatchSpacingSec: 600,
		},
	}
	for _, block := range scheduleBlocks {
		assert.Nil(t, database.CreateScheduleBlock(&block))
	}

	matches := []model.Match{
		{
			Type:      model.Qualification,
			TypeOrder: 1,
			Time:      scheduleBlocks[0].StartTime,
			Red1:      1,
			Red2:      1,
			Red3:      1,
			Blue1:     1,
			Blue2:     1,
			Blue3:     1,
		},
		{
			Type:      model.Qualification,
			TypeOrder: 2,
			Time:      scheduleBlocks[0].StartTime.Add(10 * time.Minute),
			Red1:      1,
			Red2:      1,
			Red3:      1,
			Blue1:     1,
			Blue2:     1,
			Blue3:     1,
		},
		{
			Type:      model.Qualification,
			TypeOrder: 3,
			Time:      scheduleBlocks[1].StartTime,
			Red1:      1,
			Red2:      1,
			Red3:      1,
			Blue1:     1,
			Blue2:     1,
			Blue3:     1,
		},
	}
	for _, match := range matches {
		assert.Nil(t, database.CreateMatch(&match))
	}

	params := JudgingScheduleParams{
		NumJudges:              1,
		DurationMinutes:        20,
		PreviousSpacingMinutes: 0,
		NextSpacingMinutes:     0,
	}
	err := BuildJudgingSchedule(database, params)
	assert.Nil(t, err)

	slots, err := database.GetAllJudgingSlots()
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(slots)) {
		assert.Equal(t, scheduleBlocks[0].StartTime.Add(10*time.Minute), slots[0].Time)
	}
}
