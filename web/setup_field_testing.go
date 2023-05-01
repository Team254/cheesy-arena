// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for testing the field sounds, LEDs, and PLC.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"io"
	"log"
	"net/http"
)

// Shows the Field Testing page.
func (web *Web) fieldTestingGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_field_testing.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	plc := web.arena.Plc
	data := struct {
		*model.EventSettings
		MatchSounds   []*game.MatchSound
		InputNames    []string
		RegisterNames []string
		CoilNames     []string
	}{web.arena.EventSettings, game.MatchSounds, plc.GetInputNames(), plc.GetRegisterNames(), plc.GetCoilNames()}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending realtime updates to the Field Testing page.
func (web *Web) fieldTestingWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.Plc.IoChangeNotifier())

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "playSound":
			sound, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.PlaySoundNotifier.NotifyWithMessage(sound)
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}
	}
}
