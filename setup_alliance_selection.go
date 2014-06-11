// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for conducting the alliance selection process.

package main

import (
	"fmt"
	"html/template"
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
var cachedAlliances [][]*AllianceTeam
var cachedRankedTeams []*RankedTeam

// Shows the alliance selection page.
func AllianceSelectionGetHandler(w http.ResponseWriter, r *http.Request) {
	renderAllianceSelection(w, r, "")
}

// Updates the cache with the latest input from the client.
func AllianceSelectionPostHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyAllianceSelection() {
		renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	// Reset picked state for each team in preparation for reconstructing it.
	for _, team := range cachedRankedTeams {
		team.Picked = false
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
					renderAllianceSelection(w, r, fmt.Sprintf("Invalid team number value '%s'.", teamString))
					return
				}
				found := false
				for _, team := range cachedRankedTeams {
					if team.TeamId == teamId {
						if team.Picked {
							renderAllianceSelection(w, r, fmt.Sprintf("Team %d is already part of an alliance.", teamId))
							return
						}
						found = true
						team.Picked = true
						spot.TeamId = teamId
						break
					}
				}
				if !found {
					renderAllianceSelection(w, r, fmt.Sprintf("Team %d is not present at this event.", teamId))
					return
				}
			}
		}
	}
	http.Redirect(w, r, "/setup/alliance_selection", 302)
}

// Sets up the empty alliances and populates the ranked team list.
func AllianceSelectionStartHandler(w http.ResponseWriter, r *http.Request) {
	if len(cachedAlliances) != 0 {
		renderAllianceSelection(w, r, "Can't start alliance selection when it is already in progress.")
		return
	}
	if !canModifyAllianceSelection() {
		renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	// Create a blank alliance set matching the event configuration.
	cachedAlliances = make([][]*AllianceTeam, eventSettings.NumElimAlliances)
	teamsPerAlliance := 3
	if eventSettings.SelectionRound3Order != "" {
		teamsPerAlliance = 4
	}
	for i := 0; i < eventSettings.NumElimAlliances; i++ {
		cachedAlliances[i] = make([]*AllianceTeam, teamsPerAlliance)
		for j := 0; j < teamsPerAlliance; j++ {
			cachedAlliances[i][j] = &AllianceTeam{AllianceId: i + 1, PickPosition: j}
		}
	}

	// Populate the ranked list of teams.
	rankings, err := db.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	cachedRankedTeams = make([]*RankedTeam, len(rankings))
	for i, ranking := range rankings {
		cachedRankedTeams[i] = &RankedTeam{i + 1, ranking.TeamId, false}
	}
	http.Redirect(w, r, "/setup/alliance_selection", 302)
}

// Resets the alliance selection process back to the starting point.
func AllianceSelectionResetHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyAllianceSelection() {
		renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	cachedAlliances = [][]*AllianceTeam{}
	cachedRankedTeams = []*RankedTeam{}
	http.Redirect(w, r, "/setup/alliance_selection", 302)
}

// Saves the selected alliances to the database and generates the first round of elimination matches.
func AllianceSelectionFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	if !canModifyAllianceSelection() {
		renderAllianceSelection(w, r, "Alliance selection has already been finalized.")
		return
	}

	location, _ := time.LoadLocation("Local")
	startTime, err := time.ParseInLocation("2006-01-02 03:04:05 PM", r.PostFormValue("startTime"), location)
	if err != nil {
		renderAllianceSelection(w, r, "Must specify a valid start time for the elimination rounds.")
		return
	}
	matchSpacingSec, err := strconv.Atoi(r.PostFormValue("matchSpacingSec"))
	if err != nil || matchSpacingSec <= 0 {
		renderAllianceSelection(w, r, "Must specify a valid match spacing for the elimination rounds.")
		return
	}

	// Check that all spots are filled.
	for _, alliance := range cachedAlliances {
		for _, team := range alliance {
			if team.TeamId <= 0 {
				renderAllianceSelection(w, r, "Can't finalize alliance selection until all spots have been filled.")
				return
			}
		}
	}

	// Save alliances to the database.
	for _, alliance := range cachedAlliances {
		for _, team := range alliance {
			err := db.CreateAllianceTeam(team)
			if err != nil {
				handleWebErr(w, err)
				return
			}
		}
	}

	// Generate the first round of elimination matches.
	_, err = db.UpdateEliminationSchedule(startTime, matchSpacingSec)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/alliance_selection", 302)
}

func renderAllianceSelection(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := template.ParseFiles("templates/alliance_selection.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	nextRow, nextCol := determineNextCell()
	data := struct {
		*EventSettings
		Alliances    [][]*AllianceTeam
		RankedTeams  []*RankedTeam
		NextRow      int
		NextCol      int
		ErrorMessage string
	}{eventSettings, cachedAlliances, cachedRankedTeams, nextRow, nextCol, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the alliance selection (i.e. no elimination matches exist yet).
func canModifyAllianceSelection() bool {
	matches, err := db.GetMatchesByType("elimination")
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns the row and column of the next alliance selection spot that should have keyboard autofocus.
func determineNextCell() (int, int) {
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
	if eventSettings.SelectionRound2Order == "F" {
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
	if eventSettings.SelectionRound3Order == "F" {
		for i, alliance := range cachedAlliances {
			if alliance[3].TeamId == 0 {
				return i, 3
			}
		}
	} else if eventSettings.SelectionRound3Order == "L" {
		for i := len(cachedAlliances) - 1; i >= 0; i-- {
			if cachedAlliances[i][3].TeamId == 0 {
				return i, 3
			}
		}
	}
	return -1, -1
}
