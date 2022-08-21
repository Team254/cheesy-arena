// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web API for providing JSON-formatted event data.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"strconv"
)

type MatchResultWithSummary struct {
	model.MatchResult
	RedSummary  *game.ScoreSummary
	BlueSummary *game.ScoreSummary
}

type MatchWithResult struct {
	model.Match
	Result *MatchResultWithSummary
}

type RankingWithNickname struct {
	game.Ranking
	Nickname string
}

type allianceMatchup struct {
	Round              int
	Group              int
	DisplayName        string
	RedAllianceSource  string
	BlueAllianceSource string
	RedAlliance        *model.Alliance
	BlueAlliance       *model.Alliance
	IsActive           bool
	SeriesLeader       string
	SeriesStatus       string
}

// Generates a JSON dump of the matches and results.
func (web *Web) matchesApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matches, err := web.arena.Database.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	matchesWithResults := make([]MatchWithResult, len(matches))
	for i, match := range matches {
		matchesWithResults[i].Match = match
		matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
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
func (web *Web) sponsorSlidesApiHandler(w http.ResponseWriter, r *http.Request) {
	sponsors, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	if sponsors == nil {
		// Go marshals an empty slice to null, so explicitly create it so that it appears as an empty JSON array.
		sponsors = make([]model.SponsorSlide, 0)
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

// Generates a JSON dump of the qualification rankings, primarily for use by the rankings display.
func (web *Web) rankingsApiHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
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
	teams, err := web.arena.Database.GetAllTeams()
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
	matches, err := web.arena.Database.GetMatchesByType("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	highestPlayedMatch := ""
	for _, match := range matches {
		if match.IsComplete() {
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

// Generates a JSON dump of the alliances.
func (web *Web) alliancesApiHandler(w http.ResponseWriter, r *http.Request) {
	alliances, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	jsonData, err := json.MarshalIndent(alliances, "", "  ")
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

// Websocket API for receiving arena status updates.
func (web *Web) arenaWebsocketApiHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.MatchTimingNotifier, web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier)
}

// Serves the avatar for a given team, or a default if none exists.
func (web *Web) teamAvatarsApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, err := strconv.Atoi(vars["teamId"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	avatarPath := fmt.Sprintf("%s/%d.png", partner.AvatarsDir, teamId)
	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		avatarPath = fmt.Sprintf("%s/0.png", partner.AvatarsDir)
	}

	http.ServeFile(w, r, avatarPath)
}

func (web *Web) bracketSvgApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	if err := web.generateBracketSvg(w); err != nil {
		handleWebErr(w, err)
		return
	}
}

func (web *Web) generateBracketSvg(w io.Writer) error {
	alliances, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		return err
	}
	activeMatch := web.arena.SavedMatch

	matchups := make(map[string]*allianceMatchup)
	if web.arena.PlayoffBracket != nil {
		for _, matchup := range web.arena.PlayoffBracket.GetAllMatchups() {
			allianceMatchup := allianceMatchup{
				Round:              matchup.Round,
				Group:              matchup.Group,
				DisplayName:        matchup.LongDisplayName(),
				RedAllianceSource:  matchup.RedAllianceSourceDisplayName(),
				BlueAllianceSource: matchup.BlueAllianceSourceDisplayName(),
			}
			if matchup.RedAllianceId > 0 {
				if len(alliances) > 0 {
					allianceMatchup.RedAlliance = &alliances[matchup.RedAllianceId-1]
				} else {
					allianceMatchup.RedAlliance = &model.Alliance{Id: matchup.RedAllianceId}
				}
			}
			if matchup.BlueAllianceId > 0 {
				if len(alliances) > 0 {
					allianceMatchup.BlueAlliance = &alliances[matchup.BlueAllianceId-1]
				} else {
					allianceMatchup.BlueAlliance = &model.Alliance{Id: matchup.BlueAllianceId}
				}
			}
			if activeMatch != nil {
				allianceMatchup.IsActive = activeMatch.ElimRound == matchup.Round &&
					activeMatch.ElimGroup == matchup.Group
			}
			allianceMatchup.SeriesLeader, allianceMatchup.SeriesStatus = matchup.StatusText()
			matchups[fmt.Sprintf("%d_%d", matchup.Round, matchup.Group)] = &allianceMatchup
		}
	}

	bracketType := "double"
	numAlliances := web.arena.EventSettings.NumElimAlliances
	if web.arena.EventSettings.ElimType == "single" {
		if numAlliances > 8 {
			bracketType = "16"
		} else if numAlliances > 4 {
			bracketType = "8"
		} else if numAlliances > 2 {
			bracketType = "4"
		} else {
			bracketType = "2"
		}
	}

	template, err := web.parseFiles("templates/bracket.svg")
	if err != nil {
		return err
	}
	data := struct {
		BracketType string
		Matchups    map[string]*allianceMatchup
	}{bracketType, matchups}
	return template.ExecuteTemplate(w, "bracket", data)
}
