// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for conducting the alliance selection process.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Global var to hold configurable time limit for selections. A value of zero disables the timer.
var allianceSelectionTimeLimitSec = 0

// Global var to hold a ticker used for the alliance selection timer.
var allianceSelectionTicker *time.Ticker

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

	// Update time limit value.
	allianceSelectionTimeLimitSec, _ = strconv.Atoi(r.PostFormValue("timeLimitSec"))

	// Reset picked state for each team in preparation for reconstructing it.
	for i := range web.arena.AllianceSelectionRankedTeams {
		web.arena.AllianceSelectionRankedTeams[i].Picked = false
	}

	// Iterate through all selections and update the alliances.
	for i, alliance := range web.arena.AllianceSelectionAlliances {
		for j := range alliance.TeamIds {
			teamString := r.PostFormValue(fmt.Sprintf("selection%d_%d", i, j))
			if teamString == "" {
				web.arena.AllianceSelectionAlliances[i].TeamIds[j] = 0
			} else {
				teamId, err := strconv.Atoi(teamString)
				if err != nil {
					web.renderAllianceSelection(w, r, fmt.Sprintf("Invalid team number value '%s'.", teamString))
					return
				}
				found := false
				for k, team := range web.arena.AllianceSelectionRankedTeams {
					if team.TeamId == teamId {
						if team.Picked {
							web.renderAllianceSelection(w, r,
								fmt.Sprintf("Team %d is already part of an alliance.", teamId))
							return
						}
						found = true
						web.arena.AllianceSelectionRankedTeams[k].Picked = true
						web.arena.AllianceSelectionAlliances[i].TeamIds[j] = teamId
						break
					}
				}
				if !found {
					web.renderAllianceSelection(
						w,
						r,
						fmt.Sprintf(
							"Team %d has not played any matches at this event and is ineligible for selection.", teamId,
						),
					)
					return
				}
			}
		}
	}

	if allianceSelectionTicker != nil {
		allianceSelectionTicker.Stop()
		web.arena.AllianceSelectionShowTimer = false
		web.arena.AllianceSelectionTimeRemainingSec = 0
	}
	if _, nextCol := web.determineNextCell(); nextCol > 0 && allianceSelectionTimeLimitSec > 0 {
		web.arena.AllianceSelectionTimeRemainingSec = allianceSelectionTimeLimitSec
		allianceSelectionTicker = time.NewTicker(time.Second)
		go func() {
			for range allianceSelectionTicker.C {
				web.arena.AllianceSelectionTimeRemainingSec--
				web.arena.AllianceSelectionNotifier.Notify()
				if web.arena.AllianceSelectionTimeRemainingSec == 0 {
					allianceSelectionTicker.Stop()
				}
			}
		}()
	}

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
	web.arena.AllianceSelectionAlliances = make([]model.Alliance, web.arena.EventSettings.NumPlayoffAlliances)
	teamsPerAlliance := 3
	if web.arena.EventSettings.SelectionRound3Order != "" {
		teamsPerAlliance = 4
	}
	for i := 0; i < web.arena.EventSettings.NumPlayoffAlliances; i++ {
		web.arena.AllianceSelectionAlliances[i].Id = i + 1
		web.arena.AllianceSelectionAlliances[i].TeamIds = make([]int, teamsPerAlliance)
	}

	// Populate the ranked list of teams.
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	web.arena.AllianceSelectionRankedTeams = make([]model.AllianceSelectionRankedTeam, len(rankings))
	for i, ranking := range rankings {
		web.arena.AllianceSelectionRankedTeams[i] = model.AllianceSelectionRankedTeam{
			Rank:   i + 1,
			TeamId: ranking.TeamId,
			Picked: false,
		}
	}

	web.arena.AllianceSelectionNotifier.Notify()
	http.Redirect(w, r, "/alliance_selection", 303)
}

// Resets the alliance selection process back to the starting point.
func (web *Web) allianceSelectionResetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if !web.canResetAllianceSelection() {
		web.renderAllianceSelection(w, r, "Cannot reset alliance selection; playoff matches have already started.")
		return
	}

	// Delete any playoff matches that were already created (but not played since they would fail the above check).
	err := web.deleteMatchDataForType(model.Playoff)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Delete the saved alliances.
	if err = web.arena.Database.TruncateAlliances(); err != nil {
		handleWebErr(w, err)
		return
	}

	web.arena.AllianceSelectionAlliances = []model.Alliance{}
	web.arena.AllianceSelectionRankedTeams = []model.AllianceSelectionRankedTeam{}
	web.arena.AllianceSelectionNotifier.Notify()
	http.Redirect(w, r, "/alliance_selection", 303)
}

// Saves the selected alliances to the database and generates the first round of playoff matches.
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
		for _, allianceTeamId := range alliance.TeamIds {
			if allianceTeamId <= 0 {
				web.renderAllianceSelection(w, r, "Can't finalize alliance selection until all spots have been filled.")
				return
			}
		}
	}

	// Save alliances to the database.
	for _, alliance := range web.arena.AllianceSelectionAlliances {
		// Populate the initial lineup according to the tournament rules (alliance captain in the middle, first pick on
		// the left, second pick on the right).
		alliance.Lineup[0] = alliance.TeamIds[1]
		alliance.Lineup[1] = alliance.TeamIds[0]
		alliance.Lineup[2] = alliance.TeamIds[2]

		err := web.arena.Database.CreateAlliance(&alliance)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	// Generate the first round of playoff matches.
	if err = web.arena.CreatePlayoffMatches(startTime); err != nil {
		handleWebErr(w, err)
		return
	}

	// Reset yellow cards.
	err = tournament.CalculateTeamCards(web.arena.Database, model.Playoff)
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

	// Signal displays of the bracket to update themselves.
	web.arena.ScorePostedNotifier.Notify()

	// Load the first playoff match.
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, false)
	if err == nil && len(matches) > 0 {
		_ = web.arena.LoadMatch(&matches[0])
	}

	http.Redirect(w, r, "/match_play", 303)
}

// The websocket endpoint for the alliance selection client to send control commands and receive status updates.
func (web *Web) allianceSelectionWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.AllianceSelectionNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "showTimer":
			web.arena.AllianceSelectionShowTimer = true
			web.arena.AllianceSelectionNotifier.Notify()
		case "hideTimer":
			web.arena.AllianceSelectionShowTimer = false
			web.arena.AllianceSelectionNotifier.Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}

func (web *Web) renderAllianceSelection(w http.ResponseWriter, r *http.Request, errorMessage string) {
	if len(web.arena.AllianceSelectionAlliances) == 0 {
		// The application may have been restarted since the alliance selection was conducted; try reloading the
		// alliances from the DB.
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
		Alliances    []model.Alliance
		RankedTeams  []model.AllianceSelectionRankedTeam
		NextRow      int
		NextCol      int
		ErrorMessage string
		TimeLimitSec int
	}{
		web.arena.EventSettings,
		web.arena.AllianceSelectionAlliances,
		web.arena.AllianceSelectionRankedTeams,
		nextRow,
		nextCol,
		errorMessage,
		allianceSelectionTimeLimitSec,
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if it is safe to change the alliance selection (i.e. no playoff matches exist yet).
func (web *Web) canModifyAllianceSelection() bool {
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, true)
	if err != nil || len(matches) > 0 {
		return false
	}
	return true
}

// Returns true if it is safe to reset the alliance selection (i.e. no playoff matches have been played yet).
func (web *Web) canResetAllianceSelection() bool {
	matches, err := web.arena.Database.GetMatchesByType(model.Playoff, true)
	if err != nil {
		return false
	}
	for _, match := range matches {
		if match.IsComplete() {
			return false
		}
	}
	return true
}

// Returns the row and column of the next alliance selection spot that should have keyboard autofocus.
func (web *Web) determineNextCell() (int, int) {
	// Check the first two columns.
	for i, alliance := range web.arena.AllianceSelectionAlliances {
		if alliance.TeamIds[0] == 0 {
			return i, 0
		}
		if alliance.TeamIds[1] == 0 {
			return i, 1
		}
	}

	// Check the third column.
	if web.arena.EventSettings.SelectionRound2Order == "F" {
		for i, alliance := range web.arena.AllianceSelectionAlliances {
			if alliance.TeamIds[2] == 0 {
				return i, 2
			}
		}
	} else {
		for i := len(web.arena.AllianceSelectionAlliances) - 1; i >= 0; i-- {
			if web.arena.AllianceSelectionAlliances[i].TeamIds[2] == 0 {
				return i, 2
			}
		}
	}

	// Check the fourth column.
	if web.arena.EventSettings.SelectionRound3Order == "F" {
		for i, alliance := range web.arena.AllianceSelectionAlliances {
			if alliance.TeamIds[3] == 0 {
				return i, 3
			}
		}
	} else if web.arena.EventSettings.SelectionRound3Order == "L" {
		for i := len(web.arena.AllianceSelectionAlliances) - 1; i >= 0; i-- {
			if web.arena.AllianceSelectionAlliances[i].TeamIds[3] == 0 {
				return i, 3
			}
		}
	}
	return -1, -1
}
