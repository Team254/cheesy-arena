// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a match at an event.

package main

import (
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
	Winner           string
}

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
	err := database.teamMap.Select(&matches, "SELECT * FROM matches WHERE type = ? AND displayname = ?",
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
	err := database.teamMap.Select(&matches, "SELECT * FROM matches WHERE type = 'elimination' AND "+
		"elimround = ? AND elimgroup = ? ORDER BY eliminstance", round, group)
	return matches, err
}

func (database *Database) GetMatchesByType(matchType string) ([]Match, error) {
	var matches []Match
	err := database.teamMap.Select(&matches,
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
