// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the event settings.

package web

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Team254/cheesy-arena/model"
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

	var playoffType model.PlayoffType
	numAlliances := 0
	if r.PostFormValue("playoffType") == "SingleEliminationPlayoff" {
		playoffType = model.SingleEliminationPlayoff
		numAlliances, _ = strconv.Atoi(r.PostFormValue("numPlayoffAlliances"))
		if numAlliances < 2 || numAlliances > 16 {
			web.renderSettings(w, r, "Number of alliances must be between 2 and 16.")
			return
		}
	} else {
		playoffType = model.DoubleEliminationPlayoff
		numAlliances = 8
	}
	if eventSettings.PlayoffType != playoffType || eventSettings.NumPlayoffAlliances != numAlliances {
		alliances, err := web.arena.Database.GetAllAlliances()
		if err != nil {
			handleWebErr(w, err)
			return
		}
		if len(alliances) > 0 {
			web.renderSettings(
				w, r, "Cannot change playoff type or size after alliance selection has been finalized.",
			)
			return
		}
	}
	eventSettings.PlayoffType = playoffType

	eventSettings.NumPlayoffAlliances = numAlliances
	eventSettings.SelectionRound2Order = r.PostFormValue("selectionRound2Order")
	eventSettings.SelectionRound3Order = r.PostFormValue("selectionRound3Order")
	eventSettings.SelectionShowUnpickedTeams = r.PostFormValue("selectionShowUnpickedTeams") == "on"
	eventSettings.TbaDownloadEnabled = r.PostFormValue("tbaDownloadEnabled") == "on"
	eventSettings.TbaPublishingEnabled = r.PostFormValue("tbaPublishingEnabled") == "on"
	eventSettings.TbaEventCode = r.PostFormValue("tbaEventCode")
	eventSettings.TbaSecretId = r.PostFormValue("tbaSecretId")
	eventSettings.TbaSecret = r.PostFormValue("tbaSecret")
	eventSettings.NexusEnabled = r.PostFormValue("nexusEnabled") == "on"
	eventSettings.NetworkSecurityEnabled = r.PostFormValue("networkSecurityEnabled") == "on"
	eventSettings.ApAddress = r.PostFormValue("apAddress")
	eventSettings.ApPassword = r.PostFormValue("apPassword")
	eventSettings.ApChannel, _ = strconv.Atoi(r.PostFormValue("apChannel"))
	eventSettings.SwitchAddress = r.PostFormValue("switchAddress")
	eventSettings.SwitchPassword = r.PostFormValue("switchPassword")
	eventSettings.SCCManagementEnabled = r.PostFormValue("sccManagementEnabled") == "on"
	eventSettings.RedSCCAddress = r.PostFormValue("redSCCAddress")
	eventSettings.BlueSCCAddress = r.PostFormValue("blueSCCAddress")
	eventSettings.SCCUsername = r.PostFormValue("sccUsername")
	eventSettings.SCCPassword = r.PostFormValue("sccPassword")
	eventSettings.SCCUpCommands = r.PostFormValue("sccUpCommands")
	eventSettings.SCCDownCommands = r.PostFormValue("sccDownCommands")
	eventSettings.PlcAddress = r.PostFormValue("plcAddress")
	eventSettings.AdminPassword = r.PostFormValue("adminPassword")
	eventSettings.TeamSignRed1Id, _ = strconv.Atoi(r.PostFormValue("teamSignRed1Id"))
	eventSettings.TeamSignRed2Id, _ = strconv.Atoi(r.PostFormValue("teamSignRed2Id"))
	eventSettings.TeamSignRed3Id, _ = strconv.Atoi(r.PostFormValue("teamSignRed3Id"))
	eventSettings.TeamSignRedTimerId, _ = strconv.Atoi(r.PostFormValue("teamSignRedTimerId"))
	eventSettings.TeamSignBlue1Id, _ = strconv.Atoi(r.PostFormValue("teamSignBlue1Id"))
	eventSettings.TeamSignBlue2Id, _ = strconv.Atoi(r.PostFormValue("teamSignBlue2Id"))
	eventSettings.TeamSignBlue3Id, _ = strconv.Atoi(r.PostFormValue("teamSignBlue3Id"))
	eventSettings.TeamSignBlueTimerId, _ = strconv.Atoi(r.PostFormValue("teamSignBlueTimerId"))
	eventSettings.UseLiteUdpPort = r.PostFormValue("useLiteUdpPort") == "on"
	eventSettings.BlackmagicAddresses = r.PostFormValue("blackmagicAddresses")
	eventSettings.WarmupDurationSec, _ = strconv.Atoi(r.PostFormValue("warmupDurationSec"))
	eventSettings.AutoDurationSec, _ = strconv.Atoi(r.PostFormValue("autoDurationSec"))
	eventSettings.PauseDurationSec, _ = strconv.Atoi(r.PostFormValue("pauseDurationSec"))
	eventSettings.TeleopDurationSec, _ = strconv.Atoi(r.PostFormValue("teleopDurationSec"))
	eventSettings.WarningRemainingDurationSec, _ = strconv.Atoi(r.PostFormValue("warningRemainingDurationSec"))
	eventSettings.AutoBonusCoralThreshold, _ = strconv.Atoi(r.PostFormValue("autoBonusCoralThreshold"))
	eventSettings.CoralBonusPerLevelThreshold, _ = strconv.Atoi(r.PostFormValue("coralBonusPerLevelThreshold"))
	eventSettings.CoralBonusCoopEnabled = r.PostFormValue("coralBonusCoopEnabled") == "on"
	eventSettings.BargeBonusPointThreshold, _ = strconv.Atoi(r.PostFormValue("bargeBonusPointThreshold"))

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

	filename := fmt.Sprintf(
		"%s-%s.db", strings.Replace(web.arena.EventSettings.Name, " ", "_", -1), time.Now().Format("20060102150405"),
	)
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
		web.renderSettings(
			w, r, "Could not read uploaded database backup file. Please verify that it a valid database file.",
		)
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

// Deletes all match data including and beyond the given tournament stage.
func (web *Web) clearDbHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchType, err := model.MatchTypeFromString(r.PathValue("type"))
	if err != nil || matchType == model.Test {
		web.renderSettings(w, r, "Invalid tournament stage to clear.")
		return

	}

	// Back up the database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_clear")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	switch matchType {
	case model.Practice:
		if err = web.deleteMatchDataForType(model.Practice); err != nil {
			handleWebErr(w, err)
			return
		}
	case model.Qualification:
		if err = web.deleteMatchDataForType(model.Qualification); err != nil {
			handleWebErr(w, err)
			return
		}
		if err = web.arena.Database.TruncateRankings(); err != nil {
			handleWebErr(w, err)
			return
		}
	case model.Playoff:
		if err = web.deleteMatchDataForType(model.Playoff); err != nil {
			handleWebErr(w, err)
			return
		}
		if err = web.arena.Database.TruncateAlliances(); err != nil {
			handleWebErr(w, err)
			return
		}
		web.arena.AllianceSelectionAlliances = []model.Alliance{}
		web.arena.AllianceSelectionRankedTeams = []model.AllianceSelectionRankedTeam{}
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Publishes the playoff alliances to the web.
func (web *Web) settingsPublishAlliancesHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishAlliances(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish alliances: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "TBA publishing is not enabled", 500)
		return
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Publishes the awards to the web.
func (web *Web) settingsPublishAwardsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishAwards(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish awards: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "TBA publishing is not enabled", 500)
		return
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Publishes the match schedule and results to the web.
func (web *Web) settingsPublishMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
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

	http.Redirect(w, r, "/setup/settings", 303)
}

// Publishes the standings to the web.
func (web *Web) settingsPublishRankingsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishRankings(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish rankings: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "TBA publishing is not enabled", 500)
		return
	}

	http.Redirect(w, r, "/setup/settings", 303)
}

// Publishes the team list to the web.
func (web *Web) settingsPublishTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishTeams(web.arena.Database)
		if err != nil {
			http.Error(w, "Failed to publish teams: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "TBA publishing is not enabled", 500)
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

// Deletes all match data (matches, results, and scheduled breaks) for the given match type.
func (web *Web) deleteMatchDataForType(matchType model.MatchType) error {
	matches, err := web.arena.Database.GetMatchesByType(matchType, true)
	if err != nil {
		return err
	}
	for _, match := range matches {
		// Loop to delete all match results for the match before deleting the match itself.
		matchResult, err := web.arena.Database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		for matchResult != nil {
			if err = web.arena.Database.DeleteMatchResult(matchResult.Id); err != nil {
				return err
			}
			matchResult, err = web.arena.Database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
		}

		if err = web.arena.Database.DeleteMatch(match.Id); err != nil {
			return err
		}
	}
	if err = web.arena.Database.DeleteScheduledBreaksByMatchType(matchType); err != nil {
		return err
	}
	return nil
}
