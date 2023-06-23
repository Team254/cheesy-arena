// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetupSchedule(t *testing.T) {
	web := setupTestWeb(t)

	for i := 0; i < 38; i++ {
		web.arena.Database.CreateTeam(&model.Team{Id: i + 101})
	}
	web.arena.Database.CreateMatch(&model.Match{Type: model.Practice, ShortName: "P1"})

	// Check the default setting values.
	recorder := web.getHttpResponse("/setup/schedule?matchType=practice")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "addBlock();")

	// Submit a schedule for generation.
	postData := "numScheduleBlocks=3&startTime0=2014-01-01 09:00:00 AM&numMatches0=7&matchSpacingSec0=480&" +
		"startTime1=2014-01-02 09:56:00 AM&numMatches1=17&matchSpacingSec1=420&startTime2=2014-01-03 01:00:00 PM&" +
		"numMatches2=40&matchSpacingSec2=360&matchType=qualification"
	recorder = web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/schedule?matchType=qualification")
	assert.Contains(t, recorder.Body.String(), "2014-01-01 09:48:00") // Last match of first block.
	assert.Contains(t, recorder.Body.String(), "2014-01-02 11:48:00") // Last match of second block.
	assert.Contains(t, recorder.Body.String(), "2014-01-03 16:54:00") // Last match of third block.

	// Save schedule and check that it was persisted.
	recorder = web.postHttpResponse("/setup/schedule/save?matchType=qualification", "")
	assert.Equal(t, 303, recorder.Code)
	matches, err := web.arena.Database.GetMatchesByType(model.Qualification, true)
	assert.Nil(t, err)
	assert.Equal(t, 64, len(matches))
	location, _ := time.LoadLocation("Local")
	assert.Equal(t, time.Date(2014, 1, 1, 9, 0, 0, 0, location).Unix(), matches[0].Time.Unix())
	assert.Equal(t, time.Date(2014, 1, 2, 9, 56, 0, 0, location).Unix(), matches[7].Time.Unix())
	assert.Equal(t, time.Date(2014, 1, 3, 13, 0, 0, 0, location).Unix(), matches[24].Time.Unix())
}

func TestSetupScheduleErrors(t *testing.T) {
	web := setupTestWeb(t)

	// No teams.
	postData := "numScheduleBlocks=1&startTime0=2014-01-01 09:00:00 AM&numMatches0=7&matchSpacingSec0=480&" +
		"matchType=practice"
	recorder := web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No team list is configured.")

	// Insufficient number of teams.
	for i := 0; i < 5; i++ {
		web.arena.Database.CreateTeam(&model.Team{Id: i + 101})
	}
	postData = "numScheduleBlocks=1&startTime0=2014-01-01 09:00:00 AM&numMatches0=7&matchSpacingSec0=480&" +
		"matchType=practice"
	recorder = web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "There must be at least 6 teams to generate a schedule.")

	// More matches per team than schedules exist for.
	web.arena.Database.CreateTeam(&model.Team{Id: 118})
	postData = "numScheduleBlocks=1&startTime0=2014-01-01 09:00:00 AM&numMatches0=700&matchSpacingSec0=480&" +
		"matchType=practice"
	recorder = web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "No schedule template exists for 6 teams and 700 matches")

	// Incomplete scheduling data received.
	postData = "numScheduleBlocks=1&startTime0=2014-01-01 09:00:00 AM&numMatches0=&matchSpacingSec0=480&" +
		"matchType=practice"
	recorder = web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Incomplete or invalid schedule block parameters specified.")

	// Previous schedule already exists.
	for i := 18; i < 38; i++ {
		web.arena.Database.CreateTeam(&model.Team{Id: i + 101})
	}
	web.arena.Database.CreateMatch(&model.Match{Type: model.Practice, ShortName: "P1"})
	web.arena.Database.CreateMatch(&model.Match{Type: model.Practice, ShortName: "P2"})
	postData = "numScheduleBlocks=1&startTime0=2014-01-01 09:00:00 AM&numMatches0=64&matchSpacingSec0=480&" +
		"matchType=practice"
	recorder = web.postHttpResponse("/setup/schedule/generate", postData)
	assert.Equal(t, 303, recorder.Code)
	recorder = web.postHttpResponse("/setup/schedule/save", postData)
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "schedule of 2 Practice matches already exists")
}
