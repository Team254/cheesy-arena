// Copyright 2016 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for conducting the team defense selection process.

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"
)

// Shows the defense selection page.
func DefenseSelectionGetHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	renderDefenseSelection(w, r, "")
}

// Updates the cache with the latest input from the client.
func DefenseSelectionPostHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	matchId, _ := strconv.Atoi(r.PostFormValue("matchId"))
	match, err := db.GetMatchById(matchId)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	redErr := validateDefenseSelection([]string{r.PostFormValue("redDefense2"),
		r.PostFormValue("redDefense3"), r.PostFormValue("redDefense4"), r.PostFormValue("redDefense5")})
	if redErr == nil {
		match.RedDefense1 = "LB"
		match.RedDefense2 = r.PostFormValue("redDefense2")
		match.RedDefense3 = r.PostFormValue("redDefense3")
		match.RedDefense4 = r.PostFormValue("redDefense4")
		match.RedDefense5 = r.PostFormValue("redDefense5")
	}
	blueErr := validateDefenseSelection([]string{r.PostFormValue("blueDefense2"),
		r.PostFormValue("blueDefense3"), r.PostFormValue("blueDefense4"), r.PostFormValue("blueDefense5")})
	if blueErr == nil {
		match.BlueDefense1 = "LB"
		match.BlueDefense2 = r.PostFormValue("blueDefense2")
		match.BlueDefense3 = r.PostFormValue("blueDefense3")
		match.BlueDefense4 = r.PostFormValue("blueDefense4")
		match.BlueDefense5 = r.PostFormValue("blueDefense5")
	}
	if redErr == nil || blueErr == nil {
		err = db.SaveMatch(match)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		mainArena.defenseSelectionNotifier.Notify(nil)
	}
	if redErr != nil {
		renderDefenseSelection(w, r, redErr.Error())
		return
	}
	if blueErr != nil {
		renderDefenseSelection(w, r, blueErr.Error())
		return
	}

	http.Redirect(w, r, "/setup/defense_selection", 302)
}

func renderDefenseSelection(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/setup_defense_selection.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	matches, err := db.GetMatchesByType("elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var unplayedMatches []Match
	for _, match := range matches {
		if match.Status != "complete" {
			unplayedMatches = append(unplayedMatches, match)
		}
	}

	data := struct {
		*EventSettings
		Matches      []Match
		DefenseNames map[string]string
		ErrorMessage string
	}{eventSettings, unplayedMatches, defenseNames, errorMessage}
	err = template.ExecuteTemplate(w, "setup_defense_selection.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Takes a slice of the defenses in positions 2-5 and returns an error if they are not valid.
func validateDefenseSelection(defenses []string) error {
	// Build map to track which defenses have been used.
	defenseCounts := make(map[string]int)
	for _, defense := range placeableDefenses {
		defenseCounts[defense] = 0
	}
	numBlankDefenses := 0

	for _, defense := range defenses {
		if defense == "" {
			numBlankDefenses++
			continue
		}

		defenseCount, ok := defenseCounts[defense]
		if !ok {
			return fmt.Errorf("Invalid defense type: %s", defense)
		}
		if defenseCount != 0 {
			return fmt.Errorf("Defense used more than once: %s", defense)
		}
		defenseCounts[defense]++
	}

	if numBlankDefenses > 0 && numBlankDefenses < 4 {
		return fmt.Errorf("Cannot leave defenses blank.")
	}

	return nil
}
