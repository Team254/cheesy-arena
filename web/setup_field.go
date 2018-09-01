// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field components.

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/led"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/vaultled"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
	"strconv"
)

// Shows the field configuration page.
func (web *Web) fieldGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_field.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	plc := web.arena.Plc
	data := struct {
		*model.EventSettings
		AllianceStationDisplays map[string]string
		InputNames              []string
		RegisterNames           []string
		CoilNames               []string
		CurrentLedMode          led.Mode
		LedModeNames            map[led.Mode]string
		CurrentVaultLedMode     vaultled.Mode
		VaultLedModeNames       map[vaultled.Mode]string
	}{web.arena.EventSettings, web.arena.AllianceStationDisplays, plc.GetInputNames(), plc.GetRegisterNames(),
		plc.GetCoilNames(), web.arena.ScaleLeds.GetCurrentMode(), led.ModeNames,
		web.arena.RedVaultLeds.CurrentForceMode, vaultled.ModeNames}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates the display-station mapping for a single display.
func (web *Web) fieldPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	displayId := r.PostFormValue("displayId")
	allianceStation := r.PostFormValue("allianceStation")
	web.arena.AllianceStationDisplays[displayId] = allianceStation
	web.arena.MatchLoadNotifier.Notify()
	http.Redirect(w, r, "/setup/field", 303)
}

// Force-reloads all the websocket-connected displays.
func (web *Web) fieldReloadDisplaysHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.arena.ReloadDisplaysNotifier.Notify()
	http.Redirect(w, r, "/setup/field", 303)
}

// Controls the field LEDs for testing or effect.
func (web *Web) fieldTestPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.MatchState != field.PreMatch {
		http.Error(w, "Arena must be in pre-match state", 400)
		return
	}

	mode, _ := strconv.Atoi(r.PostFormValue("mode"))
	ledMode := led.Mode(mode)
	web.arena.ScaleLeds.SetMode(ledMode, ledMode)
	web.arena.RedSwitchLeds.SetMode(ledMode, ledMode)
	web.arena.BlueSwitchLeds.SetMode(ledMode, ledMode)

	vaultMode, _ := strconv.Atoi(r.PostFormValue("vaultMode"))
	vaultLedMode := vaultled.Mode(vaultMode)
	web.arena.RedVaultLeds.SetAllModes(vaultLedMode)
	web.arena.BlueVaultLeds.SetAllModes(vaultLedMode)

	http.Redirect(w, r, "/setup/field", 303)
}

// The websocket endpoint for sending realtime updates to the field setup page.
func (web *Web) fieldWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.Plc.IoChangeNotifier)
}
