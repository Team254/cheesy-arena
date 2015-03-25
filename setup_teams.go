// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type TeamListings struct {
    NumberOfTeams int `json:"teamCountTotal"`
    Teams []JSONTeam `json:"teams"`
 }

type JSONTeam struct {
        ShortName string `json:"nameShort"`
        TeamNumber string   `json:"teamNumber"`
        LongName string   `json:"nameFull"`
        City string   `json:"city"`
 }

const wpaKeyLength = 8

var officialTeamInfoUrl = "https://frc-api.usfirst.org/api/v1.0/teams/2015"

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
  // Create the team variable that stores the result
  var team Team

  // If team info download is enabled, download the current teams data (caching isn't easy with the new paging system in the api)
	if eventSettings.FMSAPIDownloadEnabled && eventSettings.FMSAPIUsername != "" && eventSettings.FMSAPIAuthKey != "" {
	  // Make an HTTP GET request with basic auth to get the info
		client := &http.Client{}
		var url = officialTeamInfoUrl + "?teamNumber=" + strconv.Itoa(teamId);
		req, err := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(eventSettings.FMSAPIUsername, eventSettings.FMSAPIAuthKey)
    resp, err := client.Do(req)
    
    // Handle any errors
		if err != nil {
			return nil, err
		}
		
		// Get the response and handle errors
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		
		// Parse the response into an object
    var respJSON map[string]interface{}
		json.Unmarshal(body, &respJSON)
				
		// Check if the result is valid.  If a team is not found it won't be allowing us to just return a basic team
		if respJSON == nil {
		  team = Team{Id: teamId}
      return &team, nil
		}
		
		// Break out teams array
		var teams = respJSON["teams"].([]interface{})
		
		// Get the first team returned
    var teamData = teams[0].(map[string]interface {})        	  
    
    // Put all the info into variables (to allow for validation)
    var name string
    var nickname string
    var city string
    var stateProv string
    var country string
    var rookieYear int
    var robotName string
    
    if teamData["nameFull"] != nil { name = teamData["nameFull"].(string) }
    if teamData["nameShort"] != nil { nickname = teamData["nameShort"].(string) }
    if teamData["city"] != nil { city = teamData["city"].(string) }
    if teamData["stateProv"] != nil { stateProv = teamData["stateProv"].(string) }
    if teamData["country"] != nil { country = teamData["country"].(string) }
    if teamData["rookieYear"] != nil { rookieYear = int(teamData["rookieYear"].(float64)) }
    if teamData["robotName"] != nil { robotName = teamData["robotName"].(string) }
    
    // Use those variables to make a team object
	  team = Team{Id: teamId, Name: name, Nickname: nickname,
			City: city, StateProv: stateProv,
			Country: country, RookieYear: rookieYear,
			RobotName: robotName}
	} else {
	  // If team grab is disabled, just use the team number
	  team = Team{Id: teamId}
	}
	
	// Return the team object
	return &team, nil
}
