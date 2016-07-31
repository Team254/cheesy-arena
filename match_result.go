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
	AutoDefensesReached int
	AutoDefensesCrossed [5]int
	AutoLowGoals        int
	AutoHighGoals       int
	DefensesCrossed     [5]int
	LowGoals            int
	HighGoals           int
	Challenges          int
	Scales              int
	Fouls               []Foul
	ElimDq              bool
}

type Foul struct {
	TeamId         int
	Rule           string
	IsTechnical    bool
	TimeInMatchSec float64
}

type ScoreSummary struct {
	AutoPoints            int
	DefensePoints         int
	GoalPoints            int
	ScaleChallengePoints  int
	TeleopPoints          int
	FoulPoints            int
	BonusPoints           int
	Score                 int
	Breached              bool
	Captured              bool
	OpponentTowerStrength int
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

	// No elimination tiebreakers in 2016 Chezy Champs rules.
}

// Calculates and returns the summary fields used for ranking and display.
func scoreSummary(score *Score, opponentFouls []Foul, matchType string) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous score.
	autoDefensePoints := 0
	for _, defense := range score.AutoDefensesCrossed {
		autoDefensePoints += 10 * defense
	}
	autoGoalPoints := 5*score.AutoLowGoals + 10*score.AutoHighGoals
	summary.AutoPoints = 2*score.AutoDefensesReached + autoDefensePoints + autoGoalPoints

	// Calculate teleop score.
	teleopDefensePoints := 0
	for _, defense := range score.DefensesCrossed {
		teleopDefensePoints += 5 * defense
	}
	summary.ScaleChallengePoints = 5*score.Challenges + 15*score.Scales
	teleopGoalPoints := 2*score.LowGoals + 5*score.HighGoals
	summary.TeleopPoints = teleopDefensePoints + teleopGoalPoints + summary.ScaleChallengePoints

	// Calculate tower strength.
	numTechFouls := 0
	for _, foul := range score.Fouls {
		if foul.IsTechnical {
			numTechFouls++
		}
	}
	summary.OpponentTowerStrength = eventSettings.InitialTowerStrength + numTechFouls - score.AutoLowGoals -
		score.AutoHighGoals - score.LowGoals - score.HighGoals

	// Calculate bonuses.
	summary.BonusPoints = 0
	numDefensesDamaged := 0
	for i := 0; i < 5; i++ {
		if score.AutoDefensesCrossed[i]+score.DefensesCrossed[i] == 2 {
			numDefensesDamaged++
		}
	}
	if numDefensesDamaged >= 4 {
		summary.Breached = true
		if matchType == "elimination" {
			summary.BonusPoints += 20
		}
	}
	if score.Challenges+score.Scales == 3 && summary.OpponentTowerStrength <= 0 {
		summary.Captured = true
		if matchType == "elimination" {
			summary.BonusPoints += 25
		}
	}

	summary.FoulPoints = 5 * len(opponentFouls)
	summary.DefensePoints = autoDefensePoints + teleopDefensePoints
	summary.GoalPoints = autoGoalPoints + teleopGoalPoints
	summary.Score = summary.AutoPoints + summary.TeleopPoints + summary.FoulPoints + summary.BonusPoints

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
