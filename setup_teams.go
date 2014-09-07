// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package main

import (
	"encoding/csv"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const wpaKeyLength = 8

var officialTeamInfoUrl = "https://my.usfirst.org/frc/scoring/index.lasso?page=teamlist"
var officialTeamInfo map[int][]string

// Shows the team list.
func TeamsGetHandler(w http.ResponseWriter, r *http.Request) {
	renderTeams(w, r, false)
}

// Adds teams to the team list.
func TeamsPostHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
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
		team, err := getOfficialTeamInfo(teamNumber)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		err = db.CreateTeam(team)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}
	http.Redirect(w, r, "/setup/teams", 302)
}

// Clears the team list.
func TeamsClearHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
		return
	}

	err := db.TruncateTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 302)
}

// Shows the page to edit a team's fields.
func TeamEditGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}

	template, err := template.ParseFiles("templates/edit_team.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		*Team
	}{eventSettings, team}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Updates a team's fields.
func TeamEditPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
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
	if eventSettings.NetworkSecurityEnabled {
		team.WpaKey = r.PostFormValue("wpaKey")
		if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
			handleWebErr(w, fmt.Errorf("WPA key must be between 8 and 63 characters."))
			return
		}
	}
	err = db.SaveTeam(team)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 302)
}

// Removes a team from the team list.
func TeamDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyTeamList() {
		renderTeams(w, r, true)
		return
	}

	vars := mux.Vars(r)
	teamId, _ := strconv.Atoi(vars["id"])
	team, err := db.GetTeamById(teamId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if team == nil {
		http.Error(w, fmt.Sprintf("Error: No such team: %d", teamId), 400)
		return
	}
	err = db.DeleteTeam(team)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/teams", 302)
}

// Publishes the team list to the web.
func TeamsPublishHandler(w http.ResponseWriter, r *http.Request) {
	err := PublishTeams()
	if err != nil {
		http.Error(w, "Failed to publish teams: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/setup/teams", 302)
}

// Generates random WPA keys and saves them to the team models.
func TeamsGenerateWpaKeysHandler(w http.ResponseWriter, r *http.Request) {
	generateAllKeys := false
	if all, ok := r.URL.Query()["all"]; ok {
		generateAllKeys = all[0] == "true"
	}

	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	for _, team := range teams {
		if len(team.WpaKey) == 0 || generateAllKeys {
			team.WpaKey = uniuri.NewLen(wpaKeyLength)
			db.SaveTeam(&team)
		}
	}

	http.Redirect(w, r, "/setup/teams", 302)
}

func renderTeams(w http.ResponseWriter, r *http.Request, showErrorMessage bool) {
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := template.ParseFiles("templates/setup_teams.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		Teams            []Team
		ShowErrorMessage bool
	}{eventSettings, teams, showErrorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the team list (i.e. no matches/results exist yet).
func canModifyTeamList() bool {
	matches, err := db.GetMatchesByType("qualification")
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns the data for the given team number.
func getOfficialTeamInfo(teamId int) (*Team, error) {
	if officialTeamInfo == nil && eventSettings.TeamInfoDownloadEnabled {
		// Download all team info from the FIRST website if it is not cached.
		resp, err := http.Get(officialTeamInfoUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		re := regexp.MustCompile("(?s).*<PRE>(.*)</PRE>.*")
		teamsCsv := re.FindStringSubmatch(string(body))[1]

		reader := csv.NewReader(strings.NewReader(teamsCsv))
		reader.Comma = '\t'
		reader.FieldsPerRecord = -1
		officialTeamInfo = make(map[int][]string)
		reader.Read() // Ignore header line.
		for {
			fields, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			teamNumber, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}
			officialTeamInfo[teamNumber] = fields
		}
	}

	teamData, ok := officialTeamInfo[teamId]
	var team Team
	if ok {
		rookieYear, _ := strconv.Atoi(teamData[8])
		team = Team{Id: teamId, Name: html.UnescapeString(teamData[2]), Nickname: html.UnescapeString(teamData[7]),
			City: html.UnescapeString(teamData[4]), StateProv: html.UnescapeString(teamData[5]),
			Country: html.UnescapeString(teamData[6]), RookieYear: rookieYear,
			RobotName: html.UnescapeString(teamData[9])}
	} else {
		// If no team data exists, just fill in the team number.
		team = Team{Id: teamId}
	}
	return &team, nil
}
