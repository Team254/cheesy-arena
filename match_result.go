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
	RedFouls   []Foul
	BlueFouls  []Foul
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

type Foul struct {
	TeamId         int
	Rule           string
	TimeInMatchSec float32
	IsTechnical    bool
}

type Cards struct {
	YellowCardTeamIds []int
	RedCardTeamIds    []int
}

type ScoreSummary struct {
	AutoPoints       int
	AssistPoints     int
	TrussCatchPoints int
	GoalPoints       int
	TeleopPoints     int
	FoulPoints       int
	Score            int
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
	return scoreSummary(&matchResult.RedScore, matchResult.BlueFouls)
}

// Calculates and returns the summary fields used for ranking and display for the blue alliance.
func (matchResult *MatchResult) BlueScoreSummary() *ScoreSummary {
	return scoreSummary(&matchResult.BlueScore, matchResult.RedFouls)
}

// Calculates and returns the summary fields used for ranking and display.
func scoreSummary(score *Score, opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Calculate autonomous score.
	summary.AutoPoints = 5*score.AutoMobilityBonuses + 20*score.AutoHighHot + 15*score.AutoHigh +
		11*score.AutoLowHot + 6*score.AutoLow

	// Calculate teleop score.
	summary.GoalPoints = 10*score.AutoClearHigh + 1*score.AutoClearLow
	for _, cycle := range score.Cycles {
		if cycle.Truss {
			summary.TrussCatchPoints += 10
			if cycle.Catch {
				summary.TrussCatchPoints += 10
			}
		}
		if cycle.ScoredHigh {
			summary.GoalPoints += 10
		} else if cycle.ScoredLow {
			summary.GoalPoints += 1
		}
		if cycle.ScoredHigh || cycle.ScoredLow {
			if cycle.Assists == 2 {
				summary.AssistPoints += 10
			} else if cycle.Assists == 3 {
				summary.AssistPoints += 30
			}
		}
	}

	// Calculate foul score.
	summary.FoulPoints = 0
	for _, foul := range opponentFouls {
		if foul.IsTechnical {
			summary.FoulPoints += 50
		} else {
			summary.FoulPoints += 20
		}
	}

	// Fill in summed values.
	summary.TeleopPoints = summary.AssistPoints + summary.TrussCatchPoints + summary.GoalPoints
	summary.Score = summary.AutoPoints + summary.TeleopPoints + summary.FoulPoints

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
