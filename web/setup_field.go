// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field components.

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"net/http"
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
		FieldTestMode           string
		Inputs                  []bool
		Counters                []uint16
		Coils                   []bool
	}{web.arena.EventSettings, web.arena.AllianceStationDisplays, web.arena.FieldTestMode, web.arena.Plc.Inputs[:],
		web.arena.Plc.Counters[:], web.arena.Plc.Coils[:]}
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

	// TODO(patrick): Update for 2018.
	mode := r.PostFormValue("mode")
	/*
		switch mode {
		case "boiler":
			web.arena.Plc.SetBoilerMotors(true)
			web.arena.Plc.SetRotorMotors(0, 0)
			web.arena.Plc.SetRotorLights(0, 0)
			web.arena.Plc.SetTouchpadLights([3]bool{false, false, false}, [3]bool{false, false, false})
		case "rotor1":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(1, 1)
			web.arena.Plc.SetRotorLights(1, 1)
			web.arena.Plc.SetTouchpadLights([3]bool{true, false, false}, [3]bool{true, false, false})
		case "rotor2":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(2, 2)
			web.arena.Plc.SetRotorLights(2, 2)
			web.arena.Plc.SetTouchpadLights([3]bool{false, true, false}, [3]bool{false, true, false})
		case "rotor3":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(3, 3)
			web.arena.Plc.SetRotorLights(2, 2)
			web.arena.Plc.SetTouchpadLights([3]bool{false, false, true}, [3]bool{false, false, true})
		case "rotor4":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(4, 4)
			web.arena.Plc.SetRotorLights(2, 2)
			web.arena.Plc.SetTouchpadLights([3]bool{false, false, false}, [3]bool{false, false, false})
		case "red":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(4, 0)
			web.arena.Plc.SetRotorLights(2, 0)
			web.arena.Plc.SetTouchpadLights([3]bool{true, true, true}, [3]bool{false, false, false})
		case "blue":
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(0, 4)
			web.arena.Plc.SetRotorLights(0, 2)
			web.arena.Plc.SetTouchpadLights([3]bool{false, false, false}, [3]bool{true, true, true})
		default:
			web.arena.Plc.SetBoilerMotors(false)
			web.arena.Plc.SetRotorMotors(0, 0)
			web.arena.Plc.SetRotorLights(0, 0)
			web.arena.Plc.SetTouchpadLights([3]bool{false, false, false}, [3]bool{false, false, false})
		}
	*/

	web.arena.FieldTestMode = mode
	http.Redirect(w, r, "/setup/field", 303)
}
