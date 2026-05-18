// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for editing match results.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"net/http"
	"strconv"
)

type MatchReviewListItem struct {
	Id          int
	ShortName   string
	Time        string
	RedTeams    []int
	BlueTeams   []int
	RedScore    int
	BlueScore   int
	RedSummary  *game.ScoreSummary
	BlueSummary *game.ScoreSummary
	ColorClass  string
	IsComplete  bool
}

type MatchReviewEditAlliance struct {
	Alliance          string
	Teams             []int
	Summary           *game.ScoreSummary
	ShowRankingPoints bool
}

type MatchReviewSummaryResponse struct {
	RedSummary  *game.ScoreSummary
	BlueSummary *game.ScoreSummary
}

// Shows the match review interface.
func (web *Web) matchReviewHandler(w http.ResponseWriter, r *http.Request) {
	practiceMatches, err := web.buildMatchReviewList(model.Practice)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := web.buildMatchReviewList(model.Qualification)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	playoffMatches, err := web.buildMatchReviewList(model.Playoff)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/match_review.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[model.MatchType][]MatchReviewListItem{
		model.Practice:      practiceMatches,
		model.Qualification: qualificationMatches,
		model.Playoff:       playoffMatches,
	}
	currentMatchType := web.arena.CurrentMatch.Type
	if currentMatchType == model.Test {
		currentMatchType = model.Practice
	}
	data := struct {
		*model.EventSettings
		MatchesByType    map[model.MatchType][]MatchReviewListItem
		CurrentMatchType model.MatchType
	}{web.arena.EventSettings, matchesByType, currentMatchType}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Shows the page to edit the results for a match.
func (web *Web) matchReviewEditGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	match, matchResult, isCurrent, err := web.getMatchResultFromRequest(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/edit_match_result.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchResultJson, err := json.Marshal(matchResult)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	alliances := []MatchReviewEditAlliance{
		{
			Alliance:          "red",
			Teams:             []int{match.Red1, match.Red2, match.Red3},
			Summary:           matchResult.RedScoreSummary(),
			ShowRankingPoints: match.Type != model.Playoff,
		},
		{
			Alliance:          "blue",
			Teams:             []int{match.Blue1, match.Blue2, match.Blue3},
			Summary:           matchResult.BlueScoreSummary(),
			ShowRankingPoints: match.Type != model.Playoff,
		},
	}
	data := struct {
		*model.EventSettings
		Match           *model.Match
		MatchResultJson string
		IsCurrentMatch  bool
		Rules           map[int]*game.Rule
		Alliances       []MatchReviewEditAlliance
	}{web.arena.EventSettings, match, string(matchResultJson), isCurrent, game.GetAllRules(), alliances}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Calculates score summaries for an in-progress match result without saving it.
func (web *Web) matchReviewSummaryPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	match, _, _, err := web.getMatchResultFromRequest(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	var matchResult model.MatchResult
	if err = json.NewDecoder(r.Body).Decode(&matchResult); err != nil {
		handleWebErr(w, err)
		return
	}
	if matchResult.MatchId != match.Id {
		handleWebErr(w, fmt.Errorf("Error: match ID %d from result does not match expected", matchResult.MatchId))
		return
	}
	normalizeMatchResult(&matchResult)

	response := MatchReviewSummaryResponse{
		RedSummary:  matchResult.RedScoreSummary(),
		BlueSummary: matchResult.BlueScoreSummary(),
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates the results for a match.
func (web *Web) matchReviewEditPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	match, _, isCurrent, err := web.getMatchResultFromRequest(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	var matchResult model.MatchResult
	if err = json.Unmarshal([]byte(r.PostFormValue("matchResultJson")), &matchResult); err != nil {
		handleWebErr(w, err)
		return
	}
	if matchResult.MatchId != match.Id {
		handleWebErr(w, fmt.Errorf("Error: match ID %d from result does not match expected", matchResult.MatchId))
		return
	}
	normalizeMatchResult(&matchResult)

	if isCurrent {
		// If editing the current match, just save it back to memory.
		web.arena.RedRealtimeScore.CurrentScore = *matchResult.RedScore
		web.arena.BlueRealtimeScore.CurrentScore = *matchResult.BlueScore
		web.arena.RedRealtimeScore.Cards = matchResult.RedCards
		web.arena.BlueRealtimeScore.Cards = matchResult.BlueCards

		web.arena.RealtimeScoreNotifier.Notify()

		http.Redirect(w, r, "/match_play", 303)
	} else {
		err = web.commitMatchScore(match, &matchResult, true)
		if err != nil {
			handleWebErr(w, err)
			return
		}

		http.Redirect(w, r, "/match_review", 303)
	}
}

// Load the match result for the match referenced in the HTTP query string.
func (web *Web) getMatchResultFromRequest(r *http.Request) (*model.Match, *model.MatchResult, bool, error) {
	// If editing the current match, get it from memory instead of the DB.
	if r.PathValue("matchId") == "current" {
		return web.arena.CurrentMatch, web.getCurrentMatchResult(), true, nil
	}

	matchId, _ := strconv.Atoi(r.PathValue("matchId"))
	match, err := web.arena.Database.GetMatchById(matchId)
	if err != nil {
		return nil, nil, false, err
	}
	if match == nil {
		return nil, nil, false, fmt.Errorf("Error: No such match: %d", matchId)
	}
	matchResult, err := web.arena.Database.GetMatchResultForMatch(matchId)
	if err != nil {
		return nil, nil, false, err
	}
	if matchResult == nil {
		// We're scoring a match that hasn't been played yet, but that's okay.
		matchResult = model.NewMatchResult()
		matchResult.MatchId = matchId
		matchResult.MatchType = match.Type
	}

	return match, matchResult, false, nil
}

func normalizeMatchResult(matchResult *model.MatchResult) {
	if matchResult.RedScore == nil {
		matchResult.RedScore = new(game.Score)
	}
	if matchResult.BlueScore == nil {
		matchResult.BlueScore = new(game.Score)
	}
	if matchResult.RedCards == nil {
		matchResult.RedCards = make(map[string]string)
	}
	if matchResult.BlueCards == nil {
		matchResult.BlueCards = make(map[string]string)
	}
}

// Constructs the list of matches to display in the match review interface.
func (web *Web) buildMatchReviewList(matchType model.MatchType) ([]MatchReviewListItem, error) {
	matches, err := web.arena.Database.GetMatchesByType(matchType, false)
	if err != nil {
		return []MatchReviewListItem{}, err
	}

	matchReviewList := make([]MatchReviewListItem, len(matches))
	for i, match := range matches {
		matchReviewList[i].Id = match.Id
		matchReviewList[i].ShortName = match.ShortName
		matchReviewList[i].Time = match.Time.Local().Format("Mon 1/02 03:04 PM")
		matchReviewList[i].RedTeams = []int{match.Red1, match.Red2, match.Red3}
		matchReviewList[i].BlueTeams = []int{match.Blue1, match.Blue2, match.Blue3}
		matchReviewList[i].RedSummary = new(game.ScoreSummary)
		matchReviewList[i].BlueSummary = new(game.ScoreSummary)
		matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return []MatchReviewListItem{}, err
		}
		if matchResult != nil {
			matchReviewList[i].RedSummary = matchResult.RedScoreSummary()
			matchReviewList[i].BlueSummary = matchResult.BlueScoreSummary()
			matchReviewList[i].RedScore = matchReviewList[i].RedSummary.Score
			matchReviewList[i].BlueScore = matchReviewList[i].BlueSummary.Score
		}
		switch match.Status {
		case game.RedWonMatch:
			matchReviewList[i].ColorClass = "red"
			matchReviewList[i].IsComplete = true
		case game.BlueWonMatch:
			matchReviewList[i].ColorClass = "blue"
			matchReviewList[i].IsComplete = true
		case game.TieMatch:
			matchReviewList[i].ColorClass = "yellow"
			matchReviewList[i].IsComplete = true
		default:
			matchReviewList[i].ColorClass = ""
			matchReviewList[i].IsComplete = false
		}
	}

	return matchReviewList, nil
}
