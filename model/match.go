// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a match at an event.

package model

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"sort"
	"strings"
	"time"
)

//go:generate stringer -type=MatchType
type MatchType int

const (
	Test MatchType = iota
	Practice
	Qualification
	Playoff
)

func (t MatchType) Get() MatchType {
	return t
}

type Match struct {
	Id                  int `db:"id"`
	Type                MatchType
	TypeOrder           int
	Time                time.Time
	LongName            string
	ShortName           string
	NameDetail          string
	PlayoffRound        int
	PlayoffGroup        int
	PlayoffInstance     int
	PlayoffRedAlliance  int
	PlayoffBlueAlliance int
	Red1                int
	Red1IsSurrogate     bool
	Red2                int
	Red2IsSurrogate     bool
	Red3                int
	Red3IsSurrogate     bool
	Blue1               int
	Blue1IsSurrogate    bool
	Blue2               int
	Blue2IsSurrogate    bool
	Blue3               int
	Blue3IsSurrogate    bool
	StartedAt           time.Time
	ScoreCommittedAt    time.Time
	FieldReadyAt        time.Time
	Status              game.MatchStatus
}

func (database *Database) CreateMatch(match *Match) error {
	return database.matchTable.create(match)
}

func (database *Database) GetMatchById(id int) (*Match, error) {
	return database.matchTable.getById(id)
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

// TODO(pat): Deprecate this method
func (database *Database) GetMatchByName(matchType MatchType, shortName string) (*Match, error) {
	matches, err := database.matchTable.getAll()
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if match.Type == matchType && match.ShortName == shortName {
			return &match, nil
		}
	}
	return nil, nil
}

func (database *Database) GetMatchByTypeOrder(matchType MatchType, typeOrder int) (*Match, error) {
	matches, err := database.GetMatchesByType(matchType)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if match.TypeOrder == typeOrder {
			return &match, nil
		}
	}
	return nil, nil
}

func (database *Database) GetMatchesByPlayoffRoundGroup(round int, group int) ([]Match, error) {
	matches, err := database.GetMatchesByType(Playoff)
	if err != nil {
		return nil, err
	}

	var matchingMatches []Match
	for _, match := range matches {
		if match.PlayoffRound == round && match.PlayoffGroup == group {
			matchingMatches = append(matchingMatches, match)
		}
	}
	return matchingMatches, nil
}

func (database *Database) GetMatchesByType(matchType MatchType) ([]Match, error) {
	matches, err := database.matchTable.getAll()
	if err != nil {
		return nil, err
	}

	var matchingMatches []Match
	for _, match := range matches {
		if match.Type == matchType {
			matchingMatches = append(matchingMatches, match)
		}
	}

	sort.Slice(matchingMatches, func(i, j int) bool {
		if matchingMatches[i].PlayoffRound == matchingMatches[j].PlayoffRound {
			if matchingMatches[i].PlayoffInstance == matchingMatches[j].PlayoffInstance {
				if matchingMatches[i].PlayoffGroup == matchingMatches[j].PlayoffGroup {
					return matchingMatches[i].TypeOrder < matchingMatches[j].TypeOrder
				}
				return matchingMatches[i].PlayoffGroup < matchingMatches[j].PlayoffGroup
			}
			return matchingMatches[i].PlayoffInstance < matchingMatches[j].PlayoffInstance
		}
		return matchingMatches[i].PlayoffRound < matchingMatches[j].PlayoffRound
	})
	return matchingMatches, nil
}

func (match *Match) IsComplete() bool {
	return match.Status != game.MatchNotPlayed
}

// Returns true if the match is of a type that allows substitution of teams.
func (match *Match) ShouldAllowSubstitution() bool {
	return match.Type != Qualification
}

// Returns true if the red and yellow cards should be updated as a result of the match.
func (match *Match) ShouldUpdateCards() bool {
	return match.Type == Qualification || match.Type == Playoff
}

// Returns true if the rankings should be updated as a result of the match.
func (match *Match) ShouldUpdateRankings() bool {
	return match.Type == Qualification
}

// Returns true if the playoff match set should be updated as a result of the match.
func (match *Match) ShouldUpdatePlayoffMatches() bool {
	return match.Type == Playoff
}

// Returns the enum equivalent of the given match type string.
func MatchTypeFromString(matchTypeString string) (MatchType, error) {
	switch strings.ToLower(matchTypeString) {
	case "test":
		return Test, nil
	case "practice":
		return Practice, nil
	case "qualification":
		return Qualification, nil
	case "playoff":
		return Playoff, nil
	}
	return 0, fmt.Errorf("invalid match type %q", matchTypeString)
}
