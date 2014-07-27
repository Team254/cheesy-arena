// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for displays.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the pit display which shows scrolling rankings.
func PitDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/pit_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.Execute(w, data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders the announcer display which shows team info and scores for the current match.
func AnnouncerDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/announcer_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Assemble info about the current match.
	matchType := mainArena.currentMatch.CapitalizedType()
	red1 := mainArena.AllianceStations["R1"].team
	red2 := mainArena.AllianceStations["R2"].team
	red3 := mainArena.AllianceStations["R3"].team
	blue1 := mainArena.AllianceStations["B1"].team
	blue2 := mainArena.AllianceStations["B2"].team
	blue3 := mainArena.AllianceStations["B3"].team

	// Assemble info about the saved match result.
	var redScoreSummary, blueScoreSummary *ScoreSummary
	var savedMatchType, savedMatchDisplayName string
	if mainArena.savedMatchResult != nil {
		redScoreSummary = mainArena.savedMatchResult.RedScoreSummary()
		blueScoreSummary = mainArena.savedMatchResult.BlueScoreSummary()
		match, err := db.GetMatchById(mainArena.savedMatchResult.MatchId)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		savedMatchType = match.CapitalizedType()
		savedMatchDisplayName = match.DisplayName
	}
	data := struct {
		*EventSettings
		MatchType             string
		MatchDisplayName      string
		Red1                  *Team
		Red2                  *Team
		Red3                  *Team
		Blue1                 *Team
		Blue2                 *Team
		Blue3                 *Team
		SavedMatchResult      *MatchResult
		SavedMatchType        string
		SavedMatchDisplayName string
		RedScoreSummary       *ScoreSummary
		BlueScoreSummary      *ScoreSummary
	}{eventSettings, matchType, mainArena.currentMatch.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		mainArena.savedMatchResult, savedMatchType, savedMatchDisplayName, redScoreSummary, blueScoreSummary}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the announcer display client to send control commands and receive status updates.
func AnnouncerDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	scorePostedListener := mainArena.scorePostedNotifier.Listen()

	// Send the various notifications immediately upon connection.
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data := MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)}
	err = websocket.Write("matchTime", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{mainArena.MatchState, matchTimeSec.(int)}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			}
			err = websocket.Write(messageType, message)
			if err != nil {
				// The client has probably closed the connection; nothing to do here.
				return
			}
		}
	}()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, _, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}

		switch messageType {
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}
	}
}
