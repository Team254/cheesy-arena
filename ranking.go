// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for team ranking data at an event.

package main

import (
	"math/rand"
	"sort"
)

type Ranking struct {
	TeamId             int
	Rank               int
	QualificationScore int
	AssistPoints       int
	AutoPoints         int
	TrussCatchPoints   int
	GoalFoulPoints     int
	Random             float64
	Wins               int
	Losses             int
	Ties               int
	Disqualifications  int
	Played             int
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
	sortedRankings := sortRankings(rankings)

	// Stuff the match results into the database in an atomic operation.
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

func addMatchResultToRankings(rankings map[int]*Ranking, teamId int, matchResult *MatchResult, isRed bool) {
	ranking := rankings[teamId]
	if ranking == nil {
		ranking = &Ranking{TeamId: teamId}
		rankings[teamId] = ranking
	}
	ranking.Played += 1

	// Don't award any points if the team was disqualified.
	for _, dqTeamId := range matchResult.Cards.RedCardTeamIds {
		if teamId == dqTeamId {
			ranking.Disqualifications += 1
			return
		}
	}

	var ownScore, opponentScore *ScoreSummary
	if isRed {
		ownScore = matchResult.RedScoreSummary()
		opponentScore = matchResult.BlueScoreSummary()
	} else {
		ownScore = matchResult.BlueScoreSummary()
		opponentScore = matchResult.RedScoreSummary()
	}

	// Assign win/loss/tie points.
	if ownScore.Score > opponentScore.Score {
		ranking.QualificationScore += 2
		ranking.Wins += 1
	} else if ownScore.Score < opponentScore.Score {
		ranking.Losses += 1
	} else {
		ranking.QualificationScore += 1
		ranking.Ties += 1
	}

	// Assign tiebreakers points.
	ranking.AssistPoints += ownScore.AssistPoints
	ranking.AutoPoints += ownScore.AutoPoints
	ranking.TrussCatchPoints += ownScore.TrussCatchPoints
	ranking.GoalFoulPoints += ownScore.GoalPoints + ownScore.FoulPoints

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

func (rankings Rankings) Len() int {
	return len(rankings)
}

func (rankings Rankings) Less(i, j int) bool {
	a := rankings[i]
	b := rankings[j]
	if a.QualificationScore*b.Played == b.QualificationScore*a.Played {
		if a.AssistPoints*b.Played == b.AssistPoints*a.Played {
			if a.AutoPoints*b.Played == b.AutoPoints*a.Played {
				if a.TrussCatchPoints*b.Played == b.TrussCatchPoints*a.Played {
					if a.GoalFoulPoints*b.Played == b.GoalFoulPoints*a.Played {
						return a.Random > b.Random
					}
					// Use cross-multiplication to keep it in integer math.
					return a.GoalFoulPoints*b.Played > b.GoalFoulPoints*a.Played
				}
				return a.TrussCatchPoints*b.Played > b.TrussCatchPoints*a.Played
			}
			return a.AutoPoints*b.Played > b.AutoPoints*a.Played
		}
		return a.AssistPoints*b.Played > b.AssistPoints*a.Played
	}
	return a.QualificationScore*b.Played > b.QualificationScore*a.Played
}

func (rankings Rankings) Swap(i, j int) {
	rankings[i], rankings[j] = rankings[j], rankings[i]
}
