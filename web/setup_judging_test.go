// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestSetupJudging(t *testing.T) {
	web := setupTestWeb(t)

	// Check that the page renders.
	recorder := web.getHttpResponse("/setup/judging")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Judge Scheduling")

	// Generate teams and matches to test against.
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 1}))
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 2}))
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 3}))
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 4}))
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 5}))
	assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 6}))
	match := model.Match{
		Type:      model.Qualification,
		TypeOrder: 1,
		Time:      time.Now().Add(1 * time.Hour),
		Red1:      1,
		Red2:      2,
		Red3:      3,
		Blue1:     4,
		Blue2:     5,
		Blue3:     6,
	}
	assert.Nil(t, web.arena.Database.CreateMatch(&match))
	match = model.Match{
		Type:      model.Qualification,
		TypeOrder: 2,
		Time:      time.Now().Add(2 * time.Hour),
		Red1:      6,
		Red2:      5,
		Red3:      4,
		Blue1:     3,
		Blue2:     2,
		Blue3:     1,
	}
	assert.Nil(t, web.arena.Database.CreateMatch(&match))

	// Generate a judging schedule with valid parameters.
	params := url.Values{}
	params.Set("numJudges", "3")
	params.Set("durationMinutes", "10")
	params.Set("previousSpacingMinutes", "15")
	params.Set("nextSpacingMinutes", "15")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 303, recorder.Code)

	// Verify that judging slots were created.
	slots, err := web.arena.Database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.NotEmpty(t, slots)

	// Try to generate another judging schedule when one already exists.
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "already exists")

	// Clear the judging schedule.
	recorder = web.postHttpResponse("/setup/judging/clear", "")
	assert.Equal(t, 303, recorder.Code)

	// Verify that judging slots were cleared.
	slots, err = web.arena.Database.GetAllJudgingSlots()
	assert.Nil(t, err)
	assert.Empty(t, slots)

	// Test with invalid parameters.
	params = url.Values{}
	params.Set("numJudges", "0")
	params.Set("durationMinutes", "10")
	params.Set("previousSpacingMinutes", "15")
	params.Set("nextSpacingMinutes", "15")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Number of judges must be a positive integer")

	params.Set("numJudges", "3")
	params.Set("durationMinutes", "invalid")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Visit duration must be a positive integer")

	params.Set("durationMinutes", "10")
	params.Set("previousSpacingMinutes", "-5")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Minimum spacing after previous match must be a positive integer")

	params.Set("previousSpacingMinutes", "15")
	params.Set("nextSpacingMinutes", "0")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Minimum spacing before next match must be a positive integer")

	// Delete all qualification matches and verify error.
	assert.Nil(t, web.arena.Database.TruncateMatches())
	params.Set("nextSpacingMinutes", "15")
	recorder = web.postHttpResponse("/setup/judging/generate", params.Encode())
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No qualification matches found")
}
