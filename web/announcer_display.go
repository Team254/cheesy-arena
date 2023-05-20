// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for announcer display.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Renders the announcer display which shows team info and scores for the current match.
func (web *Web) announcerDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	template, err := web.parseFiles("templates/announcer_display.html", "templates/base.html")
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

// Renders a partial template for when a new match is loaded.
func (web *Web) announcerDisplayMatchLoadHandler(w http.ResponseWriter, r *http.Request) {
	template, err := web.parseFiles("templates/announcer_display_match_load.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	err = template.ExecuteTemplate(w, "announcer_display_match_load", web.arena.GenerateMatchLoadMessage())
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders a partial template for when a final score is posted.
func (web *Web) announcerDisplayScorePostedHandler(w http.ResponseWriter, r *http.Request) {
	template, err := web.parseFiles("templates/announcer_display_score_posted.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	err = template.ExecuteTemplate(w, "announcer_display_score_posted", web.arena.GenerateScorePostedMessage())
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the announcer display client to send control commands and receive status updates.
func (web *Web) announcerDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	ws.HandleNotifiers(display.Notifier, web.arena.MatchTimingNotifier, web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier, web.arena.ScorePostedNotifier,
		web.arena.AudienceDisplayModeNotifier, web.arena.ReloadDisplaysNotifier)
}
