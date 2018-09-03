// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the FTA diagnostic display.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Renders the FTA diagnostic display.
func (web *Web) ftaDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/fta_display.html", "templates/base.html")
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

// The websocket endpoint for the FTA display client to receive status updates.
func (web *Web) ftaDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(patrick): Enable authentication once Safari (for iPad) supports it over Websocket.

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.ArenaStatusNotifier, web.arena.ReloadDisplaysNotifier)
}
