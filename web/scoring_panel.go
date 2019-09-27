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
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
)

// Maps a numbered bay on the scoring panel to the field that it represents in the Score model.
type bayMapping struct {
	BayId       int
	Shortcut    string
	RedElement  string
	RedIndex    int
	BlueElement string
	BlueIndex   int
}

var bayMappings = []*bayMapping{
	{0, "q", "rocketNearRight", 2, "rocketFarRight", 2},
	{1, "a", "rocketNearRight", 1, "rocketFarRight", 1},
	{2, "z", "rocketNearRight", 0, "rocketFarRight", 0},
	{3, "w", "rocketNearLeft", 2, "rocketFarLeft", 2},
	{4, "s", "rocketNearLeft", 1, "rocketFarLeft", 1},
	{5, "x", "rocketNearLeft", 0, "rocketFarLeft", 0},
	{6, "e", "cargoShip", 0, "cargoShip", 7},
	{7, "d", "cargoShip", 1, "cargoShip", 6},
	{8, "c", "cargoShip", 2, "cargoShip", 5},
	{9, "v", "cargoShip", 3, "cargoShip", 4},
	{10, "b", "cargoShip", 4, "cargoShip", 3},
	{11, "n", "cargoShip", 5, "cargoShip", 2},
	{12, "j", "cargoShip", 6, "cargoShip", 1},
	{13, "i", "cargoShip", 7, "cargoShip", 0},
	{14, "o", "rocketFarRight", 2, "rocketNearRight", 2},
	{15, "k", "rocketFarRight", 1, "rocketNearRight", 1},
	{16, "m", "rocketFarRight", 0, "rocketNearRight", 0},
	{17, "p", "rocketFarLeft", 2, "rocketNearLeft", 2},
	{18, "l", "rocketFarLeft", 1, "rocketNearLeft", 1},
	{19, ",", "rocketFarLeft", 0, "rocketNearLeft", 0},
}

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
		Alliance    string
		BayMappings []*bayMapping
	}{web.arena.EventSettings, alliance, bayMappings}
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
		} else if number, err := strconv.Atoi(command); err == nil && number >= 1 && number <= 9 {
			// Handle per-robot scoring fields.
			if number <= 3 && web.arena.MatchState == field.PreMatch {
				index := number - 1
				score.RobotStartLevels[index]++
				if score.RobotStartLevels[index] == 4 {
					score.RobotStartLevels[index] = 0
				}
				scoreChanged = true
			} else if number > 3 && number <= 6 && web.arena.MatchState != field.PreMatch {
				index := number - 4
				score.SandstormBonuses[index] =
					!score.SandstormBonuses[index]
				scoreChanged = true
			} else if number > 6 && web.arena.MatchState != field.PreMatch {
				index := number - 7
				score.RobotEndLevels[index]++
				if score.RobotEndLevels[index] == 4 {
					score.RobotEndLevels[index] = 0
				}
				scoreChanged = true
			}
		} else {
			// Handle cargo bays.
			var bayMapping *bayMapping
			for _, mapping := range bayMappings {
				if mapping.Shortcut == command {
					bayMapping = mapping
					break
				}
			}
			if bayMapping != nil {
				element := bayMapping.RedElement
				index := bayMapping.RedIndex
				if alliance == "blue" {
					element = bayMapping.BlueElement
					index = bayMapping.BlueIndex
				}
				switch element {
				case "cargoShip":
					scoreChanged = web.toggleCargoShipBay(&score.CargoBays[index], index)
				case "rocketNearLeft":
					scoreChanged = web.toggleRocketBay(&score.RocketNearLeftBays[index])
				case "rocketNearRight":
					scoreChanged = web.toggleRocketBay(&score.RocketNearRightBays[index])
				case "rocketFarLeft":
					scoreChanged = web.toggleRocketBay(&score.RocketFarLeftBays[index])
				case "rocketFarRight":
					scoreChanged = web.toggleRocketBay(&score.RocketFarRightBays[index])
				}
			}
		}

		if scoreChanged {
			if web.arena.MatchState == field.PreMatch {
				score.CargoBaysPreMatch = score.CargoBays
			}
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Advances the given cargo ship bay through the states applicable to the current status of the field.
func (web *Web) toggleCargoShipBay(bay *game.BayStatus, index int) bool {
	if (index == 3 || index == 4) && web.arena.MatchState == field.PreMatch {
		// Only the side bays can be preloaded.
		return false
	}

	if web.arena.MatchState == field.PreMatch {
		*bay++
		if *bay == game.BayHatchCargo {
			// Skip the hatch+cargo state pre-match as it is invalid.
			*bay = game.BayCargo
		} else if *bay > game.BayCargo {
			*bay = game.BayEmpty
		}
	} else {
		if *bay == game.BayCargo {
			// If the bay was pre-loaded with cargo, go immediately to hatch+cargo during first toggle.
			*bay = game.BayHatchCargo
		} else {
			*bay++
			if *bay == game.BayCargo {
				// Skip the cargo-only state during the match as it can't stay in on its own.
				*bay = game.BayEmpty
			}
		}
	}
	return true
}

// Advances the given rocket bay through the states applicable to the current status of the field.
func (web *Web) toggleRocketBay(bay *game.BayStatus) bool {
	if web.arena.MatchState != field.PreMatch {
		*bay++
		if *bay == game.BayCargo {
			// Skip the cargo-only state as it's not applicable to rocket bays.
			*bay = game.BayEmpty
		}
		return true
	}
	return false
}
