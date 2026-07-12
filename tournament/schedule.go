// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for creating practice and qualification match schedules.

package tournament

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/model"
)

const (
	schedulesDir  = "schedules"
	TeamsPerMatch = 6
)

var schedulePerm = rand.Perm

// Creates a random schedule for the given parameters and returns it as a list of matches.
func BuildRandomSchedule(
	teams []model.Team, scheduleBlocks []model.ScheduleBlock, matchType model.MatchType,
) ([]model.Match, error) {
	// Load the anonymized, pre-randomized match schedule for the given number of teams and matches per team.
	numTeams := len(teams)
	numMatches := countMatches(scheduleBlocks)
	matchesPerTeam := int(float32(numMatches*TeamsPerMatch) / float32(numTeams))

	// Keep the number of matches equal to the total requested in the schedule blocks.

	file, err := os.Open(
		fmt.Sprintf("%s/%d_%d.csv", filepath.Join(model.BaseDir, schedulesDir), numTeams, matchesPerTeam),
	)
	if err != nil {
		return nil, fmt.Errorf("No schedule template exists for %d teams and %d matches", numTeams, matchesPerTeam)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	csvLines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(csvLines) != numMatches {
		return nil, fmt.Errorf("Schedule file contains %d matches, expected %d", len(csvLines), numMatches)
	}

	// Convert string fields from schedule to integers.
	anonSchedule := make([][12]int, numMatches)
	for i := 0; i < numMatches; i++ {
		for j := 0; j < 12; j++ {
			anonSchedule[i][j], err = strconv.Atoi(csvLines[i][j])
			if err != nil {
				return nil, err
			}
		}
	}

	// Generate a random permutation of the team ordering to fill into the pre-randomized schedule.
	teamShuffle := schedulePerm(numTeams)
	matches := make([]model.Match, numMatches)
	for i, anonMatch := range anonSchedule {
		matches[i].Type = matchType
		matches[i].TypeOrder = i + 1
		if matchType == model.Practice {
			matches[i].ShortName = fmt.Sprintf("P%d", i+1)
			matches[i].LongName = fmt.Sprintf("Practice %d", i+1)
			matches[i].TbaMatchKey.CompLevel = "p"
		} else if matchType == model.Qualification {
			matches[i].ShortName = fmt.Sprintf("Q%d", i+1)
			matches[i].LongName = fmt.Sprintf("Qualification %d", i+1)
			matches[i].TbaMatchKey.CompLevel = "qm"
		} else {
			return nil, fmt.Errorf("invalid match type %q", matchType)
		}
		// Use only the first two team slots per alliance from the existing 6-team templates.
		matches[i].Red1 = teams[teamShuffle[anonMatch[0]-1]].Id
		matches[i].Red1IsSurrogate = anonMatch[1] == 1
		matches[i].Red2 = teams[teamShuffle[anonMatch[2]-1]].Id
		matches[i].Red2IsSurrogate = anonMatch[3] == 1
		matches[i].Red3 = 0
		matches[i].Red3IsSurrogate = false
		matches[i].Blue1 = teams[teamShuffle[anonMatch[6]-1]].Id
		matches[i].Blue1IsSurrogate = anonMatch[7] == 1
		matches[i].Blue2 = teams[teamShuffle[anonMatch[8]-1]].Id
		matches[i].Blue2IsSurrogate = anonMatch[9] == 1
		matches[i].Blue3 = 0
		matches[i].Blue3IsSurrogate = false
		matches[i].TbaMatchKey.MatchNumber = i + 1
	}

	// Fill in the match times.
	matchIndex := 0
	for _, block := range scheduleBlocks {
		for i := 0; i < block.NumMatches && matchIndex < numMatches; i++ {
			matches[matchIndex].Time = block.StartTime.Add(time.Duration(i*block.MatchSpacingSec) * time.Second)
			matchIndex++
		}
	}

	return matches, nil
}

// Returns the total number of matches that can be run within the given schedule blocks.
func countMatches(scheduleBlocks []model.ScheduleBlock) int {
	numMatches := 0
	for _, block := range scheduleBlocks {
		numMatches += block.NumMatches
	}
	return numMatches
}
