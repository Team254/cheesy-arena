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
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

type ScoringPosition struct {
	Alliance         string
	NearSide         bool
	ScoresAuto       bool
	ScoresEndgame    bool
	ScoresBarge      bool
	ScoresProcessor  bool
	LeftmostReefPole int
}

var positionParameters = map[string]ScoringPosition{
	"red_near": {
		Alliance:         "red",
		NearSide:         true,
		ScoresAuto:       true,
		ScoresEndgame:    true,
		ScoresBarge:      true,
		ScoresProcessor:  false,
		LeftmostReefPole: 6,
	},
	"red_far": {
		Alliance:         "red",
		NearSide:         false,
		ScoresAuto:       false,
		ScoresEndgame:    false,
		ScoresBarge:      false,
		ScoresProcessor:  true,
		LeftmostReefPole: 0,
	},
	"blue_near": {
		Alliance:         "blue",
		NearSide:         true,
		ScoresAuto:       false,
		ScoresEndgame:    false,
		ScoresBarge:      false,
		ScoresProcessor:  true,
		LeftmostReefPole: 0,
	},
	"blue_far": {
		Alliance:         "blue",
		NearSide:         false,
		ScoresAuto:       true,
		ScoresEndgame:    true,
		ScoresBarge:      true,
		ScoresProcessor:  false,
		LeftmostReefPole: 6,
	},
}

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	parameters, ok := positionParameters[position]
	if !ok {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
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
		PositionName string
		Position     ScoringPosition
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), position, parameters}
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

	alliance := r.PathValue("alliance")
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
		command, data, err := ws.Read()
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
		} else if command == "reef" {
			args := struct {
				ReefPosition int
				ReefLevel    int
				Current      bool
				Autonomous   bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.ReefPosition >= 1 && args.ReefPosition <= 12 && args.ReefLevel >= 2 && args.ReefLevel <= 4 {
				level := game.Level(args.ReefLevel - 2)
				reefIndex := args.ReefPosition - 1
				if args.Current {
					score.Reef.Branches[level][reefIndex] = !score.Reef.Branches[level][reefIndex]
					scoreChanged = true
				}
				if args.Autonomous {
					score.Reef.AutoBranches[level][reefIndex] = !score.Reef.AutoBranches[level][reefIndex]
					scoreChanged = true
				}
				scoreChanged = true
			}

		} else if command == "endgame" {
			args := struct {
				TeamPosition  int
				EndgameStatus int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.TeamPosition >= 1 && args.TeamPosition <= 3 && args.EndgameStatus >= 0 && args.EndgameStatus <= 3 {
				endgameStatus := game.EndgameStatus(args.EndgameStatus)
				score.EndgameStatuses[args.TeamPosition-1] = endgameStatus
				scoreChanged = true
			}
		} else if command == "leave" {
			args := struct {
				TeamPosition int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
				score.LeaveStatuses[args.TeamPosition-1] = !score.LeaveStatuses[args.TeamPosition-1]
				scoreChanged = true
			}
		} else {
			args := struct {
				Adjustment int
				Current    bool
				Autonomous bool
				NearSide   bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			switch command {
			case "barge":
				score.BargeAlgae = max(0, score.BargeAlgae+args.Adjustment)
				scoreChanged = true
			case "processor":
				score.ProcessorAlgae = max(0, score.ProcessorAlgae+args.Adjustment)
				scoreChanged = true
			case "trough":
				if args.Current {
					if args.NearSide {
						score.Reef.TroughNear = max(0, score.Reef.TroughNear+args.Adjustment)
					} else {
						score.Reef.TroughFar = max(0, score.Reef.TroughFar+args.Adjustment)
					}
					scoreChanged = true
				}
				if args.Autonomous {
					if args.NearSide {
						score.Reef.AutoTroughNear = max(0, score.Reef.AutoTroughNear+args.Adjustment)
					} else {
						score.Reef.AutoTroughFar = max(0, score.Reef.AutoTroughFar+args.Adjustment)
					}
					scoreChanged = true
				}
			}
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}
