// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestNonExistentSchedule(t *testing.T) {
	teams := make([]Team, 6)
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 2, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	expectedErr := "No schedule exists for 6 teams and 2 matches"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}
}

func TestMalformedSchedule(t *testing.T) {
	scheduleFile, _ := os.Create("schedules/6_1.csv")
	defer os.Remove("schedules/6_1.csv")
	scheduleFile.WriteString("1,0,2,0,3,0,4,0,5,0,6,0\n6,0,5,0,4,0,3,0,2,0,1,0\n")
	scheduleFile.Close()
	teams := make([]Team, 6)
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 1, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	expectedErr := "Schedule file contains 2 matches, expected 1"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}

	os.Remove("schedules/6_1.csv")
	scheduleFile, _ = os.Create("schedules/6_1.csv")
	scheduleFile.WriteString("1,0,asdf,0,3,0,4,0,5,0,6,0\n")
	scheduleFile.Close()
	_, err = BuildRandomSchedule(teams, scheduleBlocks, "test")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "strconv.ParseInt")
	}
}

func TestScheduleTeams(t *testing.T) {
	rand.Seed(0)

	numTeams := 18
	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 6, 60}}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	assert.Nil(t, err)
	assert.Equal(t, Match{1, "test", "1", time.Unix(0, 0).UTC(), 107, false, 102, false, 117, false, 115,
		false, 106, false, 116, false, "", time.Unix(0, 0).UTC()}, matches[0])
	assert.Equal(t, Match{2, "test", "2", time.Unix(60, 0).UTC(), 109, false, 113, false, 104, false, 108,
		false, 112, false, 118, false, "", time.Unix(0, 0).UTC()}, matches[1])
	assert.Equal(t, Match{3, "test", "3", time.Unix(120, 0).UTC(), 103, false, 111, false, 105, false, 114,
		false, 101, false, 110, false, "", time.Unix(0, 0).UTC()}, matches[2])
	assert.Equal(t, Match{4, "test", "4", time.Unix(180, 0).UTC(), 102, false, 104, false, 115, false, 117,
		false, 118, false, 113, false, "", time.Unix(0, 0).UTC()}, matches[3])
	assert.Equal(t, Match{5, "test", "5", time.Unix(240, 0).UTC(), 108, false, 114, false, 106, false, 103,
		false, 101, false, 116, false, "", time.Unix(0, 0).UTC()}, matches[4])
	assert.Equal(t, Match{6, "test", "6", time.Unix(300, 0).UTC(), 109, false, 112, false, 111, false, 110,
		false, 107, false, 105, false, "", time.Unix(0, 0).UTC()}, matches[5])
}

func TestScheduleTiming(t *testing.T) {
	teams := make([]Team, 18)
	scheduleBlocks := []ScheduleBlock{{time.Unix(100, 0).UTC(), 10, 75},
		{time.Unix(20000, 0).UTC(), 5, 1000},
		{time.Unix(100000, 0).UTC(), 15, 29}}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(100, 0).UTC(), matches[0].Time)
	assert.Equal(t, time.Unix(775, 0).UTC(), matches[9].Time)
	assert.Equal(t, time.Unix(20000, 0).UTC(), matches[10].Time)
	assert.Equal(t, time.Unix(24000, 0).UTC(), matches[14].Time)
	assert.Equal(t, time.Unix(100000, 0).UTC(), matches[15].Time)
	assert.Equal(t, time.Unix(100406, 0).UTC(), matches[29].Time)
}

func TestScheduleSurrogates(t *testing.T) {
	rand.Seed(0)

	numTeams := 38
	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 64, 60}}
	matches, _ := BuildRandomSchedule(teams, scheduleBlocks, "test")
	for i, match := range matches {
		if i == 13 || i == 14 {
			if !match.Red1IsSurrogate || match.Red2IsSurrogate || match.Red3IsSurrogate ||
				!match.Blue1IsSurrogate || match.Blue2IsSurrogate || match.Blue3IsSurrogate {
				t.Errorf("Surrogates wrong for match %d", i+1)
			}
		} else {
			if match.Red1IsSurrogate || match.Red2IsSurrogate || match.Red3IsSurrogate ||
				match.Blue1IsSurrogate || match.Blue2IsSurrogate || match.Blue3IsSurrogate {
				t.Errorf("Expected match %d to be free of surrogates", i+1)
			}
		}
	}
}
