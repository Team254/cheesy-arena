// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the FTA diagnostic display.

package main

import (
	"net/http"
	"text/template"
)

// Renders the FTA diagnostic display.
func FtaDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/fta_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
