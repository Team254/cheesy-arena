// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for announcer display.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the announcer display which shows team info and scores for the current match.
func AnnouncerDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/announcer_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
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
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := mainArena.scorePostedNotifier.Listen()
	defer close(scorePostedListener)
	audienceDisplayListener := mainArena.audienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	data = struct {
		MatchType        string
		MatchDisplayName string
		Red1             *Team
		Red2             *Team
		Red3             *Team
		Blue1            *Team
		Blue2            *Team
		Blue3            *Team
	}{mainArena.currentMatch.CapitalizedType(), mainArena.currentMatch.DisplayName,
		mainArena.AllianceStations["R1"].team, mainArena.AllianceStations["R2"].team,
		mainArena.AllianceStations["R3"].team, mainArena.AllianceStations["B1"].team,
		mainArena.AllianceStations["B2"].team, mainArena.AllianceStations["B3"].team}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{mainArena.redRealtimeScore.Score(), mainArena.blueRealtimeScore.Score()}
	err = websocket.Write("realtimeScore", data)
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
				messageType = "setMatch"
				message = struct {
					MatchType        string
					MatchDisplayName string
					Red1             *Team
					Red2             *Team
					Red3             *Team
					Blue1            *Team
					Blue2            *Team
					Blue3            *Team
				}{mainArena.currentMatch.CapitalizedType(), mainArena.currentMatch.DisplayName,
					mainArena.AllianceStations["R1"].team, mainArena.AllianceStations["R2"].team,
					mainArena.AllianceStations["R3"].team, mainArena.AllianceStations["B1"].team,
					mainArena.AllianceStations["B2"].team, mainArena.AllianceStations["B3"].team}
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{mainArena.MatchState, matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					BlueScore int
				}{mainArena.redRealtimeScore.Score(), mainArena.blueRealtimeScore.Score()}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					MatchType        string
					MatchDisplayName string
					RedScoreSummary  *ScoreSummary
					BlueScoreSummary *ScoreSummary
					RedFouls         []Foul
					BlueFouls        []Foul
					RedCards         map[string]string
					BlueCards        map[string]string
				}{mainArena.savedMatch.CapitalizedType(), mainArena.savedMatch.DisplayName,
					mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary(),
					mainArena.savedMatchResult.RedScore.Fouls, mainArena.savedMatchResult.BlueScore.Fouls,
					mainArena.savedMatchResult.RedCards, mainArena.savedMatchResult.BlueCards}
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = mainArena.audienceDisplayScreen
			case _, ok := <-reloadDisplaysListener:
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
		messageType, data, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}

		switch messageType {
		case "setAudienceDisplay":
			// The announcer can make the final score screen show when they are ready to announce the score.
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.audienceDisplayScreen = screen
			mainArena.audienceDisplayNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
