// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
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
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	err = websocket.Write("score", *score)
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
		case "preload":
			if !(*score).AutoCommitted {
				preloadedBallsStr, ok := data.(string)
				if !ok {
					websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
					continue
				}
				preloadedBalls, err := strconv.Atoi(preloadedBallsStr)
				(*score).AutoPreloadedBalls = preloadedBalls
				if err != nil {
					websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
					continue
				}
			}
		case "mobility":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoMobilityBonuses += 1
			}
		case "scoredHighHot":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoHighHot += 1
			}
		case "scoredHigh":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoHigh += 1
			} else if !(*score).TeleopCommitted && !(*score).CurrentCycle.ScoredHigh {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = true
				(*score).CurrentCycle.ScoredLow = false
				(*score).CurrentCycle.DeadBall = false
			}
		case "scoredLowHot":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoLowHot += 1
			}
		case "scoredLow":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoLow += 1
			} else if !(*score).TeleopCommitted && !(*score).CurrentCycle.ScoredLow {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = false
				(*score).CurrentCycle.ScoredLow = true
				(*score).CurrentCycle.DeadBall = false
			}
		case "assist":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				(*score).CurrentCycle.Assists < 3 {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Assists += 1
			}
		case "truss":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				!(*score).CurrentCycle.Truss {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Truss = true
			}
		case "catch":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				!(*score).CurrentCycle.Catch && (*score).CurrentCycle.Truss {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Catch = true
			}
		case "deadBall":
			if !(*score).TeleopCommitted && !(*score).CurrentCycle.DeadBall {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = false
				(*score).CurrentCycle.ScoredLow = false
				(*score).CurrentCycle.DeadBall = true
			}
		case "commit":
			if !(*score).AutoCommitted {
				(*score).AutoLeftoverBalls = (*score).AutoPreloadedBalls - (*score).CurrentScore.AutoHighHot -
					(*score).CurrentScore.AutoHigh - (*score).CurrentScore.AutoLowHot -
					(*score).CurrentScore.AutoLow
				(*score).AutoCommitted = true
			} else if !(*score).TeleopCommitted {
				if (*score).CurrentCycle.ScoredHigh || (*score).CurrentCycle.ScoredLow ||
					(*score).CurrentCycle.DeadBall {
					// Check whether this is a leftover ball from autonomous.
					if (*score).AutoLeftoverBalls > 0 {
						if (*score).CurrentCycle.ScoredHigh {
							(*score).CurrentScore.AutoClearHigh += 1
						} else if (*score).CurrentCycle.ScoredLow {
							(*score).CurrentScore.AutoClearLow += 1
						} else {
							(*score).CurrentScore.AutoClearDead += 1
						}
						(*score).AutoLeftoverBalls -= 1
					} else {
						(*score).CurrentScore.Cycles = append((*score).CurrentScore.Cycles, (*score).CurrentCycle)
					}
					(*score).CurrentCycle = Cycle{}
					(*score).undoCycles = []Cycle{}
				}
			}
		case "commitMatch":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow committing the score until the match is over.
				continue
			}
			(*score).AutoCommitted = true
			(*score).TeleopCommitted = true
			if (*score).CurrentCycle != (Cycle{}) {
				// Commit last cycle.
				(*score).CurrentScore.Cycles = append((*score).CurrentScore.Cycles, (*score).CurrentCycle)
			}
			mainArena.scoringStatusNotifier.Notify(nil)
		case "undo":
			if !(*score).AutoCommitted && len((*score).undoAutoScores) > 0 {
				(*score).CurrentScore = (*score).undoAutoScores[len((*score).undoAutoScores)-1]
				(*score).undoAutoScores = (*score).undoAutoScores[0 : len((*score).undoAutoScores)-1]
			} else if !(*score).TeleopCommitted && len((*score).undoCycles) > 0 {
				(*score).CurrentCycle = (*score).undoCycles[len((*score).undoCycles)-1]
				(*score).undoCycles = (*score).undoCycles[0 : len((*score).undoCycles)-1]
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
