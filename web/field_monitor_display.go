// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the field monitor display showing robot connection status.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

// Renders the field monitor display.
func (web *Web) fieldMonitorDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("fta") == "true" && !web.userIsAdmin(w, r) {
		return
	}

	if !web.enforceDisplayConfiguration(w, r, map[string]string{"reversed": "false", "fta": "false"}) {
		return
	}

	template, err := web.parseFiles("templates/field_monitor_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "field_monitor_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the field monitor display client to receive status updates.
func (web *Web) fieldMonitorDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	isFta := r.URL.Query().Get("fta") == "true"
	if isFta && !web.userIsAdmin(w, r) {
		return
	}

	display, err := web.registerDisplay(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchTimingNotifier, display.Notifier, web.arena.ArenaStatusNotifier,
		web.arena.EventStatusNotifier, web.arena.RealtimeScoreNotifier, web.arena.MatchTimeNotifier,
		web.arena.MatchLoadNotifier, web.arena.ReloadDisplaysNotifier)

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

		if command == "updateTeamNotes" {
			if isFta {
				args := struct {
					Station string
					Notes   string
				}{}
				err = mapstructure.Decode(data, &args)
				if err != nil {
					ws.WriteError(err.Error())
					continue
				}

				if allianceStation, ok := web.arena.AllianceStations[args.Station]; ok {
					if allianceStation.Team != nil {
						allianceStation.Team.FtaNotes = args.Notes
						if err := web.arena.Database.UpdateTeam(allianceStation.Team); err != nil {
							ws.WriteError(err.Error())
						}
						web.arena.ArenaStatusNotifier.Notify()
					} else {
						ws.WriteError("No team present")
					}
				} else {
					ws.WriteError("Invalid alliance station")
				}
			} else {
				ws.WriteError("Must be in FTA mode to update team notes")
			}
		}
	}
}
