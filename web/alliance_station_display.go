// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the alliance station display.

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"log"
	"net/http"
)

// Renders the team number and status display shown above each alliance station.
func (web *Web) allianceStationDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	if !web.enforceDisplayConfiguration(w, r, map[string]string{"station": "R1"}) {
		return
	}

	template, err := web.parseFiles("templates/alliance_station_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "alliance_station_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the alliance station display client to receive status updates.
func (web *Web) allianceStationDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	display, err := web.registerDisplay(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Inform the client what the match period timing parameters are configured to.
	err = ws.Write("matchTiming", game.MatchTiming)
	if err != nil {
		log.Println(err)
		return
	}

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.AllianceStationDisplayModeNotifier, web.arena.ArenaStatusNotifier,
		web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.DisplayConfigurationNotifier, web.arena.ReloadDisplaysNotifier)
}
