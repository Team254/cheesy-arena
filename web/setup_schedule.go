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
var cachedMatches = make(map[string][]model.Match)
var cachedTeamFirstMatches = make(map[string]map[int]string)

// Shows the schedule editing page.
func (web *Web) scheduleGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchType := getMatchType(r)
	if matchType == "" {
		http.Redirect(w, r, "/setup/schedule?matchType=practice", 302)
	}

	if matchType != "practice" && matchType != "qualification" {
		handleWebErr(w, fmt.Errorf("Invalid match type '%s'.", matchType))
		return
	}

	web.renderSchedule(w, r, "")
}

// Generates the schedule, presents it for review without saving it, and saves the schedule blocks to the database.
func (web *Web) scheduleGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchType := getMatchType(r)
	useBalancedSchedules := getScheduleBalancing(r)
	teamStrengthsPath := getTeamStrengthsPath(r)
	scheduleBlocks, err := getScheduleBlocks(r)
	// Save blocks even if there is an error, so that any good ones are not discarded.
	deleteBlocksErr := web.arena.Database.DeleteScheduleBlocksByMatchType(matchType)
	if deleteBlocksErr != nil {
		handleWebErr(w, err)
		return
	}
	for _, block := range scheduleBlocks {
		block.MatchType = matchType
		createBlockErr := web.arena.Database.CreateScheduleBlock(&block)
		if createBlockErr != nil {
			handleWebErr(w, err)
			return
		}
	}
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
	matches, err := tournament.BuildRandomSchedule(teams, scheduleBlocks, r.PostFormValue("matchType"), useBalancedSchedules == "enabled", teamStrengthsPath)
	if err != nil {
		web.renderSchedule(w, r, fmt.Sprintf("Error generating schedule: %s.", err.Error()))
		return
	}
	cachedMatches[matchType] = matches

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
	cachedTeamFirstMatches[matchType] = teamFirstMatches

	http.Redirect(w, r, "/setup/schedule?matchType="+matchType, 303)
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

	http.Redirect(w, r, "/setup/schedule?matchType="+getMatchType(r), 303)
}

// Saves the generated schedule to the database.
func (web *Web) scheduleSavePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchType := getMatchType(r)
	existingMatches, err := web.arena.Database.GetMatchesByType(matchType)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(existingMatches) > 0 {
		web.renderSchedule(w, r, fmt.Sprintf("Can't save schedule because a schedule of %d %s matches already "+
			"exists. Clear it first on the Settings page.", len(existingMatches), matchType))
		return
	}

	for _, match := range cachedMatches[matchType] {
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

	if web.arena.EventSettings.TbaPublishingEnabled && matchType != "practice" {
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

	http.Redirect(w, r, "/setup/schedule?matchType="+matchType, 303)
}

func (web *Web) renderSchedule(w http.ResponseWriter, r *http.Request, errorMessage string) {
	matchType := getMatchType(r)
	scheduleBlocks, err := web.arena.Database.GetScheduleBlocksByMatchType(matchType)
	if err != nil {
		handleWebErr(w, err)
		return
	}

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
		ScheduleBlocks   []model.ScheduleBlock
		NumTeams         int
		Matches          []model.Match
		TeamFirstMatches map[int]string
		ErrorMessage     string
	}{web.arena.EventSettings, matchType, scheduleBlocks, len(teams), cachedMatches[matchType],
		cachedTeamFirstMatches[matchType], errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Converts the post form variables into a slice of schedule blocks.
func getScheduleBlocks(r *http.Request) ([]model.ScheduleBlock, error) {
	numScheduleBlocks, err := strconv.Atoi(r.PostFormValue("numScheduleBlocks"))
	if err != nil {
		return []model.ScheduleBlock{}, err
	}
	var returnErr error
	scheduleBlocks := make([]model.ScheduleBlock, numScheduleBlocks)
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

func getMatchType(r *http.Request) string {
	if matchType, ok := r.URL.Query()["matchType"]; ok {
		return matchType[0]
	}
	return r.PostFormValue("matchType")
}

func getScheduleBalancing(r *http.Request) string {
	if balancing, ok := r.URL.Query()["balancing"]; ok {
		return balancing[0]
	}
	return r.PostFormValue("balancing")
}

func getTeamStrengthsPath(r *http.Request) string {
	if teamStrengthsPath, ok := r.URL.Query()["teamStrengthsPath"]; ok {
		return teamStrengthsPath[0]
	}
	return r.PostFormValue("teamStrengthsPath")
}