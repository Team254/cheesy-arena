// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field components.

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
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
	data := struct {
		*model.EventSettings
		AllianceStationDisplays map[string]string
		Inputs                  []bool
		Counters                []uint16
		Coils                   []bool
		LedMode                 int
	}{web.arena.EventSettings, web.arena.AllianceStationDisplays, web.arena.Plc.Inputs[:],
		web.arena.Plc.Registers[:], web.arena.Plc.Coils[:], web.arena.RedSwitchLedStrip.Mode}
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
	web.arena.MatchLoadTeamsNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/field", 303)
}

// Force-reloads all the websocket-connected displays.
func (web *Web) fieldReloadDisplaysHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.arena.ReloadDisplaysNotifier.Notify(nil)
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
	web.arena.RedSwitchLedStrip.SetMode(mode)

	http.Redirect(w, r, "/setup/field", 303)
}
