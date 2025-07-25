// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for generating judging schedules.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"net/http"
	"sort"
	"strconv"
)

var judgingScheduleParams = tournament.JudgingScheduleParams{
	NumJudges:              5,
	DurationMinutes:        15,
	PreviousSpacingMinutes: 20,
	NextSpacingMinutes:     20,
}

// Shows the judging schedule setup page.
func (web *Web) judgingGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.renderJudging(w, r, "")
}

// Generates a judging schedule based on the parameters and saves it to the database.
func (web *Web) judgingGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	numJudges, err := strconv.Atoi(r.PostFormValue("numJudges"))
	if err != nil || numJudges <= 0 {
		web.renderJudging(w, r, "Number of judges must be a positive integer.")
		return
	}
	durationMinutes, err := strconv.Atoi(r.PostFormValue("durationMinutes"))
	if err != nil || durationMinutes <= 0 {
		web.renderJudging(w, r, "Visit duration must be a positive integer.")
		return
	}
	previousSpacingMinutes, err := strconv.Atoi(r.PostFormValue("previousSpacingMinutes"))
	if err != nil || previousSpacingMinutes <= 0 {
		web.renderJudging(w, r, "Minimum spacing after previous match must be a positive integer.")
		return
	}
	nextSpacingMinutes, err := strconv.Atoi(r.PostFormValue("nextSpacingMinutes"))
	if err != nil || nextSpacingMinutes <= 0 {
		web.renderJudging(w, r, "Minimum spacing before next match must be a positive integer.")
		return
	}
	qualMatches, err := web.arena.Database.GetMatchesByType(model.Qualification, true)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(qualMatches) == 0 {
		web.renderJudging(w, r, "No qualification matches found. Generate the qualification schedule first.")
		return
	}
	slots, err := web.arena.Database.GetAllJudgingSlots()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(slots) > 0 {
		web.renderJudging(w, r, "Judging schedule already exists. Clear it first before generating a new one.")
		return
	}

	judgingScheduleParams.NumJudges = numJudges
	judgingScheduleParams.DurationMinutes = durationMinutes
	judgingScheduleParams.PreviousSpacingMinutes = previousSpacingMinutes
	judgingScheduleParams.NextSpacingMinutes = nextSpacingMinutes
	err = tournament.BuildJudgingSchedule(web.arena.Database, judgingScheduleParams)
	if err != nil {
		web.renderJudging(w, r, fmt.Sprintf("Error generating judging schedule: %s", err.Error()))
		return
	}

	http.Redirect(w, r, "/setup/judging", 303)
}

// Clears the judging schedule.
func (web *Web) judgingClearPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if err := web.arena.Database.TruncateJudgingSlots(); err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/judging", 303)
}

// Renders the judging setup page with an optional error message.
func (web *Web) renderJudging(w http.ResponseWriter, r *http.Request, errorMessage string) {
	slots, err := web.arena.Database.GetAllJudgingSlots()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Sort slots by judge team and then by time for display.
	sort.Slice(
		slots,
		func(i, j int) bool {
			if slots[i].JudgeNumber != slots[j].JudgeNumber {
				return slots[i].JudgeNumber < slots[j].JudgeNumber
			}
			return slots[i].Time.Before(slots[j].Time)
		},
	)

	template, err := web.parseFiles("templates/setup_judging.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		JudgingScheduleParams tournament.JudgingScheduleParams
		JudgingSlots          []model.JudgingSlot
		ErrorMessage          string
	}{web.arena.EventSettings, judgingScheduleParams, slots, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
