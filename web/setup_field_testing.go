// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for testing the field sounds, LEDs, and PLC.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

const fieldTestingOverrideDisabledMessage = "Cannot override coil while match is in progress."

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
	}{
		web.arena.EventSettings,
		game.UniqueMatchSounds(),
		plc.GetInputNames(),
		plc.GetRegisterNames(),
		plc.GetCoilNames(),
	}
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
	go ws.HandleNotifiers(web.arena.Plc.IoChangeNotifier(), web.arena.ArenaStatusNotifier)

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
		case "setPlcCoilOverride":
			args := struct {
				Index    int
				Override string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			if !fieldTestingOverridesAllowed(web.arena.MatchState) {
				ws.WriteError(fieldTestingOverrideDisabledMessage)
				continue
			}

			switch args.Override {
			case "auto":
				web.arena.Plc.ClearCoilOverride(args.Index)
			case "on":
				web.arena.Plc.SetCoilOverride(args.Index, true)
			case "off":
				web.arena.Plc.SetCoilOverride(args.Index, false)
			default:
				ws.WriteError(fmt.Sprintf("Invalid coil override state '%s'.", args.Override))
				continue
			}
			web.arena.Plc.IoChangeNotifier().Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}
	}
}

func fieldTestingOverridesAllowed(matchState field.MatchState) bool {
	return matchState == field.PreMatch || matchState == field.PostMatch || matchState == field.TimeoutActive ||
		matchState == field.PostTimeout
}
