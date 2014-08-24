// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field components.

package main

import (
	"html/template"
	"net/http"
)

// Shows the field configuration page.
func FieldGetHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/field.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		AllianceStationDisplays map[string]string
	}{eventSettings, mainArena.allianceStationDisplays}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates the display-station mapping for a single display.
func FieldPostHandler(w http.ResponseWriter, r *http.Request) {
	displayId := r.PostFormValue("displayId")
	allianceStation := r.PostFormValue("allianceStation")
	mainArena.allianceStationDisplays[displayId] = allianceStation
	mainArena.matchLoadTeamsNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/field", 302)
}

// Force-reloads all the websocket-connected displays.
func FieldReloadDisplaysHandler(w http.ResponseWriter, r *http.Request) {
	mainArena.reloadDisplaysNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/field", 302)
}
