// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for the results (score and fouls) from a match at an event.

package main

import (
	"encoding/json"
)

type MatchResult struct {
	Id         int
	MatchId    int
	PlayNumber int
	RedScore   Score
	BlueScore  Score
	RedCards   map[string]string
	BlueCards  map[string]string
}

type MatchResultDb struct {
	Id            int
	MatchId       int
	PlayNumber    int
	RedScoreJson  string
	BlueScoreJson string
	RedCardsJson  string
	BlueCardsJson string
}

type Score struct {
	AutoRobotSet       bool
	AutoToteSet        bool
	AutoContainerSet   bool
	AutoStackedToteSet bool
	Totes              int
	ContainerLevels    []int
	ContainerLitter    int
	LandfillLitter     int
	UnprocessedLitter  int
	CoopertitionSet    bool
	CoopertitionStack  bool
	Fouls              []Foul
	ElimDq             bool
}

type Foul struct {
	TeamId         int
	Rule           string
	TimeInMatchSec float64
}

type ScoreSummary struct {
	CoopertitionPoints int
	AutoPoints         int
	ContainerPoints    int
	TotePoints         int
	LitterPoints       int
	FoulPoints         int
	Score              int
}

// Returns a new match result object with empty slices instead of nil.
func NewMatchResult() *MatchResult {
	matchResult := new(MatchResult)
	matchResult.RedScore.ContainerLevels = []int{}
	matchResult.BlueScore.ContainerLevels = []int{}
	matchResult.RedScore.Fouls = []Foul{}
	matchResult.BlueScore.Fouls = []Foul{}
	matchResult.RedCards = make(map[string]string)
	matchResult.BlueCards = make(map[string]string)
	return matchResult
}

func (database *Database) CreateMatchResult(matchResult *MatchResult) error {
	matchResultDb, err := matchResult.serialize()
	if err != nil {
		return err
	}
	err = database.matchResultMap.Insert(matchResultDb)
	if err != nil {
		return err
	}
	matchResult.Id = matchResultDb.Id
	return nil
}

func (database *Database) GetMatchResultForMatch(matchId int) (*MatchResult, error) {
	var matchResults []MatchResultDb
	query := "SELECT * FROM match_results WHERE matchid = ? ORDER BY playnumber DESC LIMIT 1"
	err := database.matchResultMap.Select(&matchResults, query, matchId)
	if err != nil {
		return nil, err
	}
	if len(matchResults) == 0 {
		return nil, nil
	}
	matchResult, err := matchResults[0].deserialize()
	if err != nil {
		return nil, err
	}
	return matchResult, err
}

func (database *Database) SaveMatchResult(matchResult *MatchResult) error {
	matchResultDb, err := matchResult.serialize()
	if err != nil {
		return err
	}
	_, err = database.matchResultMap.Update(matchResultDb)
	return err
}

func (database *Database) DeleteMatchResult(matchResult *MatchResult) error {
	matchResultDb, err := matchResult.serialize()
	if err != nil {
		return err
	}
	_, err = database.matchResultMap.Delete(matchResultDb)
	return err
}

func (database *Database) TruncateMatchResults() error {
	return database.matchResultMap.TruncateTables()
}

// Calculates and returns the summary fields used for ranking and display for the red alliance.
func (matchResult *MatchResult) RedScoreSummary() *ScoreSummary {
	return scoreSummary(&matchResult.RedScore)
}

// Calculates and returns the summary fields used for ranking and display for the blue alliance.
func (matchResult *MatchResult) BlueScoreSummary() *ScoreSummary {
	return scoreSummary(&matchResult.BlueScore)
}

// Checks the score for disqualifications or a tie and adjusts it appropriately.
func (matchResult *MatchResult) CorrectEliminationScore() {
	matchResult.RedScore.ElimDq = false
	for _, card := range matchResult.RedCards {
		if card == "red" {
			matchResult.RedScore.ElimDq = true
		}
	}
	for _, card := range matchResult.BlueCards {
		if card == "red" {
			matchResult.BlueScore.ElimDq = true
		}
	}

	// No elimination tiebreakers in 2015.
}

// Calculates and returns the summary fields used for ranking and display.
func scoreSummary(score *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous score.
	summary.AutoPoints = 0
	if score.AutoRobotSet {
		summary.AutoPoints += 4
	}
	if score.AutoContainerSet {
		summary.AutoPoints += 8
	}
	if score.AutoStackedToteSet {
		summary.AutoPoints += 20
	} else if score.AutoToteSet {
		summary.AutoPoints += 6
	}

	// Calculate teleop score.
	summary.ContainerPoints = 0
	for _, containerLevel := range score.ContainerLevels {
		summary.ContainerPoints += 4 * containerLevel
	}
	summary.TotePoints = 2 * score.Totes
	summary.LitterPoints = 6*score.ContainerLitter + score.LandfillLitter + 4*score.UnprocessedLitter
	if score.CoopertitionStack {
		summary.CoopertitionPoints = 40
	} else if score.CoopertitionSet {
		summary.CoopertitionPoints = 20
	}
	summary.FoulPoints = 6 * len(score.Fouls)

	summary.Score = summary.CoopertitionPoints + summary.AutoPoints + summary.ContainerPoints +
		summary.TotePoints + summary.LitterPoints - summary.FoulPoints

	return summary
}

// Converts the nested struct MatchResult to the DB version that has JSON fields.
func (matchResult *MatchResult) serialize() (*MatchResultDb, error) {
	matchResultDb := MatchResultDb{Id: matchResult.Id, MatchId: matchResult.MatchId, PlayNumber: matchResult.PlayNumber}
	if err := serializeHelper(&matchResultDb.RedScoreJson, matchResult.RedScore); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.BlueScoreJson, matchResult.BlueScore); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.RedCardsJson, matchResult.RedCards); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.BlueCardsJson, matchResult.BlueCards); err != nil {
		return nil, err
	}
	return &matchResultDb, nil
}

func serializeHelper(target *string, source interface{}) error {
	bytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	*target = string(bytes)
	return nil
}

// Converts the DB MatchResult with JSON fields to the nested struct version.
func (matchResultDb *MatchResultDb) deserialize() (*MatchResult, error) {
	matchResult := MatchResult{Id: matchResultDb.Id, MatchId: matchResultDb.MatchId, PlayNumber: matchResultDb.PlayNumber}
	if err := json.Unmarshal([]byte(matchResultDb.RedScoreJson), &matchResult.RedScore); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(matchResultDb.BlueScoreJson), &matchResult.BlueScore); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(matchResultDb.RedCardsJson), &matchResult.RedCards); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(matchResultDb.BlueCardsJson), &matchResult.BlueCards); err != nil {
		return nil, err
	}
	return &matchResult, nil
}
