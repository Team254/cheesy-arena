// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for displays.

package main

import (
	"net/http"
	"text/template"
)

// Renders the pit display which shows scrolling rankings.
func PitDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/pit_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.Execute(w, data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
