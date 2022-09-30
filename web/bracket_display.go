// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the bracket display.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Renders the display which shows the playoff bracket.
func (web *Web) bracketDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	template, err := web.parseFiles("templates/bracket_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "bracket_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the bracket display.
func (web *Web) bracketDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(display.Notifier, web.arena.MatchLoadNotifier, web.arena.ReloadDisplaysNotifier)
}
