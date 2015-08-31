// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the referee interface.

package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var rules = []string{"G4", "G5", "G6", "G6-1", "G16", "G17", "G18", "G19", "G20", "G21", "G22", "G23", "G24",
	"G26", "G30", "G31", "G33", "G34"}

// Renders the referee interface for assigning fouls.
func RefereeDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/referee_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	match := mainArena.currentMatch
	matchType := match.CapitalizedType()
	red1 := mainArena.AllianceStations["R1"].team
	if red1 == nil {
		red1 = &Team{}
	}
	red2 := mainArena.AllianceStations["R2"].team
	if red2 == nil {
		red2 = &Team{}
	}
	red3 := mainArena.AllianceStations["R3"].team
	if red3 == nil {
		red3 = &Team{}
	}
	blue1 := mainArena.AllianceStations["B1"].team
	if blue1 == nil {
		blue1 = &Team{}
	}
	blue2 := mainArena.AllianceStations["B2"].team
	if blue2 == nil {
		blue2 = &Team{}
	}
	blue3 := mainArena.AllianceStations["B3"].team
	if blue3 == nil {
		blue3 = &Team{}
	}
	data := struct {
		*EventSettings
		MatchType        string
		MatchDisplayName string
		Red1             *Team
		Red2             *Team
		Red3             *Team
		Blue1            *Team
		Blue2            *Team
		Blue3            *Team
		RedFouls         []Foul
		BlueFouls        []Foul
		RedCards         map[string]string
		BlueCards        map[string]string
		Rules            []string
		EntryEnabled     bool
	}{eventSettings, matchType, match.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		mainArena.redRealtimeScore.CurrentScore.Fouls, mainArena.blueRealtimeScore.CurrentScore.Fouls,
		mainArena.redRealtimeScore.Cards, mainArena.blueRealtimeScore.Cards, rules,
		!(mainArena.redRealtimeScore.FoulsCommitted && mainArena.blueRealtimeScore.FoulsCommitted)}
	err = template.ExecuteTemplate(w, "referee_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the refereee interface client to send control commands and receive status updates.
func RefereeDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
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
		case "addFoul":
			args := struct {
				Alliance string
				TeamId   int
				Rule     string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := Foul{TeamId: args.TeamId, Rule: args.Rule, TimeInMatchSec: mainArena.MatchTimeSec()}
			if args.Alliance == "red" {
				mainArena.redRealtimeScore.CurrentScore.Fouls =
					append(mainArena.redRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				mainArena.blueRealtimeScore.CurrentScore.Fouls =
					append(mainArena.blueRealtimeScore.CurrentScore.Fouls, foul)
			}
			mainArena.realtimeScoreNotifier.Notify(nil)
		case "deleteFoul":
			args := struct {
				Alliance       string
				TeamId         int
				Rule           string
				TimeInMatchSec float64
				IsTechnical    bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Remove the foul from the correct alliance's list.
			deleteFoul := Foul{TeamId: args.TeamId, Rule: args.Rule, TimeInMatchSec: args.TimeInMatchSec}
			var fouls *[]Foul
			if args.Alliance == "red" {
				fouls = &mainArena.redRealtimeScore.CurrentScore.Fouls
			} else {
				fouls = &mainArena.blueRealtimeScore.CurrentScore.Fouls
			}
			for i, foul := range *fouls {
				if foul == deleteFoul {
					*fouls = append((*fouls)[:i], (*fouls)[i+1:]...)
					break
				}
			}
			mainArena.realtimeScoreNotifier.Notify(nil)
		case "card":
			args := struct {
				Alliance string
				TeamId   int
				Card     string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Set the card in the correct alliance's score.
			var cards map[string]string
			if args.Alliance == "red" {
				cards = mainArena.redRealtimeScore.Cards
			} else {
				cards = mainArena.blueRealtimeScore.Cards
			}
			cards[strconv.Itoa(args.TeamId)] = args.Card
			continue
		case "signalReset":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow clearing the field until the match is over.
				continue
			}
			mainArena.allianceStationDisplayScreen = "fieldReset"
			mainArena.allianceStationDisplayNotifier.Notify(nil)
			continue // Don't reload.
		case "commitMatch":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow committing the fouls until the match is over.
				continue
			}
			mainArena.redRealtimeScore.FoulsCommitted = true
			mainArena.blueRealtimeScore.FoulsCommitted = true
			mainArena.allianceStationDisplayScreen = "fieldReset"
			mainArena.allianceStationDisplayNotifier.Notify(nil)
			mainArena.scoringStatusNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Force a reload of the client to render the updated foul list.
		err = websocket.Write("reload", nil)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}
