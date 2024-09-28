// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for managing scheduled breaks.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"net/http"
	"strconv"
)

// Shows the breaks configuration page.
func (web *Web) breaksGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_breaks.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	breaks, err := web.arena.Database.GetScheduledBreaksByMatchType(model.Playoff)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		ScheduledBreaks []model.ScheduledBreak
	}{web.arena.EventSettings, breaks}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the modified breaks to the database.
func (web *Web) breaksPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	scheduledBreakId, _ := strconv.Atoi(r.PostFormValue("id"))
	scheduledBreak, err := web.arena.Database.GetScheduledBreakById(scheduledBreakId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	scheduledBreak.Description = r.PostFormValue("description")
	if err = web.arena.Database.UpdateScheduledBreak(scheduledBreak); err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/breaks", 303)
}
