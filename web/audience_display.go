// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for audience screen display.

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
)

// Renders the audience display to be chroma keyed over the video feed.
func (web *Web) audienceDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	template, err := web.parseFiles("templates/audience_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "audience_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the audience display client to receive status updates.
func (web *Web) audienceDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	audienceDisplayListener := web.arena.AudienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	matchLoadTeamsListener := web.arena.MatchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := web.arena.MatchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := web.arena.RealtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := web.arena.ScorePostedNotifier.Listen()
	defer close(scorePostedListener)
	playSoundListener := web.arena.PlaySoundNotifier.Listen()
	defer close(playSoundListener)
	allianceSelectionListener := web.arena.AllianceSelectionNotifier.Listen()
	defer close(allianceSelectionListener)
	lowerThirdListener := web.arena.LowerThirdNotifier.Listen()
	defer close(lowerThirdListener)
	reloadDisplaysListener := web.arena.ReloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
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
	err = websocket.Write("setAudienceDisplay", web.arena.AudienceDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		Match     *model.Match
		MatchName string
	}{web.arena.CurrentMatch, web.arena.CurrentMatch.CapitalizedType()}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore         *game.Score
		BlueScore        *game.Score
		RedScoreSummary  *game.ScoreSummary
		BlueScoreSummary *game.ScoreSummary
	}{&web.arena.RedRealtimeScore.CurrentScore, &web.arena.BlueRealtimeScore.CurrentScore,
		web.arena.RedScoreSummary(), web.arena.BlueScoreSummary()}
	err = websocket.Write("realtimeScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		Match     *model.Match
		MatchName string
		RedScore  *game.ScoreSummary
		BlueScore *game.ScoreSummary
	}{web.arena.SavedMatch, web.arena.SavedMatch.CapitalizedType(),
		web.arena.SavedMatchResult.RedScoreSummary(), web.arena.SavedMatchResult.BlueScoreSummary()}
	err = websocket.Write("setFinalScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("allianceSelection", cachedAlliances)
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
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = web.arena.AudienceDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				message = struct {
					Match     *model.Match
					MatchName string
				}{web.arena.CurrentMatch, web.arena.CurrentMatch.CapitalizedType()}
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
					RedScore         *game.Score
					BlueScore        *game.Score
					RedScoreSummary  *game.ScoreSummary
					BlueScoreSummary *game.ScoreSummary
				}{&web.arena.RedRealtimeScore.CurrentScore, &web.arena.BlueRealtimeScore.CurrentScore,
					web.arena.RedScoreSummary(), web.arena.BlueScoreSummary()}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					Match     *model.Match
					MatchName string
					RedScore  *game.ScoreSummary
					BlueScore *game.ScoreSummary
				}{web.arena.SavedMatch, web.arena.SavedMatch.CapitalizedType(),
					web.arena.SavedMatchResult.RedScoreSummary(), web.arena.SavedMatchResult.BlueScoreSummary()}
			case sound, ok := <-playSoundListener:
				if !ok {
					return
				}
				messageType = "playSound"
				message = sound
			case _, ok := <-allianceSelectionListener:
				if !ok {
					return
				}
				messageType = "allianceSelection"
				message = cachedAlliances
			case lowerThird, ok := <-lowerThirdListener:
				if !ok {
					return
				}
				messageType = "lowerThird"
				message = lowerThird
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
		_, _, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}
