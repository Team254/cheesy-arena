// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Configuration and functions for the event server web interface.

package web

import (
	"bitbucket.org/rj/httpauth-go"
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

const (
	adminUser  = "admin"
	readerUser = "reader"
)

type Web struct {
	arena           *field.Arena
	cookieAuth      *httpauth.Cookie
	templateHelpers template.FuncMap
}

func NewWeb(arena *field.Arena) *Web {
	web := &Web{arena: arena}
	web.cookieAuth = httpauth.NewCookie("Cheesy Arena", "", web.checkAuthPassword)

	// Helper functions that can be used inside templates.
	web.templateHelpers = template.FuncMap{
		// Allows sub-templates to be invoked with multiple arguments.
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("Invalid dict call.")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("Dict keys must be strings.")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	return web
}

// ServeWebInterface starts the webserver and blocks, waiting on requests. Does not return until the application exits.
func (web *Web) ServeWebInterface(port int) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/", web.newHandler())
	log.Printf("Serving HTTP requests on port %d", port)

	// Start Server
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// Serves the root page of Cheesy Arena.
func (web *Web) indexHandler(w http.ResponseWriter, r *http.Request) {
	template, err := web.parseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Returns true if the given user is authorized for admin operations. Used for HTTP cookie authentication.
func (web *Web) userIsAdmin(w http.ResponseWriter, r *http.Request) bool {
	if web.arena.EventSettings.AdminPassword == "" {
		// Disable auth if there is no password configured.
		return true
	}
	if web.cookieAuth.Authorize(r) == adminUser {
		return true
	} else {
		http.Redirect(w, r, "/login?redirect="+r.URL.Path, 307)
		return false
	}
}

// Returns true if the given user is authorized for read-only operations. Used for HTTP cookie authentication.
func (web *Web) userIsReader(w http.ResponseWriter, r *http.Request) bool {
	if web.arena.EventSettings.ReaderPassword == "" {
		// Disable auth if there is no password configured.
		return true
	}
	if username := web.cookieAuth.Authorize(r); username == readerUser || username == adminUser {
		return true
	} else {
		http.Redirect(w, r, "/login?redirect="+r.URL.Path, 307)
		return false
	}
}

func (web *Web) checkAuthPassword(user, password string) bool {
	switch user {
	case adminUser:
		return password == web.arena.EventSettings.AdminPassword
	case readerUser:
		return password == web.arena.EventSettings.ReaderPassword
	default:
		return false
	}
}

// Sets up the mapping between URLs and handlers.
func (web *Web) newHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", web.indexHandler).Methods("GET")
	router.HandleFunc("/alliance_selection", web.allianceSelectionGetHandler).Methods("GET")
	router.HandleFunc("/alliance_selection", web.allianceSelectionPostHandler).Methods("POST")
	router.HandleFunc("/alliance_selection/finalize", web.allianceSelectionFinalizeHandler).Methods("POST")
	router.HandleFunc("/alliance_selection/publish", web.allianceSelectionPublishHandler).Methods("POST")
	router.HandleFunc("/alliance_selection/reset", web.allianceSelectionResetHandler).Methods("POST")
	router.HandleFunc("/alliance_selection/start", web.allianceSelectionStartHandler).Methods("POST")
	router.HandleFunc("/api/alliances", web.alliancesApiHandler).Methods("GET")
	router.HandleFunc("/api/matches/{type}", web.matchesApiHandler).Methods("GET")
	router.HandleFunc("/api/rankings", web.rankingsApiHandler).Methods("GET")
	router.HandleFunc("/api/sponsor_slides", web.sponsorSlidesApiHandler).Methods("GET")
	router.HandleFunc("/display", web.placeholderDisplayHandler).Methods("GET")
	router.HandleFunc("/display/websocket", web.placeholderDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/alliance_station", web.allianceStationDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/alliance_station/websocket", web.allianceStationDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/announcer", web.announcerDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/announcer/websocket", web.announcerDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/audience", web.audienceDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/audience/websocket", web.audienceDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/field_monitor", web.fieldMonitorDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/field_monitor/websocket", web.fieldMonitorDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/pit", web.pitDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/pit/websocket", web.pitDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/queueing", web.queueingDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/queueing/websocket", web.queueingDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/displays/twitch", web.twitchDisplayHandler).Methods("GET")
	router.HandleFunc("/displays/twitch/websocket", web.twitchDisplayWebsocketHandler).Methods("GET")
	router.HandleFunc("/login", web.loginHandler).Methods("GET")
	router.HandleFunc("/login", web.loginPostHandler).Methods("POST")
	router.HandleFunc("/match_play", web.matchPlayHandler).Methods("GET")
	router.HandleFunc("/match_play/{matchId}/load", web.matchPlayLoadHandler).Methods("GET")
	router.HandleFunc("/match_play/{matchId}/show_result", web.matchPlayShowResultHandler).Methods("GET")
	router.HandleFunc("/match_play/websocket", web.matchPlayWebsocketHandler).Methods("GET")
	router.HandleFunc("/match_review", web.matchReviewHandler).Methods("GET")
	router.HandleFunc("/match_review/{matchId}/edit", web.matchReviewEditGetHandler).Methods("GET")
	router.HandleFunc("/match_review/{matchId}/edit", web.matchReviewEditPostHandler).Methods("POST")
	router.HandleFunc("/panels/scoring/{alliance}", web.scoringPanelHandler).Methods("GET")
	router.HandleFunc("/panels/scoring/{alliance}/websocket", web.scoringPanelWebsocketHandler).Methods("GET")
	router.HandleFunc("/panels/referee", web.refereePanelHandler).Methods("GET")
	router.HandleFunc("/panels/referee/websocket", web.refereePanelWebsocketHandler).Methods("GET")
	router.HandleFunc("/reports/csv/rankings", web.rankingsCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/rankings", web.rankingsPdfReportHandler).Methods("GET")
	router.HandleFunc("/reports/csv/schedule/{type}", web.scheduleCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/schedule/{type}", web.schedulePdfReportHandler).Methods("GET")
	router.HandleFunc("/reports/csv/teams", web.teamsCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/teams", web.teamsPdfReportHandler).Methods("GET")
	router.HandleFunc("/reports/csv/wpa_keys", web.wpaKeysCsvReportHandler).Methods("GET")
	router.HandleFunc("/setup/db/clear", web.clearDbHandler).Methods("POST")
	router.HandleFunc("/setup/db/restore", web.restoreDbHandler).Methods("POST")
	router.HandleFunc("/setup/db/save", web.saveDbHandler).Methods("GET")
	router.HandleFunc("/setup/displays", web.displaysGetHandler).Methods("GET")
	router.HandleFunc("/setup/displays/websocket", web.displaysWebsocketHandler).Methods("GET")
	router.HandleFunc("/setup/led_plc", web.ledPlcGetHandler).Methods("GET")
	router.HandleFunc("/setup/led_plc/websocket", web.ledPlcWebsocketHandler).Methods("GET")
	router.HandleFunc("/setup/lower_thirds", web.lowerThirdsGetHandler).Methods("GET")
	router.HandleFunc("/setup/lower_thirds/websocket", web.lowerThirdsWebsocketHandler).Methods("GET")
	router.HandleFunc("/setup/schedule", web.scheduleGetHandler).Methods("GET")
	router.HandleFunc("/setup/schedule/generate", web.scheduleGeneratePostHandler).Methods("POST")
	router.HandleFunc("/setup/schedule/republish", web.scheduleRepublishPostHandler).Methods("POST")
	router.HandleFunc("/setup/schedule/save", web.scheduleSavePostHandler).Methods("POST")
	router.HandleFunc("/setup/settings", web.settingsGetHandler).Methods("GET")
	router.HandleFunc("/setup/settings", web.settingsPostHandler).Methods("POST")
	router.HandleFunc("/setup/sponsor_slides", web.sponsorSlidesGetHandler).Methods("GET")
	router.HandleFunc("/setup/sponsor_slides", web.sponsorSlidesPostHandler).Methods("POST")
	router.HandleFunc("/setup/teams", web.teamsGetHandler).Methods("GET")
	router.HandleFunc("/setup/teams", web.teamsPostHandler).Methods("POST")
	router.HandleFunc("/setup/teams/{id}/delete", web.teamDeletePostHandler).Methods("POST")
	router.HandleFunc("/setup/teams/{id}/edit", web.teamEditGetHandler).Methods("GET")
	router.HandleFunc("/setup/teams/{id}/edit", web.teamEditPostHandler).Methods("POST")
	router.HandleFunc("/setup/teams/clear", web.teamsClearHandler).Methods("POST")
	router.HandleFunc("/setup/teams/generate_wpa_keys", web.teamsGenerateWpaKeysHandler).Methods("GET")
	router.HandleFunc("/setup/teams/publish", web.teamsPublishHandler).Methods("POST")
	router.HandleFunc("/setup/teams/refresh", web.teamsRefreshHandler).Methods("GET")
	return router
}

// Writes the given error out as plain text with a status code of 500.
func handleWebErr(w http.ResponseWriter, err error) {
	http.Error(w, "Internal server error: "+err.Error(), 500)
}

// Prepends the base directory to the template filenames.
func (web *Web) parseFiles(filenames ...string) (*template.Template, error) {
	var paths []string
	for _, filename := range filenames {
		paths = append(paths, filepath.Join(model.BaseDir, filename))
	}

	template := template.New("").Funcs(web.templateHelpers)
	return template.ParseFiles(paths...)
}
