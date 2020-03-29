// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for queueing display.

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
	"time"
)

const (
	numMatchesToShow = 5
)

// Renders the queueing display that shows upcoming matches and timing information.
func (web *Web) queueingDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	matches, err := web.arena.Database.GetMatchesByType(web.arena.CurrentMatch.Type)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var upcomingMatches []model.Match
	for i, match := range matches {
		if match.IsComplete() {
			continue
		}
		upcomingMatches = append(upcomingMatches, match)
		if len(upcomingMatches) == numMatchesToShow {
			break
		}

		// Don't include any more matches if there is a significant gap before the next one.
		if i+1 < len(matches) && matches[i+1].Time.Sub(match.Time) > field.MaxMatchGapMin*time.Minute {
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
	}{web.arena.EventSettings, web.arena.CurrentMatch.TypePrefix(), upcomingMatches}
	err = template.ExecuteTemplate(w, "queueing_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the queueing display to receive updates.
func (web *Web) queueingDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
		web.arena.MatchTimeNotifier, web.arena.EventStatusNotifier, web.arena.ReloadDisplaysNotifier)
}
