// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for generating practice and qualification schedules.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"net/http"
	"strconv"
	"time"
)

// Global vars to hold schedules that are in the process of being generated.
var cachedMatchType string
var cachedScheduleBlocks []tournament.ScheduleBlock
var cachedMatches []model.Match
var cachedTeamFirstMatches map[int]string

// Shows the schedule editing page.
func (web *Web) scheduleGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if len(cachedScheduleBlocks) == 0 {
		tomorrow := time.Now().AddDate(0, 0, 1)
		location, _ := time.LoadLocation("Local")
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, location)
		cachedScheduleBlocks = append(cachedScheduleBlocks, tournament.ScheduleBlock{startTime, 10, 360})
		cachedMatchType = "practice"
	}
	web.renderSchedule(w, r, "")
}

// Generates the schedule and presents it for review without saving it to the database.
func (web *Web) scheduleGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	r.ParseForm()
	cachedMatchType = r.PostFormValue("matchType")
	scheduleBlocks, err := getScheduleBlocks(r)
	cachedScheduleBlocks = scheduleBlocks // Show the same blocks even if there is an error.
	if err != nil {
		web.renderSchedule(w, r, "Incomplete or invalid schedule block parameters specified.")
		return
	}

	// Build the schedule.
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(teams) == 0 {
		web.renderSchedule(w, r, "No team list is configured. Set up the list of teams at the event before "+
			"generating the schedule.")
		return
	}
	if len(teams) < 18 {
		web.renderSchedule(w, r, fmt.Sprintf("There are only %d teams. There must be at least 18 teams to generate "+
			"a schedule.", len(teams)))
		return
	}
	matches, err := tournament.BuildRandomSchedule(teams, scheduleBlocks, r.PostFormValue("matchType"))
	if err != nil {
		web.renderSchedule(w, r, fmt.Sprintf("Error generating schedule: %s.", err.Error()))
		return
	}
	cachedMatches = matches

	// Determine each team's first match.
	teamFirstMatches := make(map[int]string)
	for _, match := range matches {
		checkTeam := func(team int) {
			_, ok := teamFirstMatches[team]
			if !ok {
				teamFirstMatches[team] = match.DisplayName
			}
		}
		checkTeam(match.Red1)
		checkTeam(match.Red2)
		checkTeam(match.Red3)
		checkTeam(match.Blue1)
		checkTeam(match.Blue2)
		checkTeam(match.Blue3)
	}
	cachedTeamFirstMatches = teamFirstMatches

	http.Redirect(w, r, "/setup/schedule", 303)
}

// Publishes the schedule in the database to TBA
func (web *Web) scheduleRepublishPostHandler(w http.ResponseWriter, r *http.Request) {
	if web.arena.EventSettings.TbaPublishingEnabled {
		// Publish schedule to The Blue Alliance.
		err := web.arena.TbaClient.DeletePublishedMatches()
		if err != nil {
			http.Error(w, "Failed to delete published matches: "+err.Error(), 500)
			return
		}
		err = web.arena.TbaClient.PublishMatches(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish matches: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "TBA publishing is not enabled", 500)
		return
	}

	http.Redirect(w, r, "/setup/schedule", 303)
}

// Saves the generated schedule to the database.
func (web *Web) scheduleSavePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	existingMatches, err := web.arena.Database.GetMatchesByType(cachedMatchType)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(existingMatches) > 0 {
		web.renderSchedule(w, r, fmt.Sprintf("Can't save schedule because a schedule of %d %s matches already "+
			"exists. Clear it first on the Settings page.", len(existingMatches), cachedMatchType))
		return
	}

	for _, match := range cachedMatches {
		err = web.arena.Database.CreateMatch(&match)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	// Back up the database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "post_scheduling")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled && cachedMatchType != "practice" {
		// Publish schedule to The Blue Alliance.
		err = web.arena.TbaClient.DeletePublishedMatches()
		if err != nil {
			http.Error(w, "Failed to delete published matches: "+err.Error(), 500)
			return
		}
		err = web.arena.TbaClient.PublishMatches(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish matches: "+err.Error(), 500)
			return
		}
	}

	http.Redirect(w, r, "/setup/schedule", 303)
}

func (web *Web) renderSchedule(w http.ResponseWriter, r *http.Request, errorMessage string) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	template, err := web.parseFiles("templates/setup_schedule.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		MatchType        string
		ScheduleBlocks   []tournament.ScheduleBlock
		NumTeams         int
		Matches          []model.Match
		TeamFirstMatches map[int]string
		ErrorMessage     string
	}{web.arena.EventSettings, cachedMatchType, cachedScheduleBlocks, len(teams), cachedMatches, cachedTeamFirstMatches,
		errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Converts the post form variables into a slice of schedule blocks.
func getScheduleBlocks(r *http.Request) ([]tournament.ScheduleBlock, error) {
	numScheduleBlocks, err := strconv.Atoi(r.PostFormValue("numScheduleBlocks"))
	if err != nil {
		return []tournament.ScheduleBlock{}, err
	}
	var returnErr error
	scheduleBlocks := make([]tournament.ScheduleBlock, numScheduleBlocks)
	location, _ := time.LoadLocation("Local")
	for i := 0; i < numScheduleBlocks; i++ {
		scheduleBlocks[i].StartTime, err = time.ParseInLocation("2006-01-02 03:04:05 PM",
			r.PostFormValue(fmt.Sprintf("startTime%d", i)), location)
		if err != nil {
			returnErr = err
		}
		scheduleBlocks[i].NumMatches, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("numMatches%d", i)))
		if err != nil {
			returnErr = err
		}
		scheduleBlocks[i].MatchSpacingSec, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("matchSpacingSec%d", i)))
		if err != nil {
			returnErr = err
		}
	}
	return scheduleBlocks, returnErr
}
