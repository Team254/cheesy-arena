// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for creating practice and qualification match schedules.

package tournament

import (
	"encoding/csv"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	schedulesDir  = "schedules"
	TeamsPerMatch = 6
)

// Creates a random schedule for the given parameters and returns it as a list of matches.
func BuildRandomSchedule(teams []model.Team, scheduleBlocks []model.ScheduleBlock,
	matchType string, useBalancedSchedules bool, filePath string) ([]model.Match, error) {
	if matchType == "practice" { //Only use balanced schedules if it's a qual match
		useBalancedSchedules = false
	}
	
	// Load the anonymized, pre-randomized match schedule for the given number of teams and matches per team.
	numTeams := len(teams)
	numMatches := countMatches(scheduleBlocks)
	matchesPerTeam := int(float32(numMatches*TeamsPerMatch) / float32(numTeams))

	// Adjust the number of matches to remove any excess from non-perfect block scheduling.
	numMatches = int(math.Ceil(float64(numTeams) * float64(matchesPerTeam) / TeamsPerMatch))
	
	var file *os.File
	var err error
	if (useBalancedSchedules) {
		file, err = os.Open(fmt.Sprintf("%s/%d_%d_balanced.csv", filepath.Join(model.BaseDir, schedulesDir), numTeams,
			matchesPerTeam))
	} else {
		file, err = os.Open(fmt.Sprintf("%s/%d_%d.csv", filepath.Join(model.BaseDir, schedulesDir), numTeams,
			matchesPerTeam))
	}
	
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

	matches := make([]model.Match, numMatches)
	teamShuffle := make([]int, numTeams)
	if useBalancedSchedules { //Use a custom list of team strengths to fill into the pre-randomized schedule.
		file, err = os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("Unable to find team strengths file")
		}
		defer file.Close()
		reader = csv.NewReader(file)
		csvLines, err = reader.ReadAll()
		if err != nil {
			return nil, err
		}
		if len(csvLines) != numTeams {
			return nil, fmt.Errorf("Team Strengths File does not have the same number of teams as the team list")
		}
		
		for i:= 0; i < numTeams; i++ {
			teamShuffle[i] = -1
			
			for j := 0; j < numTeams; j++ {
				currentTeam, err := strconv.Atoi(csvLines[j][0])
				if err != nil {
					return nil, err
				}
				if currentTeam == teams[i].Id {
					teamShuffle[i] = j
				}
			}
			if teamShuffle[i] == -1 {
				return nil, fmt.Errorf("Team %d is not in the team strengths file", teams[i].Id)
			}
		}
	} else { // Generate a random permutation of the team ordering to fill into the pre-randomized schedule.
		teamShuffle = rand.Perm(numTeams)
	}
	for i, anonMatch := range anonSchedule {
		matches[i].Type = matchType
		matches[i].DisplayName = strconv.Itoa(i + 1)
		matches[i].Red1 = teams[teamShuffle[anonMatch[0]-1]].Id
		matches[i].Red1IsSurrogate = anonMatch[1] == 1
		matches[i].Red2 = teams[teamShuffle[anonMatch[2]-1]].Id
		matches[i].Red2IsSurrogate = anonMatch[3] == 1
		matches[i].Red3 = teams[teamShuffle[anonMatch[4]-1]].Id
		matches[i].Red3IsSurrogate = anonMatch[5] == 1
		matches[i].Blue1 = teams[teamShuffle[anonMatch[6]-1]].Id
		matches[i].Blue1IsSurrogate = anonMatch[7] == 1
		matches[i].Blue2 = teams[teamShuffle[anonMatch[8]-1]].Id
		matches[i].Blue2IsSurrogate = anonMatch[9] == 1
		matches[i].Blue3 = teams[teamShuffle[anonMatch[10]-1]].Id
		matches[i].Blue3IsSurrogate = anonMatch[11] == 1
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
