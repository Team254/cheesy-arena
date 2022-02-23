// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the event settings.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Shows the event settings editing page.
func (web *Web) settingsGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	web.renderSettings(w, r, "")
}

// Saves the event settings.
func (web *Web) settingsPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	eventSettings := web.arena.EventSettings

	previousEventName := eventSettings.Name
	eventSettings.Name = r.PostFormValue("name")
	if len(eventSettings.Name) < 1 && eventSettings.Name != previousEventName {
		eventSettings.Name = previousEventName
	}
	previousAdminPassword := eventSettings.AdminPassword

	numAlliances, _ := strconv.Atoi(r.PostFormValue("numElimAlliances"))
	if numAlliances < 2 || numAlliances > 16 {
		web.renderSettings(w, r, "Number of alliances must be between 2 and 16.")
		return
	}

	eventSettings.NumElimAlliances = numAlliances
	eventSettings.SelectionRound2Order = r.PostFormValue("selectionRound2Order")
	eventSettings.SelectionRound3Order = r.PostFormValue("selectionRound3Order")
	eventSettings.TBADownloadEnabled = r.PostFormValue("TBADownloadEnabled") == "on"
	eventSettings.TbaPublishingEnabled = r.PostFormValue("tbaPublishingEnabled") == "on"
	eventSettings.TbaEventCode = r.PostFormValue("tbaEventCode")
	eventSettings.TbaSecretId = r.PostFormValue("tbaSecretId")
	eventSettings.TbaSecret = r.PostFormValue("tbaSecret")
	eventSettings.NetworkSecurityEnabled = r.PostFormValue("networkSecurityEnabled") == "on"
	eventSettings.ApAddress = r.PostFormValue("apAddress")
	eventSettings.ApUsername = r.PostFormValue("apUsername")
	eventSettings.ApPassword = r.PostFormValue("apPassword")
	eventSettings.ApTeamChannel, _ = strconv.Atoi(r.PostFormValue("apTeamChannel"))
	eventSettings.ApAdminChannel, _ = strconv.Atoi(r.PostFormValue("apAdminChannel"))
	eventSettings.ApAdminWpaKey = r.PostFormValue("apAdminWpaKey")
	eventSettings.Ap2Address = r.PostFormValue("ap2Address")
	eventSettings.Ap2Username = r.PostFormValue("ap2Username")
	eventSettings.Ap2Password = r.PostFormValue("ap2Password")
	eventSettings.Ap2TeamChannel, _ = strconv.Atoi(r.PostFormValue("ap2TeamChannel"))
	eventSettings.SwitchAddress = r.PostFormValue("switchAddress")
	eventSettings.SwitchPassword = r.PostFormValue("switchPassword")
	eventSettings.PlcAddress = r.PostFormValue("plcAddress")
	eventSettings.AdminPassword = r.PostFormValue("adminPassword")
	eventSettings.WarmupDurationSec, _ = strconv.Atoi(r.PostFormValue("warmupDurationSec"))
	eventSettings.AutoDurationSec, _ = strconv.Atoi(r.PostFormValue("autoDurationSec"))
	eventSettings.PauseDurationSec, _ = strconv.Atoi(r.PostFormValue("pauseDurationSec"))
	eventSettings.TeleopDurationSec, _ = strconv.Atoi(r.PostFormValue("teleopDurationSec"))
	eventSettings.WarningRemainingDurationSec, _ = strconv.Atoi(r.PostFormValue("warningRemainingDurationSec"))
	eventSettings.QuintetThreshold, _ = strconv.Atoi(r.PostFormValue("quintetThreshold"))
	eventSettings.CargoBonusRankingPointThresholdWithoutQuintet, _ =
		strconv.Atoi(r.PostFormValue("cargoBonusRankingPointThresholdWithoutQuintet"))
	eventSettings.CargoBonusRankingPointThresholdWithQuintet, _ =
		strconv.Atoi(r.PostFormValue("cargoBonusRankingPointThresholdWithQuintet"))
	eventSettings.HangarBonusRankingPointThreshold, _ =
		strconv.Atoi(r.PostFormValue("hangarBonusRankingPointThreshold"))

	if eventSettings.Ap2TeamChannel != 0 && eventSettings.Ap2TeamChannel == eventSettings.ApTeamChannel {
		web.renderSettings(w, r, "Cannot use same channel for both access points.")
		return
	}

	err := web.arena.Database.UpdateEventSettings(eventSettings)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Refresh the arena in case any of the settings changed.
	err = web.arena.LoadSettings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	if eventSettings.AdminPassword != previousAdminPassword {
		// Delete any existing user sessions to force a logout.
		if err := web.arena.Database.TruncateUserSessions(); err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Sends a copy of the event database file to the client as a download.
func (web *Web) saveDbHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	filename := fmt.Sprintf("%s-%s.db", strings.Replace(web.arena.EventSettings.Name, " ", "_", -1),
		time.Now().Format("20060102150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := web.arena.Database.WriteBackup(w); err != nil {
		handleWebErr(w, err)
		return
	}
}

// Accepts an event database file as an upload and loads it.
func (web *Web) restoreDbHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	file, _, err := r.FormFile("databaseFile")
	if err != nil {
		web.renderSettings(w, r, "No database backup file was specified.")
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
	tempDb, err := model.OpenDatabase(tempFilePath)
	if err != nil {
		web.renderSettings(w, r, "Could not read uploaded database backup file. Please verify that it a valid "+
			"database file.")
		return
	}
	tempDb.Close()

	// Back up the current database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_restore")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Replace the current database with the new one.
	web.arena.Database.Close()
	err = os.Remove(web.arena.Database.Path)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = os.Rename(tempFilePath, web.arena.Database.Path)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	web.arena.Database, err = model.OpenDatabase(web.arena.Database.Path)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = web.arena.LoadSettings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Deletes all data except for the team list.
func (web *Web) clearDbHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	// Back up the database.
	err := web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_clear")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	err = web.arena.Database.TruncateMatches()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = web.arena.Database.TruncateMatchResults()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = web.arena.Database.TruncateRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = web.arena.Database.TruncateAllianceTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	http.Redirect(w, r, "/setup/settings", 303)
}

func (web *Web) renderSettings(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := web.parseFiles("templates/setup_settings.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		ErrorMessage string
	}{web.arena.EventSettings, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
