// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for announcer display.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
)

// Renders the announcer display which shows team info and scores for the current match.
func (web *Web) announcerDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
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
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the announcer display client to send control commands and receive status updates.
func (web *Web) announcerDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := web.arena.MatchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := web.arena.MatchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := web.arena.RealtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := web.arena.ScorePostedNotifier.Listen()
	defer close(scorePostedListener)
	audienceDisplayListener := web.arena.AudienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	reloadDisplaysListener := web.arena.ReloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	data = struct {
		MatchType        string
		MatchDisplayName string
		Red1             *model.Team
		Red2             *model.Team
		Red3             *model.Team
		Blue1            *model.Team
		Blue2            *model.Team
		Blue3            *model.Team
	}{web.arena.CurrentMatch.CapitalizedType(), web.arena.CurrentMatch.DisplayName,
		web.arena.AllianceStations["R1"].Team, web.arena.AllianceStations["R2"].Team,
		web.arena.AllianceStations["R3"].Team, web.arena.AllianceStations["B1"].Team,
		web.arena.AllianceStations["B2"].Team, web.arena.AllianceStations["B3"].Team}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", game.MatchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{web.arena.MatchState, int(web.arena.LastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
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
					Red1             *model.Team
					Red2             *model.Team
					Red3             *model.Team
					Blue1            *model.Team
					Blue2            *model.Team
					Blue3            *model.Team
				}{web.arena.CurrentMatch.CapitalizedType(), web.arena.CurrentMatch.DisplayName,
					web.arena.AllianceStations["R1"].Team, web.arena.AllianceStations["R2"].Team,
					web.arena.AllianceStations["R3"].Team, web.arena.AllianceStations["B1"].Team,
					web.arena.AllianceStations["B2"].Team, web.arena.AllianceStations["B3"].Team}
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{web.arena.MatchState, matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					BlueScore int
				}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					MatchType        string
					MatchDisplayName string
					RedScoreSummary  *game.ScoreSummary
					BlueScoreSummary *game.ScoreSummary
					RedFouls         []game.Foul
					BlueFouls        []game.Foul
					RedCards         map[string]string
					BlueCards        map[string]string
				}{web.arena.SavedMatch.CapitalizedType(), web.arena.SavedMatch.DisplayName,
					web.arena.SavedMatchResult.RedScoreSummary(), web.arena.SavedMatchResult.BlueScoreSummary(),
					web.arena.SavedMatchResult.RedScore.Fouls, web.arena.SavedMatchResult.BlueScore.Fouls,
					web.arena.SavedMatchResult.RedCards, web.arena.SavedMatchResult.BlueCards}
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = web.arena.AudienceDisplayScreen
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
			web.arena.AudienceDisplayScreen = screen
			web.arena.AudienceDisplayNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
