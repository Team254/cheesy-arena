// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web API for providing JSON-formatted event data.

package main

import (
	"encoding/json"
	"net/http"
)

// Generates a JSON dump of the qualification rankings.
func RankingsApiHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := db.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if rankings == nil {
		// Go marshals an empty slice to null, so explicitly create it so that it appears as an empty JSON array.
		rankings = make([]Ranking, 0)
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
		Rankings           []Ranking
		HighestPlayedMatch string
	}{rankings, highestPlayedMatch}
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
