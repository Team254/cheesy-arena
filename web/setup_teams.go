// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package web

import (
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const wpaKeyLength = 8

// Shows the team list.
func (web *Web) teamsGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.renderTeams(w, r, false)
}

// Adds teams to the team list.
func (web *Web) teamsPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyTeamList() {
		web.renderTeams(w, r, true)
		return
	}

	var teamNumbers []int
	for _, teamNumberString := range strings.Split(r.PostFormValue("teamNumbers"), "\r\n") {
		teamNumber, err := strconv.Atoi(teamNumberString)
		if err == nil {
			teamNumbers = append(teamNumbers, teamNumber)
		}
	}

	for _, teamNumber := range teamNumbers {
		team := model.Team{Id: teamNumber}
		if web.arena.EventSettings.TbaDownloadEnabled {
			if err := web.populateOfficialTeamInfo(&team); err != nil {
				handleWebErr(w, err)
				return
			}
		}
		if err := web.arena.Database.CreateTeam(&team); err != nil {
			handleWebErr(w, err)
			return
		}
	}
	http.Redirect(w, r, "/setup/teams", 303)
}

// Re-downloads the data for all teams from TBA and overwrites any local edits.
func (web *Web) teamsRefreshHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	for _, team := range teams {
		if err = web.populateOfficialTeamInfo(&team); err != nil {
			handleWebErr(w, err)
			return
		}
		if err = web.arena.Database.UpdateTeam(&team); err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/teams", 303)
}

// Clears the team list.
func (web *Web) teamsClearHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyTeamList() {
		web.renderTeams(w, r, true)
		return
	}

	err := web.arena.Database.TruncateTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 303)
}

// Shows the page to edit a team's fields.
func (web *Web) teamEditGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := web.arena.Database.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}

	template, err := web.parseFiles("templates/edit_team.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		*model.Team
	}{web.arena.EventSettings, team}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates a team's fields.
func (web *Web) teamEditPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := web.arena.Database.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}

	team.Name = r.PostFormValue("name")
	team.Nickname = r.PostFormValue("nickname")
	team.City = r.PostFormValue("city")
	team.StateProv = r.PostFormValue("stateProv")
	team.Country = r.PostFormValue("country")
	team.RookieYear, _ = strconv.Atoi(r.PostFormValue("rookieYear"))
	team.RobotName = r.PostFormValue("robotName")
	team.Accomplishments = r.PostFormValue("accomplishments")
	if web.arena.EventSettings.NetworkSecurityEnabled {
		team.WpaKey = r.PostFormValue("wpaKey")
		if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
			handleWebErr(w, fmt.Errorf("WPA key must be between 8 and 63 characters."))
			return
		}
	}
	team.HasConnected = r.PostFormValue("hasConnected") == "on"
	err = web.arena.Database.UpdateTeam(team)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 303)
}

// Removes a team from the team list.
func (web *Web) teamDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyTeamList() {
		web.renderTeams(w, r, true)
		return
	}

	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := web.arena.Database.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}
	err = web.arena.Database.DeleteTeam(team.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 303)
}

// Generates random WPA keys and saves them to the team models.
func (web *Web) teamsGenerateWpaKeysHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	generateAllKeys := false
	if all, ok := r.URL.Query()["all"]; ok {
		generateAllKeys = all[0] == "true"
	}

	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	for _, team := range teams {
		if len(team.WpaKey) == 0 || generateAllKeys {
			team.WpaKey = uniuri.NewLen(wpaKeyLength)
			web.arena.Database.UpdateTeam(&team)
		}
	}

	http.Redirect(w, r, "/setup/teams", 303)
}

func (web *Web) renderTeams(w http.ResponseWriter, r *http.Request, showErrorMessage bool) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/setup_teams.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		Teams            []model.Team
		ShowErrorMessage bool
	}{web.arena.EventSettings, teams, showErrorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the team list (i.e. no matches/results exist yet).
func (web *Web) canModifyTeamList() bool {
	matches, err := web.arena.Database.GetMatchesByType(model.Qualification, true)
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns the data for the given team number.
func (web *Web) populateOfficialTeamInfo(team *model.Team) error {
	tbaTeam, err := web.arena.TbaClient.GetTeam(team.Id)
	if err != nil {
		return err
	}

	// Check if the result is valid. If a team is not found, it will just not have its detail fields filled out.
	if tbaTeam.TeamNumber == 0 {
		return nil
	}

	team.Name = tbaTeam.Name
	team.Nickname = tbaTeam.Nickname
	team.City = tbaTeam.City
	team.StateProv = tbaTeam.StateProv
	team.Country = tbaTeam.Country
	team.RookieYear = tbaTeam.RookieYear
	team.RobotName, err = web.arena.TbaClient.GetRobotName(team.Id, time.Now().Year())
	if err != nil {
		return err
	}

	// Generate string of recent awards in reverse chronological order.
	recentAwards, err := web.arena.TbaClient.GetTeamAwards(team.Id)
	if err != nil {
		return err
	}
	var accomplishmentsBuffer bytes.Buffer
	for i := len(recentAwards) - 1; i >= 0; i-- {
		award := recentAwards[i]
		if time.Now().Year()-award.Year <= 1 {
			accomplishmentsBuffer.WriteString(fmt.Sprintf("<p>%d %s - %s</p>", award.Year, award.EventName,
				award.Name))
		}
	}
	team.Accomplishments = accomplishmentsBuffer.String()

	// Download and store the team's avatar; if there isn't one, ignore the error.
	web.arena.TbaClient.DownloadTeamAvatar(team.Id, time.Now().Year())

	return nil
}
