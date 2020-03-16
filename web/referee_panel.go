// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the referee interface.

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
	"strconv"
)

// Renders the referee interface for assigning fouls.
func (web *Web) refereePanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/referee_panel.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	match := web.arena.CurrentMatch
	matchType := match.CapitalizedType()
	red1 := web.arena.AllianceStations["R1"].Team
	if red1 == nil {
		red1 = &model.Team{}
	}
	red2 := web.arena.AllianceStations["R2"].Team
	if red2 == nil {
		red2 = &model.Team{}
	}
	red3 := web.arena.AllianceStations["R3"].Team
	if red3 == nil {
		red3 = &model.Team{}
	}
	blue1 := web.arena.AllianceStations["B1"].Team
	if blue1 == nil {
		blue1 = &model.Team{}
	}
	blue2 := web.arena.AllianceStations["B2"].Team
	if blue2 == nil {
		blue2 = &model.Team{}
	}
	blue3 := web.arena.AllianceStations["B3"].Team
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
		Rules            map[int]*game.Rule
		EntryEnabled     bool
	}{web.arena.EventSettings, matchType, match.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		web.arena.RedRealtimeScore.CurrentScore.Fouls, web.arena.BlueRealtimeScore.CurrentScore.Fouls,
		web.arena.RedRealtimeScore.Cards, web.arena.BlueRealtimeScore.Cards, game.GetAllRules(),
		!(web.arena.RedRealtimeScore.FoulsCommitted && web.arena.BlueRealtimeScore.FoulsCommitted)}
	err = template.ExecuteTemplate(w, "referee_panel.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the refereee interface client to send control commands and receive status updates.
func (web *Web) refereePanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "addFoul":
			args := struct {
				Alliance string
				TeamId   int
				RuleId   int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := game.Foul{RuleId: args.RuleId, TeamId: args.TeamId, TimeInMatchSec: web.arena.MatchTimeSec()}
			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			web.arena.RealtimeScoreNotifier.Notify()
		case "deleteFoul":
			args := struct {
				Alliance       string
				TeamId         int
				RuleId         int
				TimeInMatchSec float64
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Remove the foul from the correct alliance's list.
			deleteFoul := game.Foul{RuleId: args.RuleId, TeamId: args.TeamId, TimeInMatchSec: args.TimeInMatchSec}
			var fouls *[]game.Foul
			if args.Alliance == "red" {
				fouls = &web.arena.RedRealtimeScore.CurrentScore.Fouls
			} else {
				fouls = &web.arena.BlueRealtimeScore.CurrentScore.Fouls
			}
			for i, foul := range *fouls {
				if foul == deleteFoul {
					*fouls = append((*fouls)[:i], (*fouls)[i+1:]...)
					break
				}
			}
			web.arena.RealtimeScoreNotifier.Notify()
		case "card":
			args := struct {
				Alliance string
				TeamId   int
				Card     string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Set the card in the correct alliance's score.
			var cards map[string]string
			if args.Alliance == "red" {
				cards = web.arena.RedRealtimeScore.Cards
			} else {
				cards = web.arena.BlueRealtimeScore.Cards
			}
			cards[strconv.Itoa(args.TeamId)] = args.Card
			continue
		case "signalVolunteers":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldVolunteers = true
			continue // Don't reload.
		case "signalReset":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
			continue // Don't reload.
		case "commitMatch":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the fouls until the match is over.
				continue
			}
			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
			web.arena.ScoringStatusNotifier.Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Force a reload of the client to render the updated foul list.
		err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
