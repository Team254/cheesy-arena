// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for a placeholder display to be later configured by the server.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Shows a random ID to visually identify the display so that it can be configured on the server.
func (web *Web) placeholderDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	// Generate a display ID and redirect if the client doesn't already have one.
	displayId := r.URL.Query().Get("displayId")
	if displayId == "" {
		http.Redirect(w, r, fmt.Sprintf(r.URL.Path+"?displayId=%s", web.arena.NextDisplayId()), 302)
		return
	}

	template, err := web.parseFiles("templates/placeholder_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "placeholder_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending configuration commands to the display.
func (web *Web) placeholderDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	display, err := field.DisplayFromUrl(r.URL.Path, r.URL.Query())
	if err != nil {
		handleWebErr(w, err)
		return
	}
	web.arena.RegisterDisplay(display)
	defer web.arena.MarkDisplayDisconnected(display)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.DisplayConfigurationNotifier, web.arena.ReloadDisplaysNotifier)
}
