// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for controlling match play.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
)

type MatchPlayListItem struct {
	Id          int
	DisplayName string
	Time        string
	ColorClass  string
}

type MatchPlayList []MatchPlayListItem

// Global var to hold the current active tournament so that its matches are displayed by default.
var currentMatchType string

// Shows the match play control interface.
func MatchPlayHandler(w http.ResponseWriter, r *http.Request) {
	practiceMatches, err := buildMatchPlayList("practice")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := buildMatchPlayList("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eliminationMatches, err := buildMatchPlayList("elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := template.ParseFiles("templates/match_play.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[string]MatchPlayList{"practice": practiceMatches,
		"qualification": qualificationMatches, "elimination": eliminationMatches}
	if currentMatchType == "" {
		currentMatchType = "practice"
	}
	data := struct {
		*EventSettings
		MatchesByType    map[string]MatchPlayList
		CurrentMatchType string
	}{eventSettings, matchesByType, currentMatchType}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func MatchPlayFakeResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	match, err := db.GetMatchById(matchId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if match == nil {
		handleWebErr(w, fmt.Errorf("Invalid match ID %d.", matchId))
		return
	}
	matchResult := MatchResult{MatchId: match.Id}
	matchResult.RedScore = randomScore()
	matchResult.BlueScore = randomScore()
	err = CommitMatchScore(match, &matchResult)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	currentMatchType = match.Type

	http.Redirect(w, r, "/match_play", 302)
}

func CommitMatchScore(match *Match, matchResult *MatchResult) error {
	// Determine the play number for this match.
	prevMatchResult, err := db.GetMatchResultForMatch(match.Id)
	if err != nil {
		return err
	}
	if prevMatchResult != nil {
		matchResult.PlayNumber = prevMatchResult.PlayNumber + 1
	} else {
		matchResult.PlayNumber = 1
	}

	// Save the match result record to the database.
	err = db.CreateMatchResult(matchResult)
	if err != nil {
		return err
	}

	// Update and save the match record to the database.
	match.Status = "complete"
	redScore := matchResult.RedScoreSummary()
	blueScore := matchResult.BlueScoreSummary()
	if redScore.Score > blueScore.Score {
		match.Winner = "R"
	} else if redScore.Score < blueScore.Score {
		match.Winner = "B"
	} else {
		match.Winner = "T"
	}
	err = db.SaveMatch(match)
	if err != nil {
		return err
	}

	// Recalculate all the rankings.
	err = db.CalculateRankings()
	if err != nil {
		return err
	}

	return nil
}

func (list MatchPlayList) Len() int {
	return len(list)
}

func (list MatchPlayList) Less(i, j int) bool {
	return list[i].ColorClass == "" && list[j].ColorClass != ""
}

func (list MatchPlayList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func buildMatchPlayList(matchType string) (MatchPlayList, error) {
	matches, err := db.GetMatchesByType(matchType)
	if err != nil {
		return MatchPlayList{}, err
	}

	prefix := ""
	if matchType == "practice" {
		prefix = "P"
	} else if matchType == "qualification" {
		prefix = "Q"
	}
	matchPlayList := make(MatchPlayList, len(matches))
	for i, match := range matches {
		matchPlayList[i].Id = match.Id
		matchPlayList[i].DisplayName = prefix + match.DisplayName
		matchPlayList[i].Time = match.Time.Format("3:04 PM")
		switch match.Winner {
		case "R":
			matchPlayList[i].ColorClass = "danger"
		case "B":
			matchPlayList[i].ColorClass = "info"
		case "T":
			matchPlayList[i].ColorClass = "warning"
		default:
			matchPlayList[i].ColorClass = ""
		}
	}

	// Sort the list to put all completed matches at the bottom.
	sort.Stable(matchPlayList)

	return matchPlayList, nil
}

func randomScore() Score {
	cycle := Cycle{rand.Intn(3) + 1, rand.Intn(2) == 1, rand.Intn(2) == 1, rand.Intn(2) == 1, rand.Intn(2) == 1,
		rand.Intn(2) == 1}
	return Score{rand.Intn(4), rand.Intn(4), rand.Intn(4), rand.Intn(4), rand.Intn(4), 0, 0, []Cycle{cycle}}
}
