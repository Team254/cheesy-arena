// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for controlling match play.

package main

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"text/template"
	"time"
)

type MatchPlayListItem struct {
	Id          int
	DisplayName string
	Time        string
	Status      string
	ColorClass  string
}

type MatchPlayList []MatchPlayListItem

type MatchTimeMessage struct {
	MatchState   int
	MatchTimeSec int
}

// Global var to hold the current active tournament so that its matches are displayed by default.
var currentMatchType string

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

	template := template.New("").Funcs(web.templateHelpers)
	_, err = template.ParseFiles("templates/match_play.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[string]MatchPlayList{"practice": practiceMatches,
		"qualification": qualificationMatches, "elimination": eliminationMatches}
	if currentMatchType == "" {
		currentMatchType = "practice"
	}
	allowSubstitution := web.arena.CurrentMatch.Type != "qualification"
	matchResult, err := web.arena.Database.GetMatchResultForMatch(web.arena.CurrentMatch.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	isReplay := matchResult != nil
	data := struct {
		*model.EventSettings
		MatchesByType     map[string]MatchPlayList
		CurrentMatchType  string
		Match             *model.Match
		AllowSubstitution bool
		IsReplay          bool
	}{web.arena.EventSettings, matchesByType, currentMatchType, web.arena.CurrentMatch, allowSubstitution, isReplay}
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
	currentMatchType = web.arena.CurrentMatch.Type

	http.Redirect(w, r, "/match_play", 302)
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
	web.arena.SavedMatch = match
	web.arena.SavedMatchResult = matchResult
	web.arena.ScorePostedNotifier.Notify(nil)

	http.Redirect(w, r, "/match_play", 302)
}

// The websocket endpoint for the match play client to send control commands and receive status updates.
func (web *Web) matchPlayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchTimeListener := web.arena.MatchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := web.arena.RealtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	robotStatusListener := web.arena.RobotStatusNotifier.Listen()
	defer close(robotStatusListener)
	audienceDisplayListener := web.arena.AudienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	scoringStatusListener := web.arena.ScoringStatusNotifier.Listen()
	defer close(scoringStatusListener)
	allianceStationDisplayListener := web.arena.AllianceStationDisplayNotifier.Listen()
	defer close(allianceStationDisplayListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	err = websocket.Write("status", web.arena.GetStatus())
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", game.MatchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = MatchTimeMessage{web.arena.MatchState, int(web.arena.LastMatchTimeSec)}
	err = websocket.Write("matchTime", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
	err = websocket.Write("realtimeScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("setAudienceDisplay", web.arena.AudienceDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RefereeScoreReady bool
		RedScoreReady     bool
		BlueScoreReady    bool
	}{web.arena.RedRealtimeScore.FoulsCommitted && web.arena.BlueRealtimeScore.FoulsCommitted,
		web.arena.RedRealtimeScore.TeleopCommitted, web.arena.BlueRealtimeScore.TeleopCommitted}
	err = websocket.Write("scoringStatus", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("setAllianceStationDisplay", web.arena.AllianceStationDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{web.arena.MatchState, matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					BlueScore int
				}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = web.arena.GetStatus()
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = web.arena.AudienceDisplayScreen
			case _, ok := <-scoringStatusListener:
				if !ok {
					return
				}
				messageType = "scoringStatus"
				message = struct {
					RefereeScoreReady bool
					RedScoreReady     bool
					BlueScoreReady    bool
				}{web.arena.RedRealtimeScore.FoulsCommitted && web.arena.BlueRealtimeScore.FoulsCommitted,
					web.arena.RedRealtimeScore.TeleopCommitted, web.arena.BlueRealtimeScore.TeleopCommitted}
			case _, ok := <-allianceStationDisplayListener:
				if !ok {
					return
				}
				messageType = "setAllianceStationDisplay"
				message = web.arena.AllianceStationDisplayScreen
			}
			err = websocket.Write(messageType, message)
			if err != nil {
				// The client has probably closed the connection; nothing to do here.
				return
			}
		}
	}()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
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
				websocket.WriteError(err.Error())
				continue
			}
			err = web.arena.SubstituteTeam(args.Team, args.Position)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "toggleBypass":
			station, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			if _, ok := web.arena.AllianceStations[station]; !ok {
				websocket.WriteError(fmt.Sprintf("Invalid alliance station '%s'.", station))
				continue
			}
			web.arena.AllianceStations[station].Bypass = !web.arena.AllianceStations[station].Bypass
		case "startMatch":
			args := struct {
				MuteMatchSounds bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			web.arena.MuteMatchSounds = args.MuteMatchSounds
			err = web.arena.StartMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "abortMatch":
			err = web.arena.AbortMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "commitResults":
			err = web.commitCurrentMatchScore()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = web.arena.ResetMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = websocket.Write("reload", nil)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "discardResults":
			err = web.arena.ResetMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = websocket.Write("reload", nil)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "setAudienceDisplay":
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.AudienceDisplayScreen = screen
			web.arena.AudienceDisplayNotifier.Notify(nil)
			continue
		case "setAllianceStationDisplay":
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.AllianceStationDisplayScreen = screen
			web.arena.AllianceStationDisplayNotifier.Notify(nil)
			continue
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Send out the status again after handling the command, as it most likely changed as a result.
		err = websocket.Write("status", web.arena)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

// Saves the given match and result to the database, supplanting any previous result for the match.
func (web *Web) commitMatchScore(match *model.Match, matchResult *model.MatchResult, loadToShowBuffer bool) error {
	if match.Type == "elimination" {
		// Adjust the score if necessary for an elimination DQ.
		matchResult.CorrectEliminationScore()
	}

	if loadToShowBuffer {
		// Store the result in the buffer to be shown in the audience display.
		web.arena.SavedMatch = match
		web.arena.SavedMatchResult = matchResult
		web.arena.ScorePostedNotifier.Notify(nil)
	}

	if match.Type == "test" {
		// Do nothing since this is a test match and doesn't exist in the database.
		return nil
	}

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
		err := web.arena.Database.SaveMatchResult(matchResult)
		if err != nil {
			return err
		}
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
	err := web.arena.Database.SaveMatch(match)
	if err != nil {
		return err
	}

	if match.Type != "practice" {
		// Regenerate the residual yellow cards that teams may carry.
		tournament.CalculateTeamCards(web.arena.Database, match.Type)
	}

	if match.Type == "qualification" {
		// Recalculate all the rankings.
		err = tournament.CalculateRankings(web.arena.Database)
		if err != nil {
			return err
		}
	}

	if match.Type == "elimination" {
		// Generate any subsequent elimination matches.
		_, err = tournament.UpdateEliminationSchedule(web.arena.Database, time.Now().Add(time.Second*tournament.ElimMatchSpacingSec))
		if err != nil {
			return err
		}
	}

	if web.arena.EventSettings.TbaPublishingEnabled && match.Type != "practice" {
		// Publish asynchronously to The Blue Alliance.
		go func() {
			err = web.arena.TbaClient.PublishMatches(web.arena.Database)
			if err != nil {
				log.Printf("Failed to publish matches: %s", err.Error())
			}
			if match.Type == "qualification" {
				err = web.arena.TbaClient.PublishRankings(web.arena.Database)
				if err != nil {
					log.Printf("Failed to publish rankings: %s", err.Error())
				}
			}
		}()
	}

	if web.arena.EventSettings.StemTvPublishingEnabled && match.Type != "practice" {
		// Publish asynchronously to STEMtv.
		go func() {
			err = web.arena.StemTvClient.PublishMatchVideoSplit(match, time.Now())
			if err != nil {
				log.Printf("Failed to publish match video split to STEMtv: %s", err.Error())
			}
		}()
	}

	// Back up the database, but don't error out if it fails.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, fmt.Sprintf("post_%s_match_%s", match.Type, match.DisplayName))
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (web *Web) getCurrentMatchResult() *model.MatchResult {
	return &model.MatchResult{MatchId: web.arena.CurrentMatch.Id, MatchType: web.arena.CurrentMatch.Type,
		RedScore: web.arena.RedRealtimeScore.CurrentScore, BlueScore: web.arena.BlueRealtimeScore.CurrentScore,
		RedCards: web.arena.RedRealtimeScore.Cards, BlueCards: web.arena.BlueRealtimeScore.Cards}
}

// Saves the realtime result as the final score for the match currently loaded into the arena.
func (web *Web) commitCurrentMatchScore() error {
	return web.commitMatchScore(web.arena.CurrentMatch, web.getCurrentMatchResult(), true)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Len() int {
	return len(list)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Less(i, j int) bool {
	return list[i].Status != "complete" && list[j].Status == "complete"
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
		matchPlayList[i].Time = match.Time.Local().Format("3:04 PM")
		matchPlayList[i].Status = match.Status
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
		if web.arena.CurrentMatch != nil && matchPlayList[i].Id == web.arena.CurrentMatch.Id {
			matchPlayList[i].ColorClass = "success"
		}
	}

	// Sort the list to put all completed matches at the bottom.
	sort.Stable(matchPlayList)

	return matchPlayList, nil
}
