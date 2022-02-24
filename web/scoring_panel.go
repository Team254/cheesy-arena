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
	"strconv"
	"strings"
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
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
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

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if number, err := strconv.Atoi(command); err == nil && number >= 1 && number <= 6 {
			// Handle per-robot scoring fields.
			if number <= 3 {
				index := number - 1
				score.TaxiStatuses[index] = !score.TaxiStatuses[index]
				scoreChanged = true
			} else {
				index := number - 4
				score.EndgameStatuses[index]++
				if score.EndgameStatuses[index] == 5 {
					score.EndgameStatuses[index] = 0
				}
				scoreChanged = true
			}
		} else if !web.arena.Plc.IsEnabled() {
			switch strings.ToUpper(command) {
			case "Q":
				scoreChanged = decrementGoal(score.AutoCargoUpper[:])
			case "A":
				scoreChanged = decrementGoal(score.AutoCargoLower[:])
			case "W":
				scoreChanged = incrementGoal(score.AutoCargoUpper[:])
			case "S":
				scoreChanged = incrementGoal(score.AutoCargoLower[:])
			case "E":
				scoreChanged = decrementGoal(score.TeleopCargoUpper[:])
			case "D":
				scoreChanged = decrementGoal(score.TeleopCargoLower[:])
			case "R":
				scoreChanged = incrementGoal(score.TeleopCargoUpper[:])
			case "F":
				scoreChanged = incrementGoal(score.TeleopCargoLower[:])
			}

		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Increments the cargo count for the given goal.
func incrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	goal[0]++
	return true
}

// Decrements the cargo for the given goal.
func decrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	if goal[0] > 0 {
		goal[0]--
		return true
	}
	return false
}
