// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web API for providing JSON-formatted event data.

package main

import (
	"encoding/json"
	"io"
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

	data, err := json.MarshalIndent(rankings, "", "  ")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = io.WriteString(w, string(data))
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
