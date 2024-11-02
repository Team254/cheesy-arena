// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for a display to show a static logo and configurable message.

package web

import (
	"net/http"
	"time"

	"github.com/Team254/cheesy-arena/websocket"
)

// Renders the led view.
func (web *Web) redLedDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, map[string]string{"message": ""}) {
		return
	}

	template, err := web.parseFiles("templates/led_display_red.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		AmplifiedTimeRemaining float64
		BankedNotes            int
		CoopertitionActive     bool
		Amplified              bool
	}{
		AmplifiedTimeRemaining: web.arena.RedRealtimeScore.CurrentScore.AmpSpeaker.AmplifiedTimeRemaining(time.Now()),
		BankedNotes:            web.arena.RedRealtimeScore.CurrentScore.AmpSpeaker.BankedAmpNotes,
		CoopertitionActive:     web.arena.RedRealtimeScore.CurrentScore.AmpSpeaker.CoopActivated,
		Amplified:              web.arena.RedRealtimeScore.CurrentScore.AmpSpeaker.AmplifiedTimeRemaining(time.Now()) != 0,
	}
	err = template.ExecuteTemplate(w, "led_display_red.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders the led view.
func (web *Web) blueLedDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, map[string]string{"message": ""}) {
		return
	}

	template, err := web.parseFiles("templates/led_display_blue.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		AmplifiedTimeRemaining float64
		BankedNotes            int
		CoopertitionActive     bool
		Amplified              bool
	}{
		AmplifiedTimeRemaining: web.arena.BlueRealtimeScore.CurrentScore.AmpSpeaker.AmplifiedTimeRemaining(time.Now()),
		BankedNotes:            web.arena.BlueRealtimeScore.CurrentScore.AmpSpeaker.BankedAmpNotes,
		CoopertitionActive:     web.arena.BlueRealtimeScore.CurrentScore.AmpSpeaker.CoopActivated,
		Amplified:              web.arena.BlueRealtimeScore.CurrentScore.AmpSpeaker.AmplifiedTimeRemaining(time.Now()) != 0,
	}
	err = template.ExecuteTemplate(w, "led_display_blue.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending configuration commands to the display.
func (web *Web) ledDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	ws.HandleNotifiers(display.Notifier, web.arena.ReloadDisplaysNotifier, web.arena.RealtimeScoreNotifier)
}
