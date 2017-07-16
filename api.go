// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web API for providing JSON-formatted event data.

package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type MatchResultWithSummary struct {
	MatchResult
	RedSummary  *ScoreSummary
	BlueSummary *ScoreSummary
}

type MatchWithResult struct {
	Match
	Result *MatchResultWithSummary
}

type RankingWithNickname struct {
	Ranking
	Nickname string
}

// Generates a JSON dump of the matches and results.
func MatchesApiHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	vars := mux.Vars(r)
	matches, err := db.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	matchesWithResults := make([]MatchWithResult, len(matches))
	for i, match := range matches {
		matchesWithResults[i].Match = match
		matchResult, err := db.GetMatchResultForMatch(match.Id)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		var matchResultWithSummary *MatchResultWithSummary
		if matchResult != nil {
			matchResultWithSummary = &MatchResultWithSummary{MatchResult: *matchResult}
			matchResultWithSummary.RedSummary = matchResult.RedScoreSummary()
			matchResultWithSummary.BlueSummary = matchResult.BlueScoreSummary()
		}
		matchesWithResults[i].Result = matchResultWithSummary
	}

	jsonData, err := json.MarshalIndent(matchesWithResults, "", "  ")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a JSON dump of the sponsor slides for use by the audience display.
func SponsorSlidesApiHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	sponsors, err := db.GetAllSponsorSlides()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	jsonData, err := json.MarshalIndent(sponsors, "", "  ")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a JSON dump of the qualification rankings, primarily for use by the pit display.
func RankingsApiHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	rankings, err := db.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var rankingsWithNicknames []RankingWithNickname
	if rankings == nil {
		// Go marshals an empty slice to null, so explicitly create it so that it appears as an empty JSON array.
		rankingsWithNicknames = make([]RankingWithNickname, 0)
	} else {
		rankingsWithNicknames = make([]RankingWithNickname, len(rankings))
	}

	// Get team info so that nicknames can be displayed.
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	teamNicknames := make(map[int]string)
	for _, team := range teams {
		teamNicknames[team.Id] = team.Nickname
	}
	for i, ranking := range rankings {
		rankingsWithNicknames[i] = RankingWithNickname{ranking, teamNicknames[ranking.TeamId]}
	}

	// Get the last match scored so we can report that on the display.
	matches, err := db.GetMatchesByType("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	highestPlayedMatch := ""
	for _, match := range matches {
		if match.Status == "complete" {
			highestPlayedMatch = match.DisplayName
		}
	}

	data := struct {
		Rankings           []RankingWithNickname
		HighestPlayedMatch string
	}{rankingsWithNicknames, highestPlayedMatch}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
