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
	PlayoffMatchGroupId string
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
	UseTiebreakCriteria bool
	TbaMatchKey         TbaMatchKey
}

type TbaMatchKey struct {
	CompLevel   string
	SetNumber   int
	MatchNumber int
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

func (database *Database) GetMatchByTypeOrder(matchType MatchType, typeOrder int) (*Match, error) {
	matches, err := database.GetMatchesByType(matchType, true)
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

func (database *Database) GetMatchesByType(matchType MatchType, includeHidden bool) ([]Match, error) {
	matches, err := database.matchTable.getAll()
	if err != nil {
		return nil, err
	}

	var matchingMatches []Match
	for _, match := range matches {
		if match.Type == matchType && (includeHidden || match.Status != game.MatchHidden) {
			matchingMatches = append(matchingMatches, match)
		}
	}

	sort.Slice(matchingMatches, func(i, j int) bool {
		return matchingMatches[i].TypeOrder < matchingMatches[j].TypeOrder
	})
	return matchingMatches, nil
}

func (match *Match) IsComplete() bool {
	return match.Status == game.RedWonMatch || match.Status == game.BlueWonMatch || match.Status == game.TieMatch
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
