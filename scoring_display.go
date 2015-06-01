// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the scoring interface which enables input of scores in real-time.
func ScoringDisplayHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}
	var score **RealtimeScore
	if alliance == "red" {
		score = &mainArena.redRealtimeScore
	} else {
		score = &mainArena.blueRealtimeScore
	}

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
	err = websocket.Write("score", *score)
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
				messageType = "score"
				message = *score
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
		case "robotSet":
			if !(*score).AutoCommitted {
				(*score).CurrentScore.AutoRobotSet = !(*score).CurrentScore.AutoRobotSet
			}
		case "containerSet":
			if !(*score).AutoCommitted {
				(*score).CurrentScore.AutoContainerSet = !(*score).CurrentScore.AutoContainerSet
			}
		case "toteSet":
			if !(*score).AutoCommitted {
				(*score).CurrentScore.AutoToteSet = !(*score).CurrentScore.AutoToteSet
				if (*score).CurrentScore.AutoToteSet {
					(*score).CurrentScore.AutoStackedToteSet = false
				}
			} else {
				(*score).CurrentScore.CoopertitionSet = !(*score).CurrentScore.CoopertitionSet
				if (*score).CurrentScore.CoopertitionSet {
					(*score).CurrentScore.CoopertitionStack = false
				}
			}
		case "stackedToteSet":
			if !(*score).AutoCommitted {
				(*score).CurrentScore.AutoStackedToteSet = !(*score).CurrentScore.AutoStackedToteSet
				if (*score).CurrentScore.AutoStackedToteSet {
					(*score).CurrentScore.AutoToteSet = false
				}
			} else {
				(*score).CurrentScore.CoopertitionStack = !(*score).CurrentScore.CoopertitionStack
				if (*score).CurrentScore.CoopertitionStack {
					(*score).CurrentScore.CoopertitionSet = false
				}
			}
		case "commit":
			if !(*score).AutoCommitted {
				(*score).AutoCommitted = true
			} else {
				var stacks []Stack
				err = mapstructure.Decode(data, &stacks)
				if err != nil {
					websocket.WriteError(err.Error())
				}
				(*score).CurrentScore.Stacks = stacks
			}
		case "uncommitAuto":
			if (*score).AutoCommitted {
				(*score).AutoCommitted = false
			}
		case "commitMatch":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow committing the score until the match is over.
				websocket.WriteError("Cannot commit score: Match is not over.")
				continue
			}

			redScore := mainArena.redRealtimeScore.CurrentScore
			blueScore := mainArena.blueRealtimeScore.CurrentScore
			if redScore.CoopertitionSet != blueScore.CoopertitionSet ||
				redScore.CoopertitionStack != blueScore.CoopertitionStack {
				// Don't accept the score if the red and blue co-opertition points don't match up.
				websocket.ShowDialog("Cannot commit score: Red and blue co-opertition points do not match.")
				continue
			}

			(*score).AutoCommitted = true
			(*score).TeleopCommitted = true
			mainArena.scoringStatusNotifier.Notify(nil)
		case "undo":
			if len((*score).undoScores) > 0 {
				(*score).CurrentScore = (*score).undoScores[len((*score).undoScores)-1]
				(*score).undoScores = (*score).undoScores[0 : len((*score).undoScores)-1]
			}
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		mainArena.realtimeScoreNotifier.Notify(nil)

		// Send out the score again after handling the command, as it most likely changed as a result.
		err = websocket.Write("score", *score)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}
