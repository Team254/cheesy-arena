// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package main

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the scoring interface which enables input of scores in real-time.
func ScoringDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := template.ParseFiles("templates/scoring_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		Alliance string
	}{eventSettings, alliance}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func ScoringDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}
	var score **RealtimeScore
	var scoreSummaryFunc func() *game.ScoreSummary
	if alliance == "red" {
		score = &mainArena.redRealtimeScore
		scoreSummaryFunc = mainArena.RedScoreSummary
	} else {
		score = &mainArena.blueRealtimeScore
		scoreSummaryFunc = mainArena.BlueScoreSummary
	}
	autoCommitted := false

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	data := struct {
		Score         *RealtimeScore
		ScoreSummary  *game.ScoreSummary
		AutoCommitted bool
	}{*score, scoreSummaryFunc(), autoCommitted}
	err = websocket.Write("score", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
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
				message = MatchTimeMessage{mainArena.MatchState, matchTimeSec.(int)}
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
		case "mobility":
			if !autoCommitted {
				if (*score).CurrentScore.AutoMobility < 3 {
					(*score).CurrentScore.AutoMobility++
				}
			}
		case "undoMobility":
			if !autoCommitted {
				if (*score).CurrentScore.AutoMobility > 0 {
					(*score).CurrentScore.AutoMobility--
				}
			}
		case "commit":
			if mainArena.MatchState != preMatch || mainArena.currentMatch.Type == "test" {
				autoCommitted = true
			}
		case "uncommitAuto":
			autoCommitted = false
		case "commitMatch":
			if mainArena.MatchState != postMatch {
				// Don't allow committing the score until the match is over.
				websocket.WriteError("Cannot commit score: Match is not over.")
				continue
			}

			autoCommitted = true
			(*score).TeleopCommitted = true
			mainArena.scoringStatusNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		mainArena.realtimeScoreNotifier.Notify(nil)

		// Send out the score again after handling the command, as it most likely changed as a result.
		data = struct {
			Score         *RealtimeScore
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
