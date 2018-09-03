// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field displays.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"net/http"
)

// Shows the displays configuration page.
func (web *Web) displaysGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_displays.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		AllianceStationDisplays map[string]string
	}{web.arena.EventSettings, web.arena.AllianceStationDisplays}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates the display-station mapping for a single display.
func (web *Web) displaysPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	displayId := r.PostFormValue("displayId")
	allianceStation := r.PostFormValue("allianceStation")
	web.arena.AllianceStationDisplays[displayId] = allianceStation
	web.arena.MatchLoadNotifier.Notify()
	http.Redirect(w, r, "/setup/displays", 303)
}

// Force-reloads all the websocket-connected displays.
func (web *Web) displaysReloadHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.arena.ReloadDisplaysNotifier.Notify()
	http.Redirect(w, r, "/setup/displays", 303)
}
