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

// Global vars to hold the alliances that are in the process of being selected.
var cachedAlliances [][]*model.AllianceTeam
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
	for i, alliance := range cachedAlliances {
		for j, spot := range alliance {
			teamString := r.PostFormValue(fmt.Sprintf("selection%d_%d", i, j))
			if teamString == "" {
				spot.TeamId = 0
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
							web.renderAllianceSelection(w, r, fmt.Sprintf("Team %d is already part of an alliance.", teamId))
							return
						}
						found = true
						team.Picked = true
						spot.TeamId = teamId
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

	web.arena.AllianceSelectionNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/alliance_selection", 303)
}

// Sets up the empty alliances and populates the ranked team list.
func (web *Web) allianceSelectionStartHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if len(cachedAlliances) != 0 {
		web.renderAllianceSelection(w, r, "Can't start alliance selection when it is already in progress.")
		return
	}
	if !web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	// Create a blank alliance set matching the event configuration.
	cachedAlliances = make([][]*model.AllianceTeam, web.arena.EventSettings.NumElimAlliances)
	teamsPerAlliance := 3
	if web.arena.EventSettings.SelectionRound3Order != "" {
		teamsPerAlliance = 4
	}
	for i := 0; i < web.arena.EventSettings.NumElimAlliances; i++ {
		cachedAlliances[i] = make([]*model.AllianceTeam, teamsPerAlliance)
		for j := 0; j < teamsPerAlliance; j++ {
			cachedAlliances[i][j] = &model.AllianceTeam{AllianceId: i + 1, PickPosition: j}
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

	web.arena.AllianceSelectionNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/alliance_selection", 303)
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

	cachedAlliances = [][]*model.AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	web.arena.AllianceSelectionNotifier.Notify(nil)
	http.Redirect(w, r, "/setup/alliance_selection", 303)
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
	for _, alliance := range cachedAlliances {
		for _, team := range alliance {
			if team.TeamId <= 0 {
				web.renderAllianceSelection(w, r, "Can't finalize alliance selection until all spots have been filled.")
				return
			}
		}
	}

	// Save alliances to the database.
	for _, alliance := range cachedAlliances {
		for _, team := range alliance {
			err := web.arena.Database.CreateAllianceTeam(team)
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

	http.Redirect(w, r, "/setup/alliance_selection", 303)
}

// Force push alliances to TBA.
func (web *Web) allianceSelectionTbaHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.canModifyAllianceSelection() {
		web.renderAllianceSelection(w, r, "Alliance selection has not yet been finalized.")
		return
	}

	if !web.arena.EventSettings.TbaPublishingEnabled {
		web.renderAllianceSelection(w, r, "The Blue Alliance pushing is not enabled.")
	}

	err := web.arena.TbaClient.PublishAlliances(web.arena.Database)
	if err != nil {
		web.renderAllianceSelection(w, r, fmt.Sprintf("Failed to publish alliances: %s", err.Error()))
		return
	}

	// XXX: This is not an error message
	web.renderAllianceSelection(w, r, "Alliances successfully sent to The Blue Alliance")
}

func (web *Web) renderAllianceSelection(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := web.parseFiles("templates/setup_alliance_selection.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	nextRow, nextCol := web.determineNextCell()
	data := struct {
		*model.EventSettings
		Alliances          [][]*model.AllianceTeam
		RankedTeams        []*RankedTeam
		SelectionFinalized bool
		NextRow            int
		NextCol            int
		ErrorMessage       string
	}{web.arena.EventSettings, cachedAlliances, cachedRankedTeams, !web.canModifyAllianceSelection(), nextRow, nextCol, errorMessage}
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
	for i, alliance := range cachedAlliances {
		if alliance[0].TeamId == 0 {
			return i, 0
		}
		if alliance[1].TeamId == 0 {
			return i, 1
		}
	}

	// Check the third column.
	if web.arena.EventSettings.SelectionRound2Order == "F" {
		for i, alliance := range cachedAlliances {
			if alliance[2].TeamId == 0 {
				return i, 2
			}
		}
	} else {
		for i := len(cachedAlliances) - 1; i >= 0; i-- {
			if cachedAlliances[i][2].TeamId == 0 {
				return i, 2
			}
		}
	}

	// Check the fourth column.
	if web.arena.EventSettings.SelectionRound3Order == "F" {
		for i, alliance := range cachedAlliances {
			if alliance[3].TeamId == 0 {
				return i, 3
			}
		}
	} else if web.arena.EventSettings.SelectionRound3Order == "L" {
		for i := len(cachedAlliances) - 1; i >= 0; i-- {
			if cachedAlliances[i][3].TeamId == 0 {
				return i, 3
			}
		}
	}
	return -1, -1
}
