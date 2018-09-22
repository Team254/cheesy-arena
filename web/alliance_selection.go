// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for conducting the alliance selection process.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"net/http"
	"strconv"
	"time"
)

type RankedTeam struct {
	Rank   int
	TeamId int
	Picked bool
}

// Global var to hold the team rankings during the alliance selection.
var cachedRankedTeams []*RankedTeam

// Shows the alliance selection page.
func (web *Web) allianceSelectionGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.renderAllianceSelection(w, r, "")
}

// Updates the cache with the latest input from the client.
func (web *Web) allianceSelectionPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	// Reset picked state for each team in preparation for reconstructing it.
	newRankedTeams := make([]*RankedTeam, len(cachedRankedTeams))
	for i, team := range cachedRankedTeams {
		newRankedTeams[i] = &RankedTeam{team.Rank, team.TeamId, false}
	}

	// Iterate through all selections and update the alliances.
	for i, alliance := range web.arena.AllianceSelectionAlliances {
		for j := range alliance {
			teamString := r.PostFormValue(fmt.Sprintf("selection%d_%d", i, j))
			if teamString == "" {
				web.arena.AllianceSelectionAlliances[i][j].TeamId = 0
			} else {
				teamId, err := strconv.Atoi(teamString)
				if err != nil {
					web.renderAllianceSelection(w, r, fmt.Sprintf("Invalid team number value '%s'.", teamString))
					return
				}
				found := false
				for _, team := range newRankedTeams {
					if team.TeamId == teamId {
						if team.Picked {
							web.renderAllianceSelection(w, r,
								fmt.Sprintf("Team %d is already part of an alliance.", teamId))
							return
						}
						found = true
						team.Picked = true
						web.arena.AllianceSelectionAlliances[i][j].TeamId = teamId
						break
					}
				}
				if !found {
					web.renderAllianceSelection(w, r, fmt.Sprintf("Team %d is not present at this event.", teamId))
					return
				}
			}
		}
	}
	cachedRankedTeams = newRankedTeams

	web.arena.AllianceSelectionNotifier.Notify()
	http.Redirect(w, r, "/alliance_selection", 303)
}

// Sets up the empty alliances and populates the ranked team list.
func (web *Web) allianceSelectionStartHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if len(web.arena.AllianceSelectionAlliances) != 0 {
		web.renderAllianceSelection(w, r, "Can't start alliance selection when it is already in progress.")
		return
	}
	if !web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	// Create a blank alliance set matching the event configuration.
	web.arena.AllianceSelectionAlliances = make([][]model.AllianceTeam, web.arena.EventSettings.NumElimAlliances)
	teamsPerAlliance := 3
	if web.arena.EventSettings.SelectionRound3Order != "" {
		teamsPerAlliance = 4
	}
	for i := 0; i < web.arena.EventSettings.NumElimAlliances; i++ {
		web.arena.AllianceSelectionAlliances[i] = make([]model.AllianceTeam, teamsPerAlliance)
		for j := 0; j < teamsPerAlliance; j++ {
			web.arena.AllianceSelectionAlliances[i][j] = model.AllianceTeam{AllianceId: i + 1, PickPosition: j}
		}
	}

	// Populate the ranked list of teams.
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	cachedRankedTeams = make([]*RankedTeam, len(rankings))
	for i, ranking := range rankings {
		cachedRankedTeams[i] = &RankedTeam{i + 1, ranking.TeamId, false}
	}

	web.arena.AllianceSelectionNotifier.Notify()
	http.Redirect(w, r, "/alliance_selection", 303)
}

// Resets the alliance selection process back to the starting point.
func (web *Web) allianceSelectionResetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	web.arena.AllianceSelectionAlliances = [][]model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.AllianceSelectionNotifier.Notify()
	http.Redirect(w, r, "/alliance_selection", 303)
}

// Saves the selected alliances to the database and generates the first round of elimination matches.
func (web *Web) allianceSelectionFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	location, _ := time.LoadLocation("Local")
	startTime, err := time.ParseInLocation("2006-01-02 03:04:05 PM", r.PostFormValue("startTime"), location)
	if err != nil {
		web.renderAllianceSelection(w, r, "Must specify a valid start time for the playoff rounds.")
		return
	}

	// Check that all spots are filled.
	for _, alliance := range web.arena.AllianceSelectionAlliances {
		for _, team := range alliance {
			if team.TeamId <= 0 {
				web.renderAllianceSelection(w, r, "Can't finalize alliance selection until all spots have been filled.")
				return
			}
		}
	}

	// Save alliances to the database.
	for _, alliance := range web.arena.AllianceSelectionAlliances {
		for _, team := range alliance {
			err := web.arena.Database.CreateAllianceTeam(&team)
			if err != nil {
				handleWebErr(w, err)
				return
			}
		}
	}

	// Generate the first round of elimination matches.
	_, err = tournament.UpdateEliminationSchedule(web.arena.Database, startTime)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Reset yellow cards.
	err = tournament.CalculateTeamCards(web.arena.Database, "elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Back up the database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "post_alliance_selection")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		// Publish alliances and schedule to The Blue Alliance.
		err = web.arena.TbaClient.PublishAlliances(web.arena.Database)
		if err != nil {
			web.renderAllianceSelection(w, r, fmt.Sprintf("Failed to publish alliances: %s", err.Error()))
			return
		}
		err = web.arena.TbaClient.PublishMatches(web.arena.Database)
		if err != nil {
			web.renderAllianceSelection(w, r, fmt.Sprintf("Failed to publish matches: %s", err.Error()))
			return
		}
	}

	http.Redirect(w, r, "/alliance_selection", 303)
}

// Publishes the alliances to the web.
func (web *Web) allianceSelectionPublishHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	err := web.arena.TbaClient.PublishAlliances(web.arena.Database)
	if err != nil {
		http.Error(w, "Failed to publish alliances: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/alliance_selection", 303)
}

func (web *Web) renderAllianceSelection(w http.ResponseWriter, r *http.Request, errorMessage string) {
	if len(web.arena.AllianceSelectionAlliances) == 0 && !web.canModifyAllianceSelection() {
		// The application was restarted since the alliance selection was conducted; reload the alliances from the DB.
		var err error
		web.arena.AllianceSelectionAlliances, err = web.arena.Database.GetAllAlliances()
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	template, err := web.parseFiles("templates/alliance_selection.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	nextRow, nextCol := web.determineNextCell()
	data := struct {
		*model.EventSettings
		Alliances    [][]model.AllianceTeam
		RankedTeams  []*RankedTeam
		NextRow      int
		NextCol      int
		ErrorMessage string
	}{web.arena.EventSettings, web.arena.AllianceSelectionAlliances, cachedRankedTeams, nextRow, nextCol, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the alliance selection (i.e. no elimination matches exist yet).
func (web *Web) canModifyAllianceSelection() bool {
	matches, err := web.arena.Database.GetMatchesByType("elimination")
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns the row and column of the next alliance selection spot that should have keyboard autofocus.
func (web *Web) determineNextCell() (int, int) {
	// Check the first two columns.
	for i, alliance := range web.arena.AllianceSelectionAlliances {
		if alliance[0].TeamId == 0 {
			return i, 0
		}
		if alliance[1].TeamId == 0 {
			return i, 1
		}
	}

	// Check the third column.
	if web.arena.EventSettings.SelectionRound2Order == "F" {
		for i, alliance := range web.arena.AllianceSelectionAlliances {
			if alliance[2].TeamId == 0 {
				return i, 2
			}
		}
	} else {
		for i := len(web.arena.AllianceSelectionAlliances) - 1; i >= 0; i-- {
			if web.arena.AllianceSelectionAlliances[i][2].TeamId == 0 {
				return i, 2
			}
		}
	}

	// Check the fourth column.
	if web.arena.EventSettings.SelectionRound3Order == "F" {
		for i, alliance := range web.arena.AllianceSelectionAlliances {
			if alliance[3].TeamId == 0 {
				return i, 3
			}
		}
	} else if web.arena.EventSettings.SelectionRound3Order == "L" {
		for i := len(web.arena.AllianceSelectionAlliances) - 1; i >= 0; i-- {
			if web.arena.AllianceSelectionAlliances[i][3].TeamId == 0 {
				return i, 3
			}
		}
	}
	return -1, -1
}
