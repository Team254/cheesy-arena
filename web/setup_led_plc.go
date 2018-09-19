// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for testing the field LEDs and PLC.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/led"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/vaultled"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

// Shows the LED/PLC test page.
func (web *Web) ledPlcGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_led_plc.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	plc := web.arena.Plc
	data := struct {
		*model.EventSettings
		InputNames        []string
		RegisterNames     []string
		CoilNames         []string
		LedModeNames      map[led.Mode]string
		VaultLedModeNames map[vaultled.Mode]string
	}{web.arena.EventSettings, plc.GetInputNames(), plc.GetRegisterNames(), plc.GetCoilNames(), led.ModeNames,
		vaultled.ModeNames}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending realtime updates to the LED/PLC test page.
func (web *Web) ledPlcWebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
	go ws.HandleNotifiers(web.arena.LedModeNotifier, web.arena.Plc.IoChangeNotifier)

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
		case "setLedMode":
			if web.arena.MatchState != field.PreMatch && web.arena.MatchState != field.TimeoutActive &&
				web.arena.MatchState != field.PostTimeout {
				ws.WriteError("Arena must be in pre-match state")
				continue
			}
			var modeMessage field.LedModeMessage
			err = mapstructure.Decode(data, &modeMessage)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			web.arena.ScaleLeds.SetMode(modeMessage.LedMode, modeMessage.LedMode)
			web.arena.RedSwitchLeds.SetMode(modeMessage.LedMode, modeMessage.LedMode)
			web.arena.BlueSwitchLeds.SetMode(modeMessage.LedMode, modeMessage.LedMode)
			web.arena.RedVaultLeds.SetAllModes(modeMessage.VaultLedMode)
			web.arena.BlueVaultLeds.SetAllModes(modeMessage.VaultLedMode)
			web.arena.LedModeNotifier.Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
