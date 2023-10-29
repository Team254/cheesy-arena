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
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

type MatchPlayListItem struct {
	Id         int
	ShortName  string
	Time       string
	Status     game.MatchStatus
	ColorClass string
}

type MatchPlayList []MatchPlayListItem

// Shows the match play control interface.
func (web *Web) matchPlayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/match_play.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled          bool
		PlcArmorBlockStatuses map[string]bool
	}{
		web.arena.EventSettings,
		web.arena.Plc.IsEnabled(),
		web.arena.Plc.GetArmorBlockStatuses(),
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders a partial template containing the list of matches.
func (web *Web) matchPlayMatchLoadHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	practiceMatches, err := web.buildMatchPlayList(model.Practice)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := web.buildMatchPlayList(model.Qualification)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	playoffMatches, err := web.buildMatchPlayList(model.Playoff)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	matchesByType := map[model.MatchType]MatchPlayList{
		model.Practice:      practiceMatches,
		model.Qualification: qualificationMatches,
		model.Playoff:       playoffMatches,
	}
	currentMatchType := web.arena.CurrentMatch.Type
	if currentMatchType == model.Test {
		currentMatchType = model.Practice
	}

	template, err := web.parseFiles("templates/match_play_match_load.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		MatchesByType    map[model.MatchType]MatchPlayList
		CurrentMatchType model.MatchType
	}{
		matchesByType,
		currentMatchType,
	}
	err = template.ExecuteTemplate(w, "match_play_match_load.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
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
	go ws.HandleNotifiers(
		web.arena.MatchTimingNotifier,
		web.arena.AllianceStationDisplayModeNotifier,
		web.arena.ArenaStatusNotifier,
		web.arena.AudienceDisplayModeNotifier,
		web.arena.EventStatusNotifier,
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ScorePostedNotifier,
		web.arena.ScoringStatusNotifier,
	)

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
		case "loadMatch":
			args := struct {
				MatchId int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if args.MatchId == 0 {
				err = web.arena.LoadTestMatch()
			} else {
				match, err := web.arena.Database.GetMatchById(args.MatchId)
				if err != nil {
					ws.WriteError(err.Error())
					continue
				}
				if match == nil {
					ws.WriteError(fmt.Sprintf("invalid match ID %d", args.MatchId))
					continue
				}
				err = web.arena.LoadMatch(match)
			}
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "showResult":
			args := struct {
				MatchId int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if args.MatchId == 0 {
				// Load an empty match to effectively clear the buffer.
				web.arena.SavedMatch = &model.Match{}
				web.arena.SavedMatchResult = model.NewMatchResult()
				web.arena.ScorePostedNotifier.Notify()
				continue
			}
			match, err := web.arena.Database.GetMatchById(args.MatchId)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if match == nil {
				ws.WriteError(fmt.Sprintf("invalid match ID %d", args.MatchId))
				continue
			}
			matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if matchResult == nil {
				ws.WriteError(fmt.Sprintf("No result found for match ID %d.", args.MatchId))
				continue
			}
			if match.ShouldUpdateRankings() {
				web.arena.SavedRankings, err = web.arena.Database.GetAllRankings()
				if err != nil {
					ws.WriteError(err.Error())
					continue
				}
			} else {
				web.arena.SavedRankings = game.Rankings{}
			}
			web.arena.SavedMatch = match
			web.arena.SavedMatchResult = matchResult
			web.arena.ScorePostedNotifier.Notify()
		case "substituteTeams":
			args := struct {
				Red1  int
				Red2  int
				Red3  int
				Blue1 int
				Blue2 int
				Blue3 int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.SubstituteTeams(args.Red1, args.Red2, args.Red3, args.Blue1, args.Blue2, args.Blue3)
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
			if err = ws.WriteNotifier(web.arena.ArenaStatusNotifier); err != nil {
				log.Println(err)
			}
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
		case "signalReset":
			if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
		case "commitResults":
			if web.arena.MatchState != field.PostMatch {
				ws.WriteError("cannot commit match while it is in progress")
				continue
			}
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
			err = web.arena.LoadNextMatch(true)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "discardResults":
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch(false)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "setAudienceDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAudienceDisplayMode(mode)
		case "setAllianceStationDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAllianceStationDisplayMode(mode)
		case "startTimeout":
			durationSec, ok := data.(float64)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			err = web.arena.StartTimeout("Timeout", int(durationSec))
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "setTestMatchName":
			if web.arena.CurrentMatch.Type != model.Test {
				// Don't allow changing the name of a non-test match.
				continue
			}
			name, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.CurrentMatch.LongName = name
			web.arena.MatchLoadNotifier.Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}

// Saves the given match and result to the database, supplanting any previous result for the match.
func (web *Web) commitMatchScore(match *model.Match, matchResult *model.MatchResult, isMatchReviewEdit bool) error {
	var updatedRankings game.Rankings

	if match.Type == model.Playoff {
		// Adjust the score if necessary for a playoff DQ.
		matchResult.CorrectPlayoffScore()
	}

	// Update the match record.
	match.ScoreCommittedAt = time.Now()
	redScoreSummary := matchResult.RedScoreSummary()
	blueScoreSummary := matchResult.BlueScoreSummary()
	match.Status = game.DetermineMatchStatus(redScoreSummary, blueScoreSummary, match.UseTiebreakCriteria)

	if match.Type != model.Test {
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

		if match.ShouldUpdatePlayoffMatches() {
			if err = web.arena.Database.UpdateAllianceFromMatch(
				match.PlayoffRedAlliance, [3]int{match.Red1, match.Red2, match.Red3},
			); err != nil {
				return err
			}
			if err = web.arena.Database.UpdateAllianceFromMatch(
				match.PlayoffBlueAlliance, [3]int{match.Blue1, match.Blue2, match.Blue3},
			); err != nil {
				return err
			}

			// Populate any subsequent playoff matches.
			if err = web.arena.UpdatePlayoffTournament(); err != nil {
				return err
			}

			// Generate awards if the tournament is over.
			if web.arena.PlayoffTournament.IsComplete() {
				winnerAllianceId := web.arena.PlayoffTournament.WinningAllianceId()
				finalistAllianceId := web.arena.PlayoffTournament.FinalistAllianceId()
				if err = tournament.CreateOrUpdateWinnerAndFinalistAwards(
					web.arena.Database, winnerAllianceId, finalistAllianceId,
				); err != nil {
					return err
				}
			}
		}

		if web.arena.EventSettings.TbaPublishingEnabled && match.Type != model.Practice {
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
			fmt.Sprintf("post_%s_match_%s", match.Type, match.ShortName))
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
	return list[i].Status == game.MatchScheduled && list[j].Status != game.MatchScheduled
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Constructs the list of matches to display on the side of the match play interface.
func (web *Web) buildMatchPlayList(matchType model.MatchType) (MatchPlayList, error) {
	matches, err := web.arena.Database.GetMatchesByType(matchType, false)
	if err != nil {
		return MatchPlayList{}, err
	}

	matchPlayList := make(MatchPlayList, len(matches))
	for i, match := range matches {
		matchPlayList[i].Id = match.Id
		matchPlayList[i].ShortName = match.ShortName
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
