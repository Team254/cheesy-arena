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
	MatchType  string
	RedScore   Score
	BlueScore  Score
	RedCards   map[string]string
	BlueCards  map[string]string
}

type MatchResultDb struct {
	Id            int
	MatchId       int
	PlayNumber    int
	MatchType     string
	RedScoreJson  string
	BlueScoreJson string
	RedCardsJson  string
	BlueCardsJson string
}

type Score struct {
	AutoMobility int
	AutoGears    int
	AutoFuelLow  int
	AutoFuelHigh int
	Gears        int
	FuelLow      int
	FuelHigh     int
	Takeoffs     int
	Fouls        []Foul
	ElimDq       bool
}

type Foul struct {
	TeamId         int
	Rule           string
	IsTechnical    bool
	TimeInMatchSec float64
}

type ScoreSummary struct {
	AutoMobilityPoints  int
	AutoPoints          int
	RotorPoints         int
	TakeoffPoints       int
	PressurePoints      int
	BonusPoints         int
	FoulPoints          int
	Score               int
	PressureGoalReached bool
	RotorGoalReached    bool
}

// Returns a new match result object with empty slices instead of nil.
func NewMatchResult() *MatchResult {
	matchResult := new(MatchResult)
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
	return scoreSummary(&matchResult.RedScore, matchResult.BlueScore.Fouls, matchResult.MatchType)
}

// Calculates and returns the summary fields used for ranking and display for the blue alliance.
func (matchResult *MatchResult) BlueScoreSummary() *ScoreSummary {
	return scoreSummary(&matchResult.BlueScore, matchResult.RedScore.Fouls, matchResult.MatchType)
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

	// No elimination tiebreakers in 2017 Chezy Champs rules.
}

// Calculates and returns the summary fields used for ranking and display.
func scoreSummary(score *Score, opponentFouls []Foul, matchType string) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous score.
	summary.AutoMobilityPoints = 5 * score.AutoMobility
	autoRotors := numRotors(score.AutoGears)
	summary.AutoPoints = summary.AutoMobilityPoints + 60*autoRotors + score.AutoFuelHigh +
		score.AutoFuelLow/3

	// Calculate teleop score.
	teleopRotors := numRotors(score.AutoGears+score.Gears) - autoRotors
	summary.RotorPoints = 60*autoRotors + 40*teleopRotors
	summary.TakeoffPoints = 50 * score.Takeoffs
	summary.PressurePoints = (9*score.AutoFuelHigh + 3*score.AutoFuelLow + 3*score.FuelHigh + score.FuelLow) / 9

	// Calculate bonuses.
	if summary.PressurePoints >= 40 {
		summary.PressureGoalReached = true
		if matchType == "elimination" {
			summary.BonusPoints += 20
		}
	}
	if autoRotors+teleopRotors == 4 {
		summary.RotorGoalReached = true
		if matchType == "elimination" {
			summary.BonusPoints += 100
		}
	}

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		if foul.IsTechnical {
			summary.FoulPoints += 25
		} else {
			summary.FoulPoints += 5
		}
	}

	summary.Score = summary.AutoMobilityPoints + summary.RotorPoints + summary.TakeoffPoints + summary.PressurePoints +
		summary.BonusPoints + summary.FoulPoints

	return summary
}

// Converts the nested struct MatchResult to the DB version that has JSON fields.
func (matchResult *MatchResult) serialize() (*MatchResultDb, error) {
	matchResultDb := MatchResultDb{Id: matchResult.Id, MatchId: matchResult.MatchId,
		PlayNumber: matchResult.PlayNumber, MatchType: matchResult.MatchType}
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
	matchResult := MatchResult{Id: matchResultDb.Id, MatchId: matchResultDb.MatchId,
		PlayNumber: matchResultDb.PlayNumber, MatchType: matchResultDb.MatchType}
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

// Returns the number of completed rotors given the number of gears installed.
func numRotors(numGears int) int {
	rotorGears := []int{1, 3, 7, 13}
	for rotors, gears := range rotorGears {
		if numGears < gears {
			return rotors
		}
	}
	return len(rotorGears)
}
