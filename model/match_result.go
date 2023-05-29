// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for the results (score and fouls) from a match at an event.

package model

import (
	"github.com/Team254/cheesy-arena/game"
)

type MatchResult struct {
	Id         int `db:"id"`
	MatchId    int
	PlayNumber int
	MatchType  MatchType
	RedScore   *game.Score
	BlueScore  *game.Score
	RedCards   map[string]string
	BlueCards  map[string]string
}

// Returns a new match result object with empty slices instead of nil.
func NewMatchResult() *MatchResult {
	matchResult := new(MatchResult)
	matchResult.RedScore = new(game.Score)
	matchResult.BlueScore = new(game.Score)
	matchResult.RedCards = make(map[string]string)
	matchResult.BlueCards = make(map[string]string)
	return matchResult
}

func (database *Database) CreateMatchResult(matchResult *MatchResult) error {
	return database.matchResultTable.create(matchResult)
}

func (database *Database) GetMatchResultForMatch(matchId int) (*MatchResult, error) {
	matchResults, err := database.matchResultTable.getAll()
	if err != nil {
		return nil, err
	}

	var mostRecentMatchResult *MatchResult
	for i, matchResult := range matchResults {
		if matchResult.MatchId == matchId &&
			(mostRecentMatchResult == nil || matchResult.PlayNumber > mostRecentMatchResult.PlayNumber) {
			mostRecentMatchResult = &matchResults[i]
		}
	}
	return mostRecentMatchResult, nil
}

func (database *Database) UpdateMatchResult(matchResult *MatchResult) error {
	return database.matchResultTable.update(matchResult)
}

func (database *Database) DeleteMatchResult(id int) error {
	return database.matchResultTable.delete(id)
}

func (database *Database) TruncateMatchResults() error {
	return database.matchResultTable.truncate()
}

// Calculates and returns the summary fields used for ranking and display for the red alliance.
func (matchResult *MatchResult) RedScoreSummary() *game.ScoreSummary {
	return matchResult.RedScore.Summarize(matchResult.BlueScore)
}

// Calculates and returns the summary fields used for ranking and display for the blue alliance.
func (matchResult *MatchResult) BlueScoreSummary() *game.ScoreSummary {
	return matchResult.BlueScore.Summarize(matchResult.RedScore)
}

// Checks the score for disqualifications or a tie and adjusts it appropriately.
func (matchResult *MatchResult) CorrectPlayoffScore() {
	matchResult.RedScore.PlayoffDq = false
	for _, card := range matchResult.RedCards {
		if card == "red" {
			matchResult.RedScore.PlayoffDq = true
		}
	}
	for _, card := range matchResult.BlueCards {
		if card == "red" {
			matchResult.BlueScore.PlayoffDq = true
		}
	}
}
