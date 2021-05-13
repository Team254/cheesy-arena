// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a match at an event.

package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Match struct {
	Id               int `db:"id"`
	Type             string
	DisplayName      string
	Time             time.Time
	ElimRound        int
	ElimGroup        int
	ElimInstance     int
	ElimRedAlliance  int
	ElimBlueAlliance int
	Red1             int
	Red1IsSurrogate  bool
	Red2             int
	Red2IsSurrogate  bool
	Red3             int
	Red3IsSurrogate  bool
	Blue1            int
	Blue1IsSurrogate bool
	Blue2            int
	Blue2IsSurrogate bool
	Blue3            int
	Blue3IsSurrogate bool
	StartedAt        time.Time
	ScoreCommittedAt time.Time
	Status           MatchStatus
}

type MatchStatus string

const (
	RedWonMatch    MatchStatus = "R"
	BlueWonMatch   MatchStatus = "B"
	TieMatch       MatchStatus = "T"
	MatchNotPlayed MatchStatus = ""
)

var ElimRoundNames = map[int]string{1: "F", 2: "SF", 4: "QF", 8: "EF"}

func (database *Database) CreateMatch(match *Match) error {
	return database.matchTable.create(match)
}

func (database *Database) GetMatchById(id int) (*Match, error) {
	var match *Match
	err := database.matchTable.getById(id, &match)
	return match, err
}

func (database *Database) UpdateMatch(match *Match) error {
	return database.matchTable.update(match)
}

func (database *Database) DeleteMatch(id int) error {
	return database.matchTable.delete(id)
}

func (database *Database) TruncateMatches() error {
	return database.matchTable.truncate()
}

func (database *Database) GetMatchByName(matchType string, displayName string) (*Match, error) {
	var matches []Match
	if err := database.matchTable.getAll(&matches); err != nil {
		return nil, err
	}

	for _, match := range matches {
		if match.Type == matchType && match.DisplayName == displayName {
			return &match, nil
		}
	}
	return nil, nil
}

func (database *Database) GetMatchesByElimRoundGroup(round int, group int) ([]Match, error) {
	matches, err := database.GetMatchesByType("elimination")
	if err != nil {
		return nil, err
	}

	var matchingMatches []Match
	for _, match := range matches {
		if match.ElimRound == round && match.ElimGroup == group {
			matchingMatches = append(matchingMatches, match)
		}
	}
	return matchingMatches, nil
}

func (database *Database) GetMatchesByType(matchType string) ([]Match, error) {
	var matches []Match
	if err := database.matchTable.getAll(&matches); err != nil {
		return nil, err
	}

	var matchingMatches []Match
	for _, match := range matches {
		if match.Type == matchType {
			matchingMatches = append(matchingMatches, match)
		}
	}

	sort.Slice(matchingMatches, func(i, j int) bool {
		if matchingMatches[i].ElimRound == matchingMatches[j].ElimRound {
			if matchingMatches[i].ElimInstance == matchingMatches[j].ElimInstance {
				if matchingMatches[i].ElimGroup == matchingMatches[j].ElimGroup {
					return matchingMatches[i].Id < matchingMatches[j].Id
				}
				return matchingMatches[i].ElimGroup < matchingMatches[j].ElimGroup
			}
			return matchingMatches[i].ElimInstance < matchingMatches[j].ElimInstance
		}
		return matchingMatches[i].ElimRound > matchingMatches[j].ElimRound
	})
	return matchingMatches, nil
}

func (match *Match) IsComplete() bool {
	return match.Status != MatchNotPlayed
}

func (match *Match) CapitalizedType() string {
	if match.Type == "" {
		return ""
	} else if match.Type == "elimination" {
		return "Playoff"
	}
	return strings.ToUpper(match.Type[0:1]) + match.Type[1:]
}

func (match *Match) TypePrefix() string {
	if match.Type == "practice" {
		return "P"
	} else if match.Type == "qualification" {
		return "Q"
	}
	return ""
}

func (match *Match) TbaCode() string {
	if match.Type == "qualification" {
		return fmt.Sprintf("qm%s", match.DisplayName)
	} else if match.Type == "elimination" {
		return fmt.Sprintf("%s%dm%d", strings.ToLower(ElimRoundNames[match.ElimRound]), match.ElimGroup,
			match.ElimInstance)
	}
	return ""
}

// Returns true if the match is of a type that allows substitution of teams.
func (match *Match) ShouldAllowSubstitution() bool {
	return match.Type != "qualification"
}

// Returns true if the red and yellow cards should be updated as a result of the match.
func (match *Match) ShouldUpdateCards() bool {
	return match.Type == "qualification" || match.Type == "elimination"
}

// Returns true if the rankings should be updated as a result of the match.
func (match *Match) ShouldUpdateRankings() bool {
	return match.Type == "qualification"
}

// Returns true if the elimination match set should be updated as a result of the match.
func (match *Match) ShouldUpdateEliminationMatches() bool {
	return match.Type == "elimination"
}
