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
	"time"
)

const (
	earlyLateThresholdMin = 2
	maxGapMin             = 20
	numMatchesToShow      = 5
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
		if match.Status == "complete" {
			continue
		}
		upcomingMatches = append(upcomingMatches, match)
		if len(upcomingMatches) == numMatchesToShow {
			break
		}

		// Don't include any more matches if there is a significant gap before the next one.
		if i+1 < len(matches) && matches[i+1].Time.Sub(match.Time) > maxGapMin*time.Minute {
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
	}{web.arena.EventSettings, web.arena.CurrentMatch.TypePrefix(), upcomingMatches,
		generateEventStatusMessage(web.arena.CurrentMatch.Type, matches)}
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
func generateEventStatusMessage(matchType string, matches []model.Match) string {
	if matchType != "practice" && matchType != "qualification" {
		// Only practice and qualification matches have a strict schedule.
		return ""
	}
	if len(matches) == 0 || matches[len(matches)-1].Status == "complete" {
		// All matches of the current type are complete.
		return ""
	}

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if match.Status == "complete" {
			if i+1 < len(matches) && matches[i+1].Time.Sub(match.Time) > maxGapMin*time.Minute {
				break
			} else {
				minutesLate := match.StartedAt.Sub(match.Time).Minutes()
				if minutesLate > earlyLateThresholdMin {
					return fmt.Sprintf("Event is running %d minutes late", int(minutesLate))
				} else if minutesLate < -earlyLateThresholdMin {
					return fmt.Sprintf("Event is running %d minutes early", int(-minutesLate))
				}
			}
		}
	}

	return "Event is running on schedule"
}
