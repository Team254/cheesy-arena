// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the referee interface.

package main

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

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
	red1 := mainArena.AllianceStations["R1"].Team
	if red1 == nil {
		red1 = &model.Team{}
	}
	red2 := mainArena.AllianceStations["R2"].Team
	if red2 == nil {
		red2 = &model.Team{}
	}
	red3 := mainArena.AllianceStations["R3"].Team
	if red3 == nil {
		red3 = &model.Team{}
	}
	blue1 := mainArena.AllianceStations["B1"].Team
	if blue1 == nil {
		blue1 = &model.Team{}
	}
	blue2 := mainArena.AllianceStations["B2"].Team
	if blue2 == nil {
		blue2 = &model.Team{}
	}
	blue3 := mainArena.AllianceStations["B3"].Team
	if blue3 == nil {
		blue3 = &model.Team{}
	}
	data := struct {
		*model.EventSettings
		MatchType        string
		MatchDisplayName string
		Red1             *model.Team
		Red2             *model.Team
		Red3             *model.Team
		Blue1            *model.Team
		Blue2            *model.Team
		Blue3            *model.Team
		RedFouls         []game.Foul
		BlueFouls        []game.Foul
		RedCards         map[string]string
		BlueCards        map[string]string
		Rules            []game.Rule
		EntryEnabled     bool
	}{eventSettings, matchType, match.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		mainArena.redRealtimeScore.CurrentScore.Fouls, mainArena.blueRealtimeScore.CurrentScore.Fouls,
		mainArena.redRealtimeScore.Cards, mainArena.blueRealtimeScore.Cards, game.Rules,
		!(mainArena.redRealtimeScore.FoulsCommitted && mainArena.blueRealtimeScore.FoulsCommitted)}
	err = template.ExecuteTemplate(w, "referee_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the refereee interface client to send control commands and receive status updates.
func RefereeDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(patrick): Enable authentication once Safari (for iPad) supports it over Websocket.

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
				Alliance    string
				TeamId      int
				Rule        string
				IsTechnical bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := game.Foul{Rule: game.Rule{RuleNumber: args.Rule, IsTechnical: args.IsTechnical},
				TeamId: args.TeamId, TimeInMatchSec: mainArena.MatchTimeSec()}
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
				IsTechnical    bool
				TimeInMatchSec float64
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Remove the foul from the correct alliance's list.
			deleteFoul := game.Foul{Rule: game.Rule{RuleNumber: args.Rule, IsTechnical: args.IsTechnical},
				TeamId: args.TeamId, TimeInMatchSec: args.TimeInMatchSec}
			var fouls *[]game.Foul
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
			if mainArena.MatchState != postMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			mainArena.fieldReset = true
			mainArena.allianceStationDisplayScreen = "fieldReset"
			mainArena.allianceStationDisplayNotifier.Notify(nil)
			continue // Don't reload.
		case "commitMatch":
			if mainArena.MatchState != postMatch {
				// Don't allow committing the fouls until the match is over.
				continue
			}
			mainArena.redRealtimeScore.FoulsCommitted = true
			mainArena.blueRealtimeScore.FoulsCommitted = true
			mainArena.fieldReset = true
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
