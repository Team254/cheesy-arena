// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the event settings.

package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Shows the event settings editing page.
func SettingsGetHandler(w http.ResponseWriter, r *http.Request) {
	renderSettings(w, r, "")
}

// Saves the event settings.
func SettingsPostHandler(w http.ResponseWriter, r *http.Request) {
	eventSettings.Name = r.PostFormValue("name")
	eventSettings.Code = r.PostFormValue("code")
	match, _ := regexp.MatchString("^#([0-9A-Fa-f]{3}){1,2}$", r.PostFormValue("displayBackgroundColor"))
	if !match {
		renderSettings(w, r, "Display background color must be a valid hex color value.")
		return
	}
	eventSettings.DisplayBackgroundColor = r.PostFormValue("displayBackgroundColor")
	numAlliances, _ := strconv.Atoi(r.PostFormValue("numElimAlliances"))
	if numAlliances < 2 || numAlliances > 16 {
		renderSettings(w, r, "Number of alliances must be between 2 and 16.")
		return
	}

	eventSettings.NumElimAlliances = numAlliances
	eventSettings.SelectionRound2Order = r.PostFormValue("selectionRound2Order")
	eventSettings.SelectionRound3Order = r.PostFormValue("selectionRound3Order")
	eventSettings.TeamInfoDownloadEnabled = r.PostFormValue("teamInfoDownloadEnabled") == "on"
	eventSettings.AllianceDisplayHotGoals = r.PostFormValue("allianceDisplayHotGoals") == "on"
	eventSettings.RedGoalLightsAddress = r.PostFormValue("redGoalLightsAddress")
	eventSettings.BlueGoalLightsAddress = r.PostFormValue("blueGoalLightsAddress")
	eventSettings.TbaPublishingEnabled = r.PostFormValue("tbaPublishingEnabled") == "on"
	eventSettings.TbaEventCode = r.PostFormValue("tbaEventCode")
	eventSettings.TbaSecretId = r.PostFormValue("tbaSecretId")
	eventSettings.TbaSecret = r.PostFormValue("tbaSecret")
	eventSettings.NetworkSecurityEnabled = r.PostFormValue("networkSecurityEnabled") == "on"
	eventSettings.ApAddress = r.PostFormValue("apAddress")
	eventSettings.ApUsername = r.PostFormValue("apUsername")
	eventSettings.ApPassword = r.PostFormValue("apPassword")
	eventSettings.SwitchAddress = r.PostFormValue("switchAddress")
	eventSettings.SwitchPassword = r.PostFormValue("switchPassword")
	err := db.SaveEventSettings(eventSettings)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Set up the light controller connections again in case the address changed.
	err = mainArena.lights.SetupConnections()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/settings", 302)
}

// Sends a copy of the event database file to the client as a download.
func SaveDbHandler(w http.ResponseWriter, r *http.Request) {
	dbFile, err := os.Open(db.path)
	defer dbFile.Close()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	filename := fmt.Sprintf("%s-%s.db", strings.Replace(eventSettings.Name, " ", "_", -1),
		time.Now().Format("20060102150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeContent(w, r, "", time.Now(), dbFile)
}

// Accepts an event database file as an upload and loads it.
func RestoreDbHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("databaseFile")
	if err != nil {
		renderSettings(w, r, "No database backup file was specified.")
		return
	}

	// Write the file to a temporary location on disk and verify that it can be opened as a database.
	tempFile, err := ioutil.TempFile(".", "uploaded-db-")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer tempFile.Close()
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath)
	_, err = io.Copy(tempFile, file)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	tempFile.Close()
	tempDb, err := OpenDatabase(tempFilePath)
	if err != nil {
		renderSettings(w, r, "Could not read uploaded database backup file. Please verify that it a valid "+
			"database file.")
		return
	}
	tempDb.Close()

	// Back up the current database.
	err = db.Backup("pre_restore")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Replace the current database with the new one.
	db.Close()
	err = os.Rename(tempFilePath, eventDbPath)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	initDb()

	http.Redirect(w, r, "/setup/settings", 302)
}

// Deletes all data except for the team list.
func ClearDbHandler(w http.ResponseWriter, r *http.Request) {
	// Back up the database.
	err := db.Backup("pre_clear")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	err = db.TruncateMatches()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = db.TruncateMatchResults()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = db.TruncateRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = db.TruncateAllianceTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/settings", 302)
}

func renderSettings(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := template.ParseFiles("templates/settings.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		ErrorMessage string
	}{eventSettings, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
