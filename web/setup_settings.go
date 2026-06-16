// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the event settings.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	activeSettingsTab := settingsTabFromRequest(r)
	if !settingsSaveAllowed(web.arena.MatchState) {
		web.renderSettingsWithStatus(
			w, r, "Settings cannot be changed while a match is in progress or is uncommitted.", activeSettingsTab,
			http.StatusOK,
		)
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
	playoffTypeValue := r.PostFormValue("playoffType")
	playoffTypeProvided := playoffTypeValue != ""
	if playoffTypeValue == "" && eventSettings.PlayoffType == model.SingleEliminationPlayoff {
		playoffTypeValue = "SingleEliminationPlayoff"
	}
	if playoffTypeValue == "SingleEliminationPlayoff" || playoffTypeValue == "single" {
		playoffType = model.SingleEliminationPlayoff
		if r.PostFormValue("numPlayoffAlliances") == "" {
			if playoffTypeProvided {
				numAlliances = 0
			} else {
				numAlliances = eventSettings.NumPlayoffAlliances
			}
		} else {
			numAlliances, _ = strconv.Atoi(r.PostFormValue("numPlayoffAlliances"))
		}
		if numAlliances < 2 || numAlliances > 16 {
			web.renderSettingsWithStatus(w, r, "Number of alliances must be between 2 and 16.", activeSettingsTab, http.StatusOK)
			return
		}
	} else {
		playoffType = model.DoubleEliminationPlayoff
		if r.PostFormValue("numPlayoffAlliances") == "" {
			if eventSettings.PlayoffType == model.DoubleEliminationPlayoff {
				numAlliances = eventSettings.NumPlayoffAlliances
			} else {
				numAlliances = 8
			}
		} else {
			numAlliances, _ = strconv.Atoi(r.PostFormValue("numPlayoffAlliances"))
		}
		if numAlliances != 4 && numAlliances != 8 {
			web.renderSettingsWithStatus(
				w, r, "Number of alliances for double elimination must be 4 or 8.", activeSettingsTab, http.StatusOK,
			)
			return
		}
	}
	if eventSettings.PlayoffType != playoffType || eventSettings.NumPlayoffAlliances != numAlliances {
		alliances, err := web.arena.Database.GetAllAlliances()
		if err != nil {
			handleWebErr(w, err)
			return
		}
		if len(alliances) > 0 {
			web.renderSettingsWithStatus(
				w, r, "Cannot change playoff type or size after alliance selection has been finalized.", activeSettingsTab,
				http.StatusOK,
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
	eventSettings.AutoAudienceDisplayEnabled = r.PostFormValue("autoAudienceDisplayEnabled") == "on"
	eventSettings.NexusEnabled = r.PostFormValue("nexusEnabled") == "on"
	eventSettings.NexusAutoQueueEnabled = r.PostFormValue("nexusAutoQueueEnabled") == "on"
	eventSettings.NexusAutoQueueKey = r.PostFormValue("nexusAutoQueueKey")
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
	eventSettings.LedControllerAddress = r.PostFormValue("ledControllerAddress")
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
	eventSettings.CompanionAddress = r.PostFormValue("companionAddress")
	eventSettings.CompanionPort, _ = strconv.Atoi(r.PostFormValue("companionPort"))
	eventSettings.CompanionMatchPreviewPage, _ = strconv.Atoi(r.PostFormValue("companionMatchPreviewPage"))
	eventSettings.CompanionMatchPreviewRow, _ = strconv.Atoi(r.PostFormValue("companionMatchPreviewRow"))
	eventSettings.CompanionMatchPreviewColumn, _ = strconv.Atoi(r.PostFormValue("companionMatchPreviewColumn"))
	eventSettings.CompanionSetAudiencePage, _ = strconv.Atoi(r.PostFormValue("companionSetAudiencePage"))
	eventSettings.CompanionSetAudienceRow, _ = strconv.Atoi(r.PostFormValue("companionSetAudienceRow"))
	eventSettings.CompanionSetAudienceColumn, _ = strconv.Atoi(r.PostFormValue("companionSetAudienceColumn"))
	eventSettings.CompanionMatchStartPage, _ = strconv.Atoi(r.PostFormValue("companionMatchStartPage"))
	eventSettings.CompanionMatchStartRow, _ = strconv.Atoi(r.PostFormValue("companionMatchStartRow"))
	eventSettings.CompanionMatchStartColumn, _ = strconv.Atoi(r.PostFormValue("companionMatchStartColumn"))
	eventSettings.CompanionTeleopStartPage, _ = strconv.Atoi(r.PostFormValue("companionTeleopStartPage"))
	eventSettings.CompanionTeleopStartRow, _ = strconv.Atoi(r.PostFormValue("companionTeleopStartRow"))
	eventSettings.CompanionTeleopStartColumn, _ = strconv.Atoi(r.PostFormValue("companionTeleopStartColumn"))
	eventSettings.CompanionEndgameStartPage, _ = strconv.Atoi(r.PostFormValue("companionEndgameStartPage"))
	eventSettings.CompanionEndgameStartRow, _ = strconv.Atoi(r.PostFormValue("companionEndgameStartRow"))
	eventSettings.CompanionEndgameStartColumn, _ = strconv.Atoi(r.PostFormValue("companionEndgameStartColumn"))
	eventSettings.CompanionMatchEndPage, _ = strconv.Atoi(r.PostFormValue("companionMatchEndPage"))
	eventSettings.CompanionMatchEndRow, _ = strconv.Atoi(r.PostFormValue("companionMatchEndRow"))
	eventSettings.CompanionMatchEndColumn, _ = strconv.Atoi(r.PostFormValue("companionMatchEndColumn"))
	eventSettings.CompanionPostResultPage, _ = strconv.Atoi(r.PostFormValue("companionPostResultPage"))
	eventSettings.CompanionPostResultRow, _ = strconv.Atoi(r.PostFormValue("companionPostResultRow"))
	eventSettings.CompanionPostResultColumn, _ = strconv.Atoi(r.PostFormValue("companionPostResultColumn"))
	eventSettings.CompanionAllianceSelectionPage, _ = strconv.Atoi(r.PostFormValue("companionAllianceSelectionPage"))
	eventSettings.CompanionAllianceSelectionRow, _ = strconv.Atoi(r.PostFormValue("companionAllianceSelectionRow"))
	eventSettings.CompanionAllianceSelectionColumn, _ = strconv.Atoi(r.PostFormValue("companionAllianceSelectionColumn"))
	eventSettings.CompanionMatchAbortPage, _ = strconv.Atoi(r.PostFormValue("companionMatchAbortPage"))
	eventSettings.CompanionMatchAbortRow, _ = strconv.Atoi(r.PostFormValue("companionMatchAbortRow"))
	eventSettings.CompanionMatchAbortColumn, _ = strconv.Atoi(r.PostFormValue("companionMatchAbortColumn"))
	eventSettings.AutoDurationSec, _ = strconv.Atoi(r.PostFormValue("autoDurationSec"))
	eventSettings.PauseDurationSec, _ = strconv.Atoi(r.PostFormValue("pauseDurationSec"))
	eventSettings.TransitionShiftDurationSec, _ = strconv.Atoi(r.PostFormValue("transitionShiftDurationSec"))
	eventSettings.ShiftDurationSec, _ = strconv.Atoi(r.PostFormValue("shiftDurationSec"))
	eventSettings.EndgameDurationSec, _ = strconv.Atoi(r.PostFormValue("endgameDurationSec"))
	eventSettings.EnergizedBonusThreshold, _ = strconv.Atoi(r.PostFormValue("energizedBonusThreshold"))
	eventSettings.SuperchargedBonusThreshold, _ = strconv.Atoi(r.PostFormValue("superchargedBonusThreshold"))
	eventSettings.TraversalBonusThreshold, _ = strconv.Atoi(r.PostFormValue("traversalBonusThreshold"))

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

	http.Redirect(w, r, "/setup/settings#"+activeSettingsTab, 303)
}

func settingsSaveAllowed(matchState field.MatchState) bool {
	return matchState == field.PreMatch || matchState == field.TimeoutActive || matchState == field.PostTimeout
}

func settingsTabFromRequest(r *http.Request) string {
	switch r.PostFormValue("activeSettingsTab") {
	case "event", "game", "field", "publishing", "automation":
		return r.PostFormValue("activeSettingsTab")
	default:
		return "event"
	}
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
	defer file.Close()

	// Write the file to a temporary location on disk and verify that it can be opened as a database.
	tempFile, err := os.CreateTemp(filepath.Dir(web.arena.Database.Path), "uploaded-db-")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	tempFilePath := tempFile.Name()
	defer func() {
		if tempFilePath == "" {
			return
		}
		if err := os.Remove(tempFilePath); err != nil {
			log.Printf("Failed to remove temporary uploaded database file %s: %v", tempFilePath, err)
		}
	}()
	_, err = io.Copy(tempFile, file)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if err = tempFile.Close(); err != nil {
		handleWebErr(w, err)
		return
	}
	tempDb, err := model.OpenDatabase(tempFilePath)
	if err != nil {
		web.renderSettings(
			w, r, "Could not read uploaded database backup file. Please verify that it a valid database file.",
		)
		return
	}
	if err = tempDb.Close(); err != nil {
		handleWebErr(w, err)
		return
	}

	// Back up the current database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_restore")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Replace the current database with the new one.
	if err = web.arena.Database.Close(); err != nil {
		handleWebErr(w, err)
		return
	}
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
	tempFilePath = ""
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
			web.renderSettingsWithStatus(
				w, r, "Failed to publish alliances: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
	} else {
		web.renderSettingsWithStatus(w, r, "TBA publishing is not enabled", "publishing", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/setup/settings#publishing", 303)
}

// Publishes the awards to the web.
func (web *Web) settingsPublishAwardsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishAwards(web.arena.Database)
		if err != nil {
			web.renderSettingsWithStatus(
				w, r, "Failed to publish awards: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
	} else {
		web.renderSettingsWithStatus(w, r, "TBA publishing is not enabled", "publishing", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/setup/settings#publishing", 303)
}

// Publishes the match schedule and results to the web.
func (web *Web) settingsPublishMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.DeletePublishedMatches()
		if err != nil {
			web.renderSettingsWithStatus(
				w, r, "Failed to delete published matches: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
		err = web.arena.TbaClient.PublishMatches(web.arena.Database)
		if err != nil {
			web.renderSettingsWithStatus(
				w, r, "Failed to publish matches: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
	} else {
		web.renderSettingsWithStatus(w, r, "TBA publishing is not enabled", "publishing", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/setup/settings#publishing", 303)
}

// Publishes the standings to the web.
func (web *Web) settingsPublishRankingsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishRankings(web.arena.Database)
		if err != nil {
			web.renderSettingsWithStatus(
				w, r, "Failed to publish rankings: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
	} else {
		web.renderSettingsWithStatus(w, r, "TBA publishing is not enabled", "publishing", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/setup/settings#publishing", 303)
}

// Publishes the team list to the web.
func (web *Web) settingsPublishTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if web.arena.EventSettings.TbaPublishingEnabled {
		err := web.arena.TbaClient.PublishTeams(web.arena.Database)
		if err != nil {
			web.renderSettingsWithStatus(
				w, r, "Failed to publish teams: "+err.Error(), "publishing", http.StatusInternalServerError,
			)
			return
		}
	} else {
		web.renderSettingsWithStatus(w, r, "TBA publishing is not enabled", "publishing", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/setup/settings#publishing", 303)
}

func (web *Web) renderSettings(w http.ResponseWriter, r *http.Request, errorMessage string) {
	web.renderSettingsWithStatus(w, r, errorMessage, "event", http.StatusOK)
}

func (web *Web) renderSettingsWithStatus(
	w http.ResponseWriter, r *http.Request, errorMessage string, activeSettingsTab string, statusCode int,
) {
	template, err := web.parseFiles("templates/setup_settings.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		ErrorMessage      string
		ActiveSettingsTab string
		NexusBaseUrl      string
	}{web.arena.EventSettings, errorMessage, activeSettingsTab, web.arena.NexusClient.BaseUrl}
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
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
