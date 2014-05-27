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
	RedFouls   Fouls
	BlueFouls  Fouls
	Cards      Cards
}

type MatchResultDb struct {
	Id            int
	MatchId       int
	PlayNumber    int
	RedScoreJson  string
	BlueScoreJson string
	RedFoulsJson  string
	BlueFoulsJson string
	CardsJson     string
}

type Score struct {
	AutoMobilityBonuses int
	AutoHighHot         int
	AutoHigh            int
	AutoLowHot          int
	AutoLow             int
	AutoClearHigh       int
	AutoClearLow        int
	Cycles              []Cycle
}

type Cycle struct {
	Assists    int
	Truss      bool
	Catch      bool
	ScoredHigh bool
	ScoredLow  bool
	DeadBall   bool
}

type Fouls struct {
	Fouls     []Foul
	TechFouls []Foul
}

type Foul struct {
	TeamId         int
	Rule           string
	TimeInMatchSec float32
}

type Cards struct {
	YellowCardTeamIds []int
	RedCardTeamIds    []int
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

// Converts the nested struct MatchResult to the DB version that has JSON fields.
func (matchResult *MatchResult) serialize() (*MatchResultDb, error) {
	matchResultDb := MatchResultDb{Id: matchResult.Id, MatchId: matchResult.MatchId, PlayNumber: matchResult.PlayNumber}
	if err := serializeHelper(&matchResultDb.RedScoreJson, matchResult.RedScore); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.BlueScoreJson, matchResult.BlueScore); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.RedFoulsJson, matchResult.RedFouls); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.BlueFoulsJson, matchResult.BlueFouls); err != nil {
		return nil, err
	}
	if err := serializeHelper(&matchResultDb.CardsJson, matchResult.Cards); err != nil {
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
	if err := json.Unmarshal([]byte(matchResultDb.RedFoulsJson), &matchResult.RedFouls); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(matchResultDb.BlueFoulsJson), &matchResult.BlueFouls); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(matchResultDb.CardsJson), &matchResult.Cards); err != nil {
		return nil, err
	}
	return &matchResult, nil
}
