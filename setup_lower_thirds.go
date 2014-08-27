// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for managing lower thirds.

package main

import (
	"html/template"
	"net/http"
	"strconv"
)

// Shows the lower third configuration page.
func LowerThirdsGetHandler(w http.ResponseWriter, r *http.Request) {
	if auth.Authorize(r) == "" {
		auth.NotifyAuthRequired(w, r)
		return
	}

	template, err := template.ParseFiles("templates/lower_thirds.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	lowerThirds, err := db.GetAllLowerThirds()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		LowerThirds []LowerThird
	}{eventSettings, lowerThirds}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the new or modified lower third to the database and triggers showing it on the audience display.
func LowerThirdsPostHandler(w http.ResponseWriter, r *http.Request) {
	if auth.Authorize(r) == "" {
		auth.NotifyAuthRequired(w, r)
		return
	}

	lowerThirdId, _ := strconv.Atoi(r.PostFormValue("id"))
	lowerThird, err := db.GetLowerThirdById(lowerThirdId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if r.PostFormValue("action") == "delete" {
		err := db.DeleteLowerThird(lowerThird)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		// Save the lower third even if the show or hide buttons were clicked.
		if lowerThird == nil {
			lowerThird = &LowerThird{TopText: r.PostFormValue("topText"),
				BottomText: r.PostFormValue("bottomText")}
			err = db.CreateLowerThird(lowerThird)
		} else {
			lowerThird.TopText = r.PostFormValue("topText")
			lowerThird.BottomText = r.PostFormValue("bottomText")
			err = db.SaveLowerThird(lowerThird)
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}

		if r.PostFormValue("action") == "show" {
			mainArena.lowerThirdNotifier.Notify(lowerThird)
			mainArena.audienceDisplayScreen = "lowerThird"
			mainArena.audienceDisplayNotifier.Notify(nil)
		} else if r.PostFormValue("action") == "hide" {
			mainArena.audienceDisplayScreen = "blank"
			mainArena.audienceDisplayNotifier.Notify(nil)
		}
	}

	http.Redirect(w, r, "/setup/lower_thirds", 302)
}
