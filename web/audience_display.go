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
	"time"
)

type audienceScoreFields struct {
	Red          *audienceAllianceScoreFields
	Blue         *audienceAllianceScoreFields
	ScaleOwnedBy game.Alliance
}

type audienceAllianceScoreFields struct {
	Score         int
	ForceCubes    int
	LevitateCubes int
	BoostCubes    int
	ForceState    game.PowerUpState
	LevitateState game.PowerUpState
	BoostState    game.PowerUpState
	SwitchOwnedBy game.Alliance
}

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
	err = websocket.Write("matchTime", MatchTimeMessage{int(web.arena.MatchState), int(web.arena.LastMatchTimeSec)})
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
	data = web.getAudienceScoreFields()
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
				message = MatchTimeMessage{int(web.arena.MatchState), matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = web.getAudienceScoreFields()
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

// Constructs the data object sent to the audience display for the realtime scoring overlay.
func (web *Web) getAudienceScoreFields() *audienceScoreFields {
	fields := new(audienceScoreFields)
	fields.Red = getAudienceAllianceScoreFields(&web.arena.RedRealtimeScore.CurrentScore, web.arena.RedScoreSummary(),
		web.arena.RedVault, web.arena.RedSwitch)
	fields.Blue = getAudienceAllianceScoreFields(&web.arena.BlueRealtimeScore.CurrentScore,
		web.arena.BlueScoreSummary(), web.arena.BlueVault, web.arena.BlueSwitch)
	fields.ScaleOwnedBy = web.arena.Scale.GetOwnedBy()
	return fields
}

// Constructs the data object for one alliance sent to the audience display for the realtime scoring overlay.
func getAudienceAllianceScoreFields(allianceScore *game.Score, allianceScoreSummary *game.ScoreSummary,
	allianceVault *game.Vault, allianceSwitch *game.Seesaw) *audienceAllianceScoreFields {
	fields := new(audienceAllianceScoreFields)
	fields.Score = allianceScoreSummary.Score
	fields.ForceCubes = allianceScore.ForceCubes
	fields.LevitateCubes = allianceScore.LevitateCubes
	fields.BoostCubes = allianceScore.BoostCubes
	if allianceVault.ForcePowerUp != nil {
		fields.ForceState = allianceVault.ForcePowerUp.GetState(time.Now())
	} else {
		fields.ForceState = game.Unplayed
	}
	if allianceVault.LevitatePlayed {
		fields.LevitateState = game.Expired
	} else {
		fields.LevitateState = game.Unplayed
	}
	if allianceVault.BoostPowerUp != nil {
		fields.BoostState = allianceVault.BoostPowerUp.GetState(time.Now())
	} else {
		fields.BoostState = game.Unplayed
	}
	fields.SwitchOwnedBy = allianceSwitch.GetOwnedBy()
	return fields
}
