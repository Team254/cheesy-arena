// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for a display to show a configurable YouTube live video stream.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Renders the YouTube stream view.
func (web *Web) youtubeDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, map[string]string{"videoId": "", "captions": "false"}) {
		return
	}

	template, err := web.parseFiles("templates/youtube_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "youtube_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending configuration commands to the display.
func (web *Web) youtubeDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	ws.HandleNotifiers(display.Notifier, web.arena.ReloadDisplaysNotifier)
}
