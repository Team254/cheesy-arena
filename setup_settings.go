// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the event settings.

package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strconv"
)

// Shows the event settings editing page.
func SettingsGetHandler(w http.ResponseWriter, r *http.Request) {
	renderSettings(w, r, "")
}

// Saves the event settings.
func SettingsPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eventSettings.Name = r.PostFormValue("name")
	eventSettings.Code = r.PostFormValue("code")
	match, _ := regexp.MatchString("^#([0-9A-Fa-f]{3}){1,2}$", r.PostFormValue("displayBackgroundColor"))
	if !match {
		renderSettings(w, r, "Display background color must be a valid hex color value.")
		return
	}
	eventSettings.DisplayBackgroundColor = r.PostFormValue("displayBackgroundColor")
	numAlliances, _ := strconv.Atoi(r.PostFormValue("numElimAlliances"))
	if numAlliances < 2 || numAlliances > 16 {
		renderSettings(w, r, "Number of alliances must be between 2 and 16.")
		return
	}
	eventSettings.NumElimAlliances = numAlliances
	eventSettings.SelectionRound1Order = r.PostFormValue("selectionRound1Order")
	eventSettings.SelectionRound2Order = r.PostFormValue("selectionRound2Order")
	eventSettings.SelectionRound3Order = r.PostFormValue("selectionRound3Order")
	err = db.SaveEventSettings(eventSettings)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	renderSettings(w, r, "")
}

func renderSettings(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := template.ParseFiles("templates/settings.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		ErrorMessage string
	}{eventSettings, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
