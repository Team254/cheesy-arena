// Copyright 2026 Team 254. All Rights Reserved.
//
// Web handlers for the hub status display. Each tablet loads this page with a "hub" URL parameter identifying which
// physical hub it represents (e.g. "1", "2", "3", ...); the tablet then shows only that hub's current color.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
)

// Renders the full-screen color display shown on the tablet mounted at a given hub.
func (web *Web) hubDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	template, err := web.parseFiles("templates/hub_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "hub_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the hub display client to receive color updates.
func (web *Web) hubDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	defer closeWebsocket(ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(
		display.Notifier,
		web.arena.HubColorNotifier,
		web.arena.ReloadDisplaysNotifier,
	)
}
