// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for controlling match play.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/bracket"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type MatchPlayListItem struct {
	Id          int
	DisplayName string
	Time        string
	Status      game.MatchStatus
	ColorClass  string
}

type MatchPlayList []MatchPlayListItem

// Shows the match play control interface.
func (web *Web) matchPlayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	practiceMatches, err := web.buildMatchPlayList("practice")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := web.buildMatchPlayList("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eliminationMatches, err := web.buildMatchPlayList("elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/match_play.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[string]MatchPlayList{"practice": practiceMatches,
		"qualification": qualificationMatches, "elimination": eliminationMatches}
	currentMatchType := web.arena.CurrentMatch.Type
	if currentMatchType == "test" {
		currentMatchType = "practice"
	}
	redOffFieldTeams, blueOffFieldTeams, err := web.arena.Database.GetOffFieldTeamIds(web.arena.CurrentMatch)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchResult, err := web.arena.Database.GetMatchResultForMatch(web.arena.CurrentMatch.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	isReplay := matchResult != nil
	data := struct {
		*model.EventSettings
		PlcIsEnabled          bool
		MatchesByType         map[string]MatchPlayList
		CurrentMatchType      string
		Match                 *model.Match
		RedOffFieldTeams      []int
		BlueOffFieldTeams     []int
		AllowSubstitution     bool
		IsReplay              bool
		SavedMatchType        string
		SavedMatch            *model.Match
		PlcArmorBlockStatuses map[string]bool
	}{
		web.arena.EventSettings,
		web.arena.Plc.IsEnabled(),
		matchesByType,
		currentMatchType,
		web.arena.CurrentMatch,
		redOffFieldTeams,
		blueOffFieldTeams,
		web.arena.CurrentMatch.ShouldAllowSubstitution(),
		isReplay,
		web.arena.SavedMatch.CapitalizedType(),
		web.arena.SavedMatch,
		web.arena.Plc.GetArmorBlockStatuses(),
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Loads the given match onto the arena in preparation for playing it.
func (web *Web) matchPlayLoadHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	var match *model.Match
	var err error
	if matchId == 0 {
		err = web.arena.LoadTestMatch()
	} else {
		match, err = web.arena.Database.GetMatchById(matchId)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		if match == nil {
			handleWebErr(w, fmt.Errorf("Invalid match ID %d.", matchId))
			return
		}
		err = web.arena.LoadMatch(match)
	}
	if err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/match_play", 303)
}

// Loads the results for the given match into the display buffer.
func (web *Web) matchPlayShowResultHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	match, err := web.arena.Database.GetMatchById(matchId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if match == nil {
		handleWebErr(w, fmt.Errorf("Invalid match ID %d.", matchId))
		return
	}
	matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if matchResult == nil {
		handleWebErr(w, fmt.Errorf("No result found for match ID %d.", matchId))
		return
	}
	if match.ShouldUpdateRankings() {
		web.arena.SavedRankings, err = web.arena.Database.GetAllRankings()
		if err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		web.arena.SavedRankings = game.Rankings{}
	}
	web.arena.SavedMatch = match
	web.arena.SavedMatchResult = matchResult
	web.arena.ScorePostedNotifier.Notify()

	http.Redirect(w, r, "/match_play", 303)
}

// Clears the match results display buffer.
func (web *Web) matchPlayClearResultHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	// Load an empty match to effectively clear the buffer.
	web.arena.SavedMatch = &model.Match{}
	web.arena.SavedMatchResult = model.NewMatchResult()
	web.arena.ScorePostedNotifier.Notify()

	http.Redirect(w, r, "/match_play", 303)
}

// The websocket endpoint for the match play client to send control commands and receive status updates.
func (web *Web) matchPlayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchTimingNotifier, web.arena.ArenaStatusNotifier, web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier, web.arena.ScoringStatusNotifier, web.arena.AudienceDisplayModeNotifier,
		web.arena.AllianceStationDisplayModeNotifier, web.arena.EventStatusNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "substituteTeam":
			args := struct {
				Team     int
				Position string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.SubstituteTeam(args.Team, args.Position)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "toggleBypass":
			station, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			if _, ok := web.arena.AllianceStations[station]; !ok {
				ws.WriteError(fmt.Sprintf("Invalid alliance station '%s'.", station))
				continue
			}
			web.arena.AllianceStations[station].Bypass = !web.arena.AllianceStations[station].Bypass
		case "startMatch":
			args := struct {
				MuteMatchSounds bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			web.arena.MuteMatchSounds = args.MuteMatchSounds
			err = web.arena.StartMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "abortMatch":
			err = web.arena.AbortMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "signalVolunteers":
			if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldVolunteers = true
			continue // Don't reload.
		case "signalReset":
			if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
			continue // Don't reload.
		case "commitResults":
			err = web.commitCurrentMatchScore()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
			if err != nil {
				log.Println(err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "discardResults":
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
			if err != nil {
				log.Println(err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "setAudienceDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAudienceDisplayMode(mode)
			continue
		case "setAllianceStationDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAllianceStationDisplayMode(mode)
			continue
		case "startTimeout":
			durationSec, ok := data.(float64)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			err = web.arena.StartTimeout(int(durationSec))
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "setTestMatchName":
			if web.arena.CurrentMatch.Type != "test" {
				// Don't allow changing the name of a non-test match.
				continue
			}
			name, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.CurrentMatch.DisplayName = name
			web.arena.MatchLoadNotifier.Notify()
			continue
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Send out the status again after handling the command, as it most likely changed as a result.
		err = ws.WriteNotifier(web.arena.ArenaStatusNotifier)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

// Saves the given match and result to the database, supplanting any previous result for the match.
func (web *Web) commitMatchScore(match *model.Match, matchResult *model.MatchResult, isMatchReviewEdit bool) error {
	var updatedRankings game.Rankings

	if match.Type == "elimination" {
		// Adjust the score if necessary for an elimination DQ.
		matchResult.CorrectEliminationScore()
	}

	if match.Type != "test" {
		if matchResult.PlayNumber == 0 {
			// Determine the play number for this new match result.
			prevMatchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if prevMatchResult != nil {
				matchResult.PlayNumber = prevMatchResult.PlayNumber + 1
			} else {
				matchResult.PlayNumber = 1
			}

			// Save the match result record to the database.
			err = web.arena.Database.CreateMatchResult(matchResult)
			if err != nil {
				return err
			}
		} else {
			// We are updating a match result record that already exists.
			err := web.arena.Database.UpdateMatchResult(matchResult)
			if err != nil {
				return err
			}
		}

		// Update and save the match record to the database.
		match.ScoreCommittedAt = time.Now()
		redScoreSummary := matchResult.RedScoreSummary()
		blueScoreSummary := matchResult.BlueScoreSummary()
		applyElimTiebreakers := false
		if match.Type == "elimination" {
			// Playoff matches other than the finals should have ties broken by examining the scoring breakdown rather
			// than being replayed.
			if match.ElimRound < web.arena.PlayoffBracket.FinalsMatchup.Round {
				applyElimTiebreakers = true
			}
		}
		match.Status = game.DetermineMatchStatus(redScoreSummary, blueScoreSummary, applyElimTiebreakers)
		err := web.arena.Database.UpdateMatch(match)
		if err != nil {
			return err
		}

		if match.ShouldUpdateCards() {
			// Regenerate the residual yellow cards that teams may carry.
			if err = tournament.CalculateTeamCards(web.arena.Database, match.Type); err != nil {
				return err
			}
		}

		if match.ShouldUpdateRankings() {
			// Recalculate all the rankings.
			rankings, err := tournament.CalculateRankings(web.arena.Database, isMatchReviewEdit)
			if err != nil {
				return err
			}
			updatedRankings = rankings
		}

		if match.ShouldUpdateEliminationMatches() {
			if err = web.arena.Database.UpdateAllianceFromMatch(
				match.ElimRedAlliance, [3]int{match.Red1, match.Red2, match.Red3},
			); err != nil {
				return err
			}
			if err = web.arena.Database.UpdateAllianceFromMatch(
				match.ElimBlueAlliance, [3]int{match.Blue1, match.Blue2, match.Blue3},
			); err != nil {
				return err
			}

			// Generate any subsequent elimination matches.
			nextMatchTime := time.Now().Add(time.Second * bracket.ElimMatchSpacingSec)
			if err = web.arena.UpdatePlayoffBracket(&nextMatchTime); err != nil {
				return err
			}

			// Generate awards if the tournament is over.
			if web.arena.PlayoffBracket.IsComplete() {
				winnerAllianceId := web.arena.PlayoffBracket.Winner()
				finalistAllianceId := web.arena.PlayoffBracket.Finalist()
				if err = tournament.CreateOrUpdateWinnerAndFinalistAwards(
					web.arena.Database, winnerAllianceId, finalistAllianceId,
				); err != nil {
					return err
				}
			}
		}

		if web.arena.EventSettings.TbaPublishingEnabled && match.Type != "practice" {
			// Publish asynchronously to The Blue Alliance.
			go func() {
				if err = web.arena.TbaClient.PublishMatches(web.arena.Database); err != nil {
					log.Printf("Failed to publish matches: %s", err.Error())
				}
				if match.ShouldUpdateRankings() {
					if err = web.arena.TbaClient.PublishRankings(web.arena.Database); err != nil {
						log.Printf("Failed to publish rankings: %s", err.Error())
					}
				}
			}()
		}

		// Back up the database, but don't error out if it fails.
		err = web.arena.Database.Backup(web.arena.EventSettings.Name,
			fmt.Sprintf("post_%s_match_%s", match.Type, match.DisplayName))
		if err != nil {
			log.Println(err)
		}
	}

	if !isMatchReviewEdit {
		// Store the result in the buffer to be shown in the audience display.
		web.arena.SavedMatch = match
		web.arena.SavedMatchResult = matchResult
		web.arena.SavedRankings = updatedRankings
		web.arena.ScorePostedNotifier.Notify()
	}

	return nil
}

func (web *Web) getCurrentMatchResult() *model.MatchResult {
	return &model.MatchResult{MatchId: web.arena.CurrentMatch.Id, MatchType: web.arena.CurrentMatch.Type,
		RedScore: &web.arena.RedRealtimeScore.CurrentScore, BlueScore: &web.arena.BlueRealtimeScore.CurrentScore,
		RedCards: web.arena.RedRealtimeScore.Cards, BlueCards: web.arena.BlueRealtimeScore.Cards}
}

// Saves the realtime result as the final score for the match currently loaded into the arena.
func (web *Web) commitCurrentMatchScore() error {
	return web.commitMatchScore(web.arena.CurrentMatch, web.getCurrentMatchResult(), false)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Len() int {
	return len(list)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Less(i, j int) bool {
	return list[i].Status == game.MatchNotPlayed && list[j].Status != game.MatchNotPlayed
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Constructs the list of matches to display on the side of the match play interface.
func (web *Web) buildMatchPlayList(matchType string) (MatchPlayList, error) {
	matches, err := web.arena.Database.GetMatchesByType(matchType)
	if err != nil {
		return MatchPlayList{}, err
	}

	matchPlayList := make(MatchPlayList, len(matches))
	for i, match := range matches {
		matchPlayList[i].Id = match.Id
		matchPlayList[i].DisplayName = match.TypePrefix() + match.DisplayName
		matchPlayList[i].Time = match.Time.Local().Format("3:04 PM")
		matchPlayList[i].Status = match.Status
		switch match.Status {
		case game.RedWonMatch:
			matchPlayList[i].ColorClass = "danger"
		case game.BlueWonMatch:
			matchPlayList[i].ColorClass = "info"
		case game.TieMatch:
			matchPlayList[i].ColorClass = "warning"
		default:
			matchPlayList[i].ColorClass = ""
		}
		if web.arena.CurrentMatch != nil && matchPlayList[i].Id == web.arena.CurrentMatch.Id {
			matchPlayList[i].ColorClass = "success"
		}
	}

	// Sort the list to put all completed matches at the bottom.
	sort.Stable(matchPlayList)

	return matchPlayList, nil
}
