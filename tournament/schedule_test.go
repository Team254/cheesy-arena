// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNonExistentSchedule(t *testing.T) {
	teams := make([]model.Team, 5)
	scheduleBlocks := []model.ScheduleBlock{{0, model.Test, time.Unix(0, 0).UTC(), 2, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, model.Test)
	expectedErr := "No schedule template exists for 5 teams and 2 matches"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}
}

func TestMalformedSchedule(t *testing.T) {
	filename := fmt.Sprintf("%s/5_1.csv", filepath.Join(model.BaseDir, schedulesDir))
	scheduleFile, _ := os.Create(filename)
	defer os.Remove(filename)
	scheduleFile.WriteString("1,0,2,0,3,0,4,0,5,0,6,0\n6,0,5,0,4,0,3,0,2,0,1,0\n")
	scheduleFile.Close()
	teams := make([]model.Team, 5)
	scheduleBlocks := []model.ScheduleBlock{{0, model.Test, time.Unix(0, 0).UTC(), 1, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, model.Test)
	expectedErr := "Schedule file contains 2 matches, expected 1"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}

	os.Remove(filename)
	scheduleFile, _ = os.Create(filename)
	scheduleFile.WriteString("1,0,asdf,0,3,0,4,0,5,0,6,0\n")
	scheduleFile.Close()
	_, err = BuildRandomSchedule(teams, scheduleBlocks, model.Test)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "strconv.Atoi")
	}
}

func TestScheduleTeams(t *testing.T) {
	rand.Seed(0)

	numTeams := 18
	teams := make([]model.Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []model.ScheduleBlock{{0, model.Practice, time.Unix(0, 0).UTC(), 6, 60}}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, model.Practice)
	assert.Nil(t, err)
	assertMatch(t, matches[0], model.Practice, 1, 0, "P1", "Practice 1", 115, 111, 108, 109, 116, 117)
	assertMatch(t, matches[1], model.Practice, 2, 60, "P2", "Practice 2", 114, 112, 103, 101, 104, 118)
	assertMatch(t, matches[2], model.Practice, 3, 120, "P3", "Practice 3", 110, 107, 105, 106, 113, 102)
	assertMatch(t, matches[3], model.Practice, 4, 180, "P4", "Practice 4", 112, 108, 109, 101, 111, 103)
	assertMatch(t, matches[4], model.Practice, 5, 240, "P5", "Practice 5", 113, 117, 115, 110, 114, 102)
	assertMatch(t, matches[5], model.Practice, 6, 300, "P6", "Practice 6", 118, 105, 106, 107, 104, 116)

	// Check with excess room for matches in the schedule.
	scheduleBlocks = []model.ScheduleBlock{{0, model.Practice, time.Unix(0, 0).UTC(), 7, 60}}
	matches, err = BuildRandomSchedule(teams, scheduleBlocks, model.Practice)
	assert.Nil(t, err)
}

func TestScheduleTiming(t *testing.T) {
	teams := make([]model.Team, 18)
	scheduleBlocks := []model.ScheduleBlock{
		{0, model.Qualification, time.Unix(100, 0).UTC(), 10, 75},
		{0, model.Qualification, time.Unix(20000, 0).UTC(), 5, 1000},
		{0, model.Qualification, time.Unix(100000, 0).UTC(), 15, 29},
	}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, model.Qualification)
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
	teams := make([]model.Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []model.ScheduleBlock{{0, model.Qualification, time.Unix(0, 0).UTC(), 64, 60}}
	matches, _ := BuildRandomSchedule(teams, scheduleBlocks, model.Qualification)
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

func assertMatch(
	t *testing.T,
	match model.Match,
	matchType model.MatchType,
	typeOrder int,
	timeInSec int64,
	shortName, longName string,
	red1, red2, red3, blue1, blue2, blue3 int,
) {
	assert.Equal(t, matchType, match.Type)
	assert.Equal(t, typeOrder, match.TypeOrder)
	assert.Equal(t, time.Unix(timeInSec, 0).UTC(), match.Time)
	assert.Equal(t, shortName, match.ShortName)
	assert.Equal(t, longName, match.LongName)
	assert.Equal(t, "", match.NameDetail)
	assert.Equal(t, 0, match.PlayoffRedAlliance)
	assert.Equal(t, 0, match.PlayoffBlueAlliance)
	assert.Equal(t, red1, match.Red1)
	assert.Equal(t, red2, match.Red2)
	assert.Equal(t, red3, match.Red3)
	assert.Equal(t, blue1, match.Blue1)
	assert.Equal(t, blue2, match.Blue2)
	assert.Equal(t, blue3, match.Blue3)
	assert.Equal(t, "qm", match.TbaMatchKey.CompLevel)
	assert.Equal(t, 0, match.TbaMatchKey.SetNumber)
	assert.Equal(t, typeOrder, match.TbaMatchKey.MatchNumber)
}
