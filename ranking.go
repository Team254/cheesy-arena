// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package main

import (
	"math/rand"
	"sort"
	"strconv"
)

type Ranking struct {
	TeamId               int
	Rank                 int
	QualificationAverage float64
	CoopertitionPoints   int
	AutoPoints           int
	ContainerPoints      int
	TotePoints           int
	LitterPoints         int
	Random               float64
	Disqualifications    int
	Played               int
}

type Rankings []*Ranking

func (database *Database) CreateRanking(ranking *Ranking) error {
	return database.rankingMap.Insert(ranking)
}

func (database *Database) GetRankingForTeam(teamId int) (*Ranking, error) {
	ranking := new(Ranking)
	err := database.rankingMap.Get(ranking, teamId)
	if err != nil && err.Error() == "sql: no rows in result set" {
		ranking = nil
		err = nil
	}
	return ranking, err
}

func (database *Database) SaveRanking(ranking *Ranking) error {
	_, err := database.rankingMap.Update(ranking)
	return err
}

func (database *Database) DeleteRanking(ranking *Ranking) error {
	_, err := database.rankingMap.Delete(ranking)
	return err
}

func (database *Database) TruncateRankings() error {
	return database.rankingMap.TruncateTables()
}

func (database *Database) GetAllRankings() ([]Ranking, error) {
	var rankings []Ranking
	err := database.rankingMap.Select(&rankings, "SELECT * FROM rankings ORDER BY rank")
	return rankings, err
}

// Determines the rankings from the stored match results, and saves them to the database.
func (database *Database) CalculateRankings() error {
	matches, err := database.GetMatchesByType("qualification")
	if err != nil {
		return err
	}
	rankings := make(map[int]*Ranking)
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

	// Divide the total score by the number of matches played and truncate (floor) to two decimal places.
	for _, ranking := range rankings {
		ranking.QualificationAverage = float64(int(ranking.QualificationAverage*100/float64(ranking.Played))) / 100
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
		err = transaction.Insert(ranking)
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
func addMatchResultToRankings(rankings map[int]*Ranking, teamId int, matchResult *MatchResult, isRed bool) {
	ranking := rankings[teamId]
	if ranking == nil {
		ranking = &Ranking{TeamId: teamId}
		rankings[teamId] = ranking
	}
	ranking.Played += 1

	// Don't award any points if the team was disqualified.
	var cards map[string]string
	if isRed {
		cards = matchResult.RedCards
	} else {
		cards = matchResult.BlueCards
	}
	if card, ok := cards[strconv.Itoa(teamId)]; ok && card == "red" {
		ranking.Disqualifications += 1
		return
	}

	var score *ScoreSummary
	if isRed {
		score = matchResult.RedScoreSummary()
	} else {
		score = matchResult.BlueScoreSummary()
	}

	// Assign points.
	ranking.QualificationAverage += float64(score.Score)
	ranking.CoopertitionPoints += score.CoopertitionPoints
	ranking.AutoPoints += score.AutoPoints
	ranking.ContainerPoints += score.ContainerPoints
	ranking.TotePoints += score.TotePoints
	ranking.LitterPoints += score.LitterPoints

	// Store a random value to be used as the last tiebreaker if necessary.
	ranking.Random = rand.Float64()
}

func sortRankings(rankings map[int]*Ranking) Rankings {
	var sortedRankings Rankings
	for _, ranking := range rankings {
		sortedRankings = append(sortedRankings, ranking)
	}
	sort.Sort(sortedRankings)
	return sortedRankings
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Len() int {
	return len(rankings)
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Less(i, j int) bool {
	a := rankings[i]
	b := rankings[j]
	if a.QualificationAverage == b.QualificationAverage {
		// Use cross-multiplication to keep it in integer math.
		if a.CoopertitionPoints*b.Played == b.CoopertitionPoints*a.Played {
			if a.AutoPoints*b.Played == b.AutoPoints*a.Played {
				if a.ContainerPoints*b.Played == b.ContainerPoints*a.Played {
					if a.TotePoints*b.Played == b.TotePoints*a.Played {
						if a.LitterPoints*b.Played == b.LitterPoints*a.Played {
							return a.Random > b.Random
						}
						return a.LitterPoints*b.Played > b.LitterPoints*a.Played
					}
					return a.TotePoints*b.Played > b.TotePoints*a.Played
				}
				return a.ContainerPoints*b.Played > b.ContainerPoints*a.Played
			}
			return a.AutoPoints*b.Played > b.AutoPoints*a.Played
		}
		return a.CoopertitionPoints*b.Played > b.CoopertitionPoints*a.Played
	}
	return a.QualificationAverage > b.QualificationAverage
}

// Helper function to implement the required interface for Sort.
func (rankings Rankings) Swap(i, j int) {
	rankings[i], rankings[j] = rankings[j], rankings[i]
}
