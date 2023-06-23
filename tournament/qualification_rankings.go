// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for calculating the qualification rankings.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"sort"
	"strconv"
)

// Determines the rankings from the stored match results, and saves them to the database.
func CalculateRankings(database *model.Database, preservePreviousRank bool) (game.Rankings, error) {
	matches, err := database.GetMatchesByType(model.Qualification, false)
	if err != nil {
		return nil, err
	}
	rankings := make(map[int]*game.Ranking)
	for _, match := range matches {
		if !match.IsComplete() {
			continue
		}
		matchResult, err := database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return nil, err
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

	// Retrieve old rankings so that we can display changes in rank as a result of this calculation.
	oldRankings, err := database.GetAllRankings()
	if err != nil {
		return nil, err
	}
	oldRankingsMap := make(map[int]game.Ranking, len(oldRankings))
	for _, ranking := range oldRankings {
		oldRankingsMap[ranking.TeamId] = ranking
	}

	sortedRankings := sortRankings(rankings)
	for rank, ranking := range sortedRankings {
		sortedRankings[rank].Rank = rank + 1
		if oldRank, ok := oldRankingsMap[ranking.TeamId]; ok {
			if preservePreviousRank {
				sortedRankings[rank].PreviousRank = oldRank.PreviousRank
			} else {
				sortedRankings[rank].PreviousRank = oldRank.Rank
			}
		}
	}
	err = database.ReplaceAllRankings(sortedRankings)
	if err != nil {
		return nil, err
	}

	return sortedRankings, nil
}

// Checks all the match results for yellow and red cards, and updates the team model accordingly.
func CalculateTeamCards(database *model.Database, matchType model.MatchType) error {
	teams, err := database.GetAllTeams()
	if err != nil {
		return err
	}
	teamsMap := make(map[string]model.Team)
	for _, team := range teams {
		team.YellowCard = false
		teamsMap[strconv.Itoa(team.Id)] = team
	}

	matches, err := database.GetMatchesByType(matchType, false)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if !match.IsComplete() {
			continue
		}
		matchResult, err := database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		if matchResult == nil {
			return fmt.Errorf("found no match result for match %d", match.Id)
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
		err = database.UpdateTeam(&team)
		if err != nil {
			return err
		}
	}

	return nil
}

// Incrementally accounts for the given match result in the set of rankings that are being built.
func addMatchResultToRankings(
	rankings map[int]*game.Ranking, teamId int, matchResult *model.MatchResult, isRed bool,
) {
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
		sortedRankings = append(sortedRankings, *ranking)
	}
	sort.Sort(sortedRankings)
	return sortedRankings
}
