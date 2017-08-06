// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package main

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/game"
	"sort"
	"strconv"
)

type RankingDb struct {
	TeamId            int
	Rank              int
	RankingFieldsJson string
}

func (database *Database) CreateRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	return database.rankingMap.Insert(rankingDb)
}

func (database *Database) GetRankingForTeam(teamId int) (*game.Ranking, error) {
	rankingDb := new(RankingDb)
	err := database.rankingMap.Get(rankingDb, teamId)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return nil, nil
	}
	ranking, err := rankingDb.deserialize()
	if err != nil {
		return nil, err
	}
	return ranking, err
}

func (database *Database) SaveRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	_, err = database.rankingMap.Update(rankingDb)
	return err
}

func (database *Database) DeleteRanking(ranking *game.Ranking) error {
	rankingDb, err := serializeRanking(ranking)
	if err != nil {
		return err
	}
	_, err = database.rankingMap.Delete(rankingDb)
	return err
}

func (database *Database) TruncateRankings() error {
	return database.rankingMap.TruncateTables()
}

func (database *Database) GetAllRankings() ([]game.Ranking, error) {
	var rankingDbs []RankingDb
	err := database.rankingMap.Select(&rankingDbs, "SELECT * FROM rankings ORDER BY rank")
	if err != nil {
		return nil, err
	}
	var rankings []game.Ranking
	for _, rankingDb := range rankingDbs {
		ranking, err := rankingDb.deserialize()
		if err != nil {
			return nil, err
		}
		rankings = append(rankings, *ranking)
	}
	return rankings, err
}

// Determines the rankings from the stored match results, and saves them to the database.
func (database *Database) CalculateRankings() error {
	matches, err := database.GetMatchesByType("qualification")
	if err != nil {
		return err
	}
	rankings := make(map[int]*game.Ranking)
	for _, match := range matches {
		if match.Status != "complete" {
			continue
		}
		matchResult, err := database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		if !match.Red1IsSurrogate {
			addMatchResultToRankings(rankings, match.Red1, matchResult, true)
		}
		if !match.Red2IsSurrogate {
			addMatchResultToRankings(rankings, match.Red2, matchResult, true)
		}
		if !match.Red3IsSurrogate {
			addMatchResultToRankings(rankings, match.Red3, matchResult, true)
		}
		if !match.Blue1IsSurrogate {
			addMatchResultToRankings(rankings, match.Blue1, matchResult, false)
		}
		if !match.Blue2IsSurrogate {
			addMatchResultToRankings(rankings, match.Blue2, matchResult, false)
		}
		if !match.Blue3IsSurrogate {
			addMatchResultToRankings(rankings, match.Blue3, matchResult, false)
		}
	}

	sortedRankings := sortRankings(rankings)

	// Stuff the rankings into the database in an atomic operation to prevent messing them up halfway.
	transaction, err := database.rankingMap.Begin()
	if err != nil {
		return err
	}
	_, err = transaction.Exec("DELETE FROM rankings")
	if err != nil {
		return err
	}
	for rank, ranking := range sortedRankings {
		ranking.Rank = rank + 1
		rankingDb, err := serializeRanking(ranking)
		if err != nil {
			return err
		}
		err = transaction.Insert(rankingDb)
		if err != nil {
			return err
		}
	}
	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Checks all the match results for yellow and red cards, and updates the team model accordingly.
func (database *Database) CalculateTeamCards(matchType string) error {
	teams, err := database.GetAllTeams()
	if err != nil {
		return err
	}
	teamsMap := make(map[string]Team)
	for _, team := range teams {
		team.YellowCard = false
		teamsMap[strconv.Itoa(team.Id)] = team
	}

	matches, err := database.GetMatchesByType(matchType)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if match.Status != "complete" {
			continue
		}
		matchResult, err := database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}

		// Mark the team as having a yellow card if they got either a yellow or red in a previous match.
		for teamId, card := range matchResult.RedCards {
			if team, ok := teamsMap[teamId]; ok && card != "" {
				team.YellowCard = true
				teamsMap[teamId] = team
			}
		}
		for teamId, card := range matchResult.BlueCards {
			if team, ok := teamsMap[teamId]; ok && card != "" {
				team.YellowCard = true
				teamsMap[teamId] = team
			}
		}
	}

	// Save the teams to the database.
	for _, team := range teamsMap {
		err = db.SaveTeam(&team)
		if err != nil {
			return err
		}
	}

	return nil
}

// Incrementally accounts for the given match result in the set of rankings that are being built.
func addMatchResultToRankings(rankings map[int]*game.Ranking, teamId int, matchResult *MatchResult, isRed bool) {
	ranking := rankings[teamId]
	if ranking == nil {
		ranking = &game.Ranking{TeamId: teamId}
		rankings[teamId] = ranking
	}

	// Determine whether the team was disqualified.
	var cards map[string]string
	if isRed {
		cards = matchResult.RedCards
	} else {
		cards = matchResult.BlueCards
	}
	disqualified := false
	if card, ok := cards[strconv.Itoa(teamId)]; ok && card == "red" {
		disqualified = true
	}

	if isRed {
		ranking.AddScoreSummary(matchResult.RedScoreSummary(), matchResult.BlueScoreSummary(), disqualified)
	} else {
		ranking.AddScoreSummary(matchResult.BlueScoreSummary(), matchResult.RedScoreSummary(), disqualified)
	}
}

func sortRankings(rankings map[int]*game.Ranking) game.Rankings {
	var sortedRankings game.Rankings
	for _, ranking := range rankings {
		sortedRankings = append(sortedRankings, ranking)
	}
	sort.Sort(sortedRankings)
	return sortedRankings
}

// Converts the nested struct MatchResult to the DB version that has JSON fields.
func serializeRanking(ranking *game.Ranking) (*RankingDb, error) {
	rankingDb := RankingDb{TeamId: ranking.TeamId, Rank: ranking.Rank}
	if err := serializeHelper(&rankingDb.RankingFieldsJson, ranking.RankingFields); err != nil {
		return nil, err
	}
	return &rankingDb, nil
}

// Converts the DB Ranking with JSON fields to the nested struct version.
func (rankingDb *RankingDb) deserialize() (*game.Ranking, error) {
	ranking := game.Ranking{TeamId: rankingDb.TeamId, Rank: rankingDb.Rank}
	if err := json.Unmarshal([]byte(rankingDb.RankingFieldsJson), &ranking.RankingFields); err != nil {
		return nil, err
	}
	return &ranking, nil
}
