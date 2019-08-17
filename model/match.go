// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a match at an event.

package model

import (
	"fmt"
	"strings"
	"time"
)

type Match struct {
	Id               int
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
	Status           string
	StartedAt        time.Time
	ScoreCommittedAt time.Time
	Winner           string
}

var ElimRoundNames = map[int]string{1: "F", 2: "SF", 4: "QF", 8: "EF"}

func (database *Database) CreateMatch(match *Match) error {
	return database.matchMap.Insert(match)
}

func (database *Database) GetMatchById(id int) (*Match, error) {
	match := new(Match)
	err := database.matchMap.Get(match, id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		match = nil
		err = nil
	}
	return match, err
}

func (database *Database) SaveMatch(match *Match) error {
	_, err := database.matchMap.Update(match)
	return err
}

func (database *Database) DeleteMatch(match *Match) error {
	_, err := database.matchMap.Delete(match)
	return err
}

func (database *Database) TruncateMatches() error {
	return database.matchMap.TruncateTables()
}

func (database *Database) GetMatchByName(matchType string, displayName string) (*Match, error) {
	var matches []Match
	err := database.matchMap.Select(&matches, "SELECT * FROM matches WHERE type = ? AND displayname = ?",
		matchType, displayName)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return &matches[0], err
}

func (database *Database) GetMatchesByElimRoundGroup(round int, group int) ([]Match, error) {
	var matches []Match
	err := database.matchMap.Select(&matches, "SELECT * FROM matches WHERE type = 'elimination' AND "+
		"elimround = ? AND elimgroup = ? ORDER BY eliminstance", round, group)
	return matches, err
}

func (database *Database) GetMatchesByType(matchType string) ([]Match, error) {
	var matches []Match
	err := database.matchMap.Select(&matches,
		"SELECT * FROM matches WHERE type = ? ORDER BY elimround desc, eliminstance, elimgroup, id", matchType)
	return matches, err
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
