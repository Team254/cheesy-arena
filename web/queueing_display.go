// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for queueing display.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

const numMatchesToShow = 5

// Renders the queueing display that shows upcoming matches and timing information.
func (web *Web) queueingDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	matches, err := web.arena.Database.GetMatchesByType(web.arena.CurrentMatch.Type)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var upcomingMatches []model.Match
	for _, match := range matches {
		if match.Status == "complete" {
			continue
		}
		upcomingMatches = append(upcomingMatches, match)
		if len(upcomingMatches) == numMatchesToShow {
			break
		}
	}

	template, err := web.parseFiles("templates/queueing_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
		MatchTypePrefix string
		Matches         []model.Match
		StatusMessage   string
	}{web.arena.EventSettings, web.arena.CurrentMatch.TypePrefix(), upcomingMatches, generateEventStatusMessage(matches)}
	err = template.ExecuteTemplate(w, "queueing_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the queueing display to receive updates.
func (web *Web) queueingDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.MatchTimingNotifier, web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier,
		web.arena.DisplayConfigurationNotifier, web.arena.ReloadDisplaysNotifier)
}

// Returns a message indicating how early or late the event is running.
func generateEventStatusMessage(matches []model.Match) string {
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if match.Status == "complete" {
			minutesLate := match.StartedAt.Sub(match.Time).Minutes()
			if minutesLate > 2 {
				return fmt.Sprintf("Event is running %d minutes late", int(minutesLate))
			} else if minutesLate < -2 {
				return fmt.Sprintf("Event is running %d minutes early", int(-minutesLate))
			}
		}
	}

	if len(matches) > 0 {
		return "Event is running on schedule"
	} else {
		return ""
	}
}
