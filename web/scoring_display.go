// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		Alliance string
	}{web.arena.EventSettings, alliance}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}
	var score **field.RealtimeScore
	var scoreSummaryFunc func() *game.ScoreSummary
	if alliance == "red" {
		score = &web.arena.RedRealtimeScore
		scoreSummaryFunc = web.arena.RedScoreSummary
	} else {
		score = &web.arena.BlueRealtimeScore
		scoreSummaryFunc = web.arena.BlueScoreSummary
	}
	autoCommitted := false

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := web.arena.MatchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := web.arena.MatchTimeNotifier.Listen()
	defer close(matchTimeListener)
	reloadDisplaysListener := web.arena.ReloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	data := struct {
		Score         *field.RealtimeScore
		ScoreSummary  *game.ScoreSummary
		AutoCommitted bool
	}{*score, scoreSummaryFunc(), autoCommitted}
	err = websocket.Write("score", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{int(web.arena.MatchState), int(web.arena.LastMatchTimeSec)})
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
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{int(web.arena.MatchState), matchTimeSec.(int)}
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		case "autoRun":
			if !autoCommitted {
				if (*score).CurrentScore.AutoRuns < 3 {
					(*score).CurrentScore.AutoRuns++
				}
			}
		case "undoAutoRun":
			if !autoCommitted {
				if (*score).CurrentScore.AutoRuns > 0 {
					(*score).CurrentScore.AutoRuns--
				}
			}
		case "climb":
			if autoCommitted {
				if (*score).CurrentScore.Climbs < 3 {
					(*score).CurrentScore.Climbs++
				}
			}
		case "undoClimb":
			if autoCommitted {
				if (*score).CurrentScore.Climbs > 0 {
					(*score).CurrentScore.Climbs--
				}
			}
		case "commit":
			if web.arena.MatchState != field.PreMatch || web.arena.CurrentMatch.Type == "test" {
				autoCommitted = true
			}
		case "uncommitAuto":
			autoCommitted = false
		case "commitMatch":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				websocket.WriteError("Cannot commit score: Match is not over.")
				continue
			}

			autoCommitted = true
			(*score).TeleopCommitted = true
			web.arena.ScoringStatusNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		web.arena.RealtimeScoreNotifier.Notify(nil)

		// Send out the score again after handling the command, as it most likely changed as a result.
		data = struct {
			Score         *field.RealtimeScore
			ScoreSummary  *game.ScoreSummary
			AutoCommitted bool
		}{*score, scoreSummaryFunc(), autoCommitted}
		err = websocket.Write("score", data)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}
