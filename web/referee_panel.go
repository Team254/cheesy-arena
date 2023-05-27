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

	template, err := web.parseFiles("templates/referee_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders a partial template for when the foul list is updated.
func (web *Web) refereePanelFoulListHandler(w http.ResponseWriter, r *http.Request) {
	template, err := web.parseFiles("templates/referee_panel_foul_list.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		Match     *model.Match
		RedFouls  []game.Foul
		BlueFouls []game.Foul
		Rules     map[int]*game.Rule
	}{
		web.arena.CurrentMatch,
		web.arena.RedRealtimeScore.CurrentScore.Fouls,
		web.arena.BlueRealtimeScore.CurrentScore.Fouls,
		game.GetAllRules(),
	}
	err = template.ExecuteTemplate(w, "referee_panel_foul_list", data)
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
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ScoringStatusNotifier,
		web.arena.ReloadDisplaysNotifier,
	)

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
				Alliance    string
				IsTechnical bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := game.Foul{IsTechnical: args.IsTechnical}
			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			web.arena.RealtimeScoreNotifier.Notify()
		case "toggleFoulType", "updateFoulTeam", "updateFoulRule", "deleteFoul":
			args := struct {
				Alliance string
				Index    int
				TeamId   int
				RuleId   int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Find the foul in the correct alliance's list.
			var fouls *[]game.Foul
			if args.Alliance == "red" {
				fouls = &web.arena.RedRealtimeScore.CurrentScore.Fouls
			} else {
				fouls = &web.arena.BlueRealtimeScore.CurrentScore.Fouls
			}
			if args.Index >= 0 && args.Index < len(*fouls) {
				switch messageType {
				case "toggleFoulType":
					(*fouls)[args.Index].IsTechnical = !(*fouls)[args.Index].IsTechnical
					(*fouls)[args.Index].RuleId = 0
				case "deleteFoul":
					*fouls = append((*fouls)[:args.Index], (*fouls)[args.Index+1:]...)
				case "updateFoulTeam":
					if (*fouls)[args.Index].TeamId == args.TeamId {
						(*fouls)[args.Index].TeamId = 0
					} else {
						(*fouls)[args.Index].TeamId = args.TeamId
					}
				case "updateFoulRule":
					(*fouls)[args.Index].RuleId = args.RuleId
				}
				web.arena.RealtimeScoreNotifier.Notify()
			}
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
			web.arena.RealtimeScoreNotifier.Notify()
		case "signalVolunteers":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldVolunteers = true
		case "signalReset":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
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
		}
	}
}
