// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for creating practice and qualification match schedules.

package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const schedulesDir = "schedules"
const teamsPerMatch = 6

type ScheduleBlock struct {
	StartTime       time.Time
	NumMatches      int
	MatchSpacingSec int
}

// Creates a random schedule for the given parameters and returns it as a list of matches.
func BuildRandomSchedule(teams []Team, scheduleBlocks []ScheduleBlock, matchType string) ([]Match, error) {
	// Load the anonymized, pre-randomized match schedule for the given number of teams and matches per team.
	numTeams := len(teams)
	numMatches := countMatches(scheduleBlocks)
	matchesPerTeam := int(float32(numMatches*teamsPerMatch) / float32(numTeams))

	// Adjust the number of matches to remove any excess from non-perfect block scheduling.
	numMatches = int(math.Ceil(float64(numTeams) * float64(matchesPerTeam) / teamsPerMatch))

	file, err := os.Open(fmt.Sprintf("%s/%d_%d.csv", schedulesDir, numTeams, matchesPerTeam))
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
	teamShuffle := rand.Perm(numTeams)
	matches := make([]Match, numMatches)
	for i, anonMatch := range anonSchedule {
		matches[i].Type = matchType
		matches[i].DisplayName = strconv.Itoa(i + 1)
		matches[i].Red1 = teams[teamShuffle[anonMatch[0]-1]].Id
		matches[i].Red1IsSurrogate = (anonMatch[1] == 1)
		matches[i].Red2 = teams[teamShuffle[anonMatch[2]-1]].Id
		matches[i].Red2IsSurrogate = (anonMatch[3] == 1)
		matches[i].Red3 = teams[teamShuffle[anonMatch[4]-1]].Id
		matches[i].Red3IsSurrogate = (anonMatch[5] == 1)
		matches[i].Blue1 = teams[teamShuffle[anonMatch[6]-1]].Id
		matches[i].Blue1IsSurrogate = (anonMatch[7] == 1)
		matches[i].Blue2 = teams[teamShuffle[anonMatch[8]-1]].Id
		matches[i].Blue2IsSurrogate = (anonMatch[9] == 1)
		matches[i].Blue3 = teams[teamShuffle[anonMatch[10]-1]].Id
		matches[i].Blue3IsSurrogate = (anonMatch[11] == 1)
	}

	// Fill in the match times.
	matchIndex := 0
	for _, block := range scheduleBlocks {
		for i := 0; i < block.NumMatches && matchIndex < numMatches; i++ {
			matches[matchIndex].Time = block.StartTime.Add(time.Duration(i*block.MatchSpacingSec) * time.Second)
			matchIndex++
		}
	}

	randomizeDefenses(matches, numTeams)

	return matches, nil
}

// Returns the total number of matches that can be run within the given schedule blocks.
func countMatches(scheduleBlocks []ScheduleBlock) int {
	numMatches := 0
	for _, block := range scheduleBlocks {
		numMatches += block.NumMatches
	}
	return numMatches
}

// Fills in a random set of defenses per round of all teams playing.
func randomizeDefenses(matches []Match, numTeams int) {
	// Take the floor, to err on the side of a team missing a set of defenses instead of seeing it twice.
	matchesPerRound := numTeams / 6

	var defenseShuffle []int
	for i := 0; i < len(matches); i++ {
		if i%matchesPerRound == 0 {
			// Pick a new set of defenses.
			defenseShuffle = rand.Perm(len(placeableDefenses))
		}

		matches[i].RedDefense1 = "LB"
		matches[i].RedDefense2 = placeableDefenses[defenseShuffle[0]]
		matches[i].RedDefense3 = placeableDefenses[defenseShuffle[1]]
		matches[i].RedDefense4 = placeableDefenses[defenseShuffle[2]]
		matches[i].RedDefense5 = placeableDefenses[defenseShuffle[3]]
		matches[i].BlueDefense1 = "LB"
		matches[i].BlueDefense2 = placeableDefenses[defenseShuffle[0]]
		matches[i].BlueDefense3 = placeableDefenses[defenseShuffle[1]]
		matches[i].BlueDefense4 = placeableDefenses[defenseShuffle[2]]
		matches[i].BlueDefense5 = placeableDefenses[defenseShuffle[3]]
	}
}
