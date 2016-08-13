// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for audience screen display.

package main

import (
	"io"
	"log"
	"net/http"
	"text/template"
)

type RealtimeScoreFields struct {
	Score            int
	TowerStrength    int
	DefensesStrength [5]int
}

// Renders the audience display to be chroma keyed over the video feed.
func AudienceDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/audience_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "audience_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the audience display client to receive status updates.
func AudienceDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	audienceDisplayListener := mainArena.audienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := mainArena.scorePostedNotifier.Listen()
	defer close(scorePostedListener)
	playSoundListener := mainArena.playSoundNotifier.Listen()
	defer close(playSoundListener)
	allianceSelectionListener := mainArena.allianceSelectionNotifier.Listen()
	defer close(allianceSelectionListener)
	lowerThirdListener := mainArena.lowerThirdNotifier.Listen()
	defer close(lowerThirdListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
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
	err = websocket.Write("setAudienceDisplay", mainArena.audienceDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		Match     *Match
		MatchName string
	}{mainArena.currentMatch, mainArena.currentMatch.CapitalizedType()}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScoreFields  *RealtimeScoreFields
		BlueScoreFields *RealtimeScoreFields
	}{mainArena.redRealtimeScore.ScoreFields(mainArena.blueRealtimeScore.CurrentScore.Fouls),
		mainArena.blueRealtimeScore.ScoreFields(mainArena.redRealtimeScore.CurrentScore.Fouls)}
	err = websocket.Write("realtimeScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		Match     *Match
		MatchName string
		RedScore  *ScoreSummary
		BlueScore *ScoreSummary
	}{mainArena.savedMatch, mainArena.savedMatch.CapitalizedType(),
		mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary()}
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
				message = mainArena.audienceDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				message = struct {
					Match     *Match
					MatchName string
				}{mainArena.currentMatch, mainArena.currentMatch.CapitalizedType()}
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
					RedScoreFields  *RealtimeScoreFields
					BlueScoreFields *RealtimeScoreFields
				}{mainArena.redRealtimeScore.ScoreFields(mainArena.blueRealtimeScore.CurrentScore.Fouls),
					mainArena.blueRealtimeScore.ScoreFields(mainArena.redRealtimeScore.CurrentScore.Fouls)}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					Match     *Match
					MatchName string
					RedScore  *ScoreSummary
					BlueScore *ScoreSummary
				}{mainArena.savedMatch, mainArena.savedMatch.CapitalizedType(),
					mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary()}
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

// Calculates the integer score, tower strength, and defenses strength for the given realtime snapshot.
func (realtimeScore *RealtimeScore) ScoreFields(opponentFouls []Foul) *RealtimeScoreFields {
	scoreSummary := scoreSummary(&realtimeScore.CurrentScore, opponentFouls, mainArena.currentMatch.Type)
	var defensesStrength [5]int
	for i := 0; i < 5; i++ {
		defensesStrength[i] = 2 - realtimeScore.CurrentScore.AutoDefensesCrossed[i] -
			realtimeScore.CurrentScore.DefensesCrossed[i]
	}
	return &RealtimeScoreFields{scoreSummary.Score, scoreSummary.TowerStrength, defensesStrength}
}
