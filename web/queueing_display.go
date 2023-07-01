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
	numNonPlayoffMatchesToShow = 5
	numPlayoffMatchesToShow    = 4
)

// Renders the queueing display that shows upcoming matches and timing information.
func (web *Web) queueingDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	template, err := web.parseFiles("templates/queueing_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{
		web.arena.EventSettings,
	}
	err = template.ExecuteTemplate(w, "queueing_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders a partial template containing the list of matches.
func (web *Web) queueingDisplayMatchLoadHandler(w http.ResponseWriter, r *http.Request) {
	matches, err := web.arena.Database.GetMatchesByType(web.arena.CurrentMatch.Type, false)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	numMatchesToShow := numNonPlayoffMatchesToShow
	if web.arena.CurrentMatch.Type == model.Playoff {
		numMatchesToShow = numPlayoffMatchesToShow
	}

	var upcomingMatches []model.Match
	var redOffFieldTeamsByMatch, blueOffFieldTeamsByMatch [][]int
	if err != nil {
		handleWebErr(w, err)
		return
	}
	for i, match := range matches {
		if match.IsComplete() || match.TypeOrder < web.arena.CurrentMatch.TypeOrder {
			continue
		}
		upcomingMatches = append(upcomingMatches, match)
		redOffFieldTeams, blueOffFieldTeams, err := web.arena.Database.GetOffFieldTeamIds(&match)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		redOffFieldTeamsByMatch = append(redOffFieldTeamsByMatch, redOffFieldTeams)
		blueOffFieldTeamsByMatch = append(blueOffFieldTeamsByMatch, blueOffFieldTeams)

		if len(upcomingMatches) == numMatchesToShow {
			break
		}

		// Don't include any more matches if there is a significant gap before the next one.
		if i+1 < len(matches) && matches[i+1].Time.Sub(match.Time) > field.MaxMatchGapMin*time.Minute {
			break
		}
	}

	template, err := web.parseFiles("templates/queueing_display_match_load.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		Matches           []model.Match
		RedOffFieldTeams  [][]int
		BlueOffFieldTeams [][]int
	}{
		upcomingMatches,
		redOffFieldTeamsByMatch,
		blueOffFieldTeamsByMatch,
	}
	err = template.ExecuteTemplate(w, "queueing_display_match_load.html", data)
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
