// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		Alliance string
	}{web.arena.EventSettings, alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	if alliance == "red" {
		score = &web.arena.RedRealtimeScore
	} else {
		score = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		scoreChanged := false
		switch messageType {
		case "r":
			if !(*score).AutoCommitted {
				if (*score).CurrentScore.AutoRuns < 3 {
					(*score).CurrentScore.AutoRuns++
					scoreChanged = true
				}
			}
		case "R":
			if !(*score).AutoCommitted {
				if (*score).CurrentScore.AutoRuns > 0 {
					(*score).CurrentScore.AutoRuns--
					scoreChanged = true
				}
			}
		case "c":
			if (*score).AutoCommitted {
				if (*score).CurrentScore.Climbs+(*score).CurrentScore.Parks < 3 {
					(*score).CurrentScore.Climbs++
					scoreChanged = true
				}
			}
		case "C":
			if (*score).AutoCommitted {
				if (*score).CurrentScore.Climbs > 0 {
					(*score).CurrentScore.Climbs--
					scoreChanged = true
				}
			}
		case "p":
			if (*score).AutoCommitted {
				if (*score).CurrentScore.Climbs+(*score).CurrentScore.Parks < 3 {
					(*score).CurrentScore.Parks++
					scoreChanged = true
				}
			}
		case "P":
			if (*score).AutoCommitted {
				if (*score).CurrentScore.Parks > 0 {
					(*score).CurrentScore.Parks--
					scoreChanged = true
				}
			}
		case "\r":
			if (web.arena.MatchState != field.PreMatch || web.arena.CurrentMatch.Type == "test") &&
				!(*score).AutoCommitted {
				(*score).AutoCommitted = true
				scoreChanged = true
			}
		case "a":
			if (*score).AutoCommitted {
				(*score).AutoCommitted = false
				scoreChanged = true
			}
		case "commitMatch":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}

			if !(*score).TeleopCommitted {
				(*score).AutoCommitted = true
				(*score).TeleopCommitted = true
				web.arena.ScoringStatusNotifier.Notify()
				scoreChanged = true
			}
		default:
			// Unknown keypress; just swallow the message without doing anything.
			continue
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}
