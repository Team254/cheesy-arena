// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Configuration and functions for the event server web interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
)

const (
	sessionTokenCookie = "session_token"
	adminUser          = "admin"
)

type Web struct {
	arena           *field.Arena
	templateHelpers template.FuncMap
}

func NewWeb(arena *field.Arena) *Web {
	web := &Web{arena: arena}

	// Helper functions that can be used inside templates.
	web.templateHelpers = template.FuncMap{
		// Allows sub-templates to be invoked with multiple arguments.
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("Invalid dict call.")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("Dict keys must be strings.")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"add": func(a, b int) int {
			return a + b
		},
		"itoa": func(a int) string {
			return strconv.Itoa(a)
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"seq": func(count int) []int {
			seq := make([]int, count)
			for i := 0; i < count; i++ {
				seq[i] = i + 1
			}
			return seq
		},
		"toUpper": func(str string) string {
			return strings.ToUpper(str)
		},

		// MatchType enum values.
		"testMatch":          model.Test.Get,
		"practiceMatch":      model.Practice.Get,
		"qualificationMatch": model.Qualification.Get,
		"playoffMatch":       model.Playoff.Get,

		// MatchStatus enum values.
		"matchScheduled": game.MatchScheduled.Get,
		"matchHidden":    game.MatchHidden.Get,
		"redWonMatch":    game.RedWonMatch.Get,
		"blueWonMatch":   game.BlueWonMatch.Get,
		"tieMatch":       game.TieMatch.Get,
	}

	return web
}

// Starts the webserver and blocks, waiting on requests. Does not return until the application exits.
func (web *Web) ServeWebInterface(port int) {
	http.Handle("/static/", http.StripPrefix("/static/", addNoCacheHeader(http.FileServer(http.Dir("static/")))))
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

// Adds a "Cache-Control: no-cache" header to the given handler to force browser validation of last modified time.
func addNoCacheHeader(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		handler.ServeHTTP(w, r)
	})
}

// Sets up the mapping between URLs and handlers.
func (web *Web) newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", web.indexHandler)
	mux.HandleFunc("GET /alliance_selection", web.allianceSelectionGetHandler)
	mux.HandleFunc("POST /alliance_selection", web.allianceSelectionPostHandler)
	mux.HandleFunc("GET /alliance_selection/websocket", web.allianceSelectionWebsocketHandler)
	mux.HandleFunc("POST /alliance_selection/finalize", web.allianceSelectionFinalizeHandler)
	mux.HandleFunc("POST /alliance_selection/reset", web.allianceSelectionResetHandler)
	mux.HandleFunc("POST /alliance_selection/start", web.allianceSelectionStartHandler)
	mux.HandleFunc("GET /api/alliances", web.alliancesApiHandler)
	mux.HandleFunc("GET /api/arena/websocket", web.arenaWebsocketApiHandler)
	mux.HandleFunc("GET /api/bracket/svg", web.bracketSvgApiHandler)
	mux.HandleFunc("GET /api/matches/{type}", web.matchesApiHandler)
	mux.HandleFunc("GET /api/rankings", web.rankingsApiHandler)
	mux.HandleFunc("GET /api/sponsor_slides", web.sponsorSlidesApiHandler)
	mux.HandleFunc("GET /api/teams/{teamId}/avatar", web.teamAvatarsApiHandler)
	mux.HandleFunc("GET /display", web.placeholderDisplayHandler)
	mux.HandleFunc("GET /display/websocket", web.placeholderDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/alliance_station", web.allianceStationDisplayHandler)
	mux.HandleFunc("GET /displays/alliance_station/websocket", web.allianceStationDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/announcer", web.announcerDisplayHandler)
	mux.HandleFunc("GET /displays/announcer/match_load", web.announcerDisplayMatchLoadHandler)
	mux.HandleFunc("GET /displays/announcer/score_posted", web.announcerDisplayScorePostedHandler)
	mux.HandleFunc("GET /displays/announcer/websocket", web.announcerDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/audience", web.audienceDisplayHandler)
	mux.HandleFunc("GET /displays/audience/websocket", web.audienceDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/bracket", web.bracketDisplayHandler)
	mux.HandleFunc("GET /displays/bracket/websocket", web.bracketDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/field_monitor", web.fieldMonitorDisplayHandler)
	mux.HandleFunc("GET /displays/field_monitor/websocket", web.fieldMonitorDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/logo", web.logoDisplayHandler)
	mux.HandleFunc("GET /displays/logo/websocket", web.logoDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/queueing", web.queueingDisplayHandler)
	mux.HandleFunc("GET /displays/queueing/match_load", web.queueingDisplayMatchLoadHandler)
	mux.HandleFunc("GET /displays/queueing/websocket", web.queueingDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/rankings", web.rankingsDisplayHandler)
	mux.HandleFunc("GET /displays/rankings/websocket", web.rankingsDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/twitch", web.twitchDisplayHandler)
	mux.HandleFunc("GET /displays/twitch/websocket", web.twitchDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/wall", web.wallDisplayHandler)
	mux.HandleFunc("GET /displays/wall/websocket", web.wallDisplayWebsocketHandler)
	mux.HandleFunc("GET /displays/webpage", web.webpageDisplayHandler)
	mux.HandleFunc("GET /displays/webpage/websocket", web.webpageDisplayWebsocketHandler)
	mux.HandleFunc("GET /login", web.loginHandler)
	mux.HandleFunc("POST /login", web.loginPostHandler)
	mux.HandleFunc("GET /match_play", web.matchPlayHandler)
	mux.HandleFunc("GET /match_play/match_load", web.matchPlayMatchLoadHandler)
	mux.HandleFunc("GET /match_play/websocket", web.matchPlayWebsocketHandler)
	mux.HandleFunc("GET /match_logs", web.matchLogsHandler)
	mux.HandleFunc("GET /match_logs/{matchId}/{stationId}/log", web.matchLogsViewGetHandler)
	mux.HandleFunc("GET /match_review", web.matchReviewHandler)
	mux.HandleFunc("GET /match_review/{matchId}/edit", web.matchReviewEditGetHandler)
	mux.HandleFunc("POST /match_review/{matchId}/edit", web.matchReviewEditPostHandler)
	mux.HandleFunc("GET /panels/scoring/{alliance}", web.scoringPanelHandler)
	mux.HandleFunc("GET /panels/scoring/{alliance}/websocket", web.scoringPanelWebsocketHandler)
	mux.HandleFunc("GET /panels/referee", web.refereePanelHandler)
	mux.HandleFunc("GET /panels/referee/foul_list", web.refereePanelFoulListHandler)
	mux.HandleFunc("GET /panels/referee/websocket", web.refereePanelWebsocketHandler)
	mux.HandleFunc("GET /reports/csv/backups", web.backupTeamsCsvReportHandler)
	mux.HandleFunc("GET /reports/csv/fta", web.ftaCsvReportHandler)
	mux.HandleFunc("GET /reports/csv/rankings", web.rankingsCsvReportHandler)
	mux.HandleFunc("GET /reports/csv/schedule/{type}", web.scheduleCsvReportHandler)
	mux.HandleFunc("GET /reports/csv/teams", web.teamsCsvReportHandler)
	mux.HandleFunc("GET /reports/csv/wpa_keys", web.wpaKeysCsvReportHandler)
	mux.HandleFunc("GET /reports/pdf/alliances", web.alliancesPdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/backups", web.backupsPdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/bracket", web.bracketPdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/coupons", web.couponsPdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/cycle/{type}", web.cyclePdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/rankings", web.rankingsPdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/schedule/{type}", web.schedulePdfReportHandler)
	mux.HandleFunc("GET /reports/pdf/teams", web.teamsPdfReportHandler)
	mux.HandleFunc("GET /setup/awards", web.awardsGetHandler)
	mux.HandleFunc("POST /setup/awards", web.awardsPostHandler)
	mux.HandleFunc("GET /setup/breaks", web.breaksGetHandler)
	mux.HandleFunc("POST /setup/breaks", web.breaksPostHandler)
	mux.HandleFunc("POST /setup/db/clear/{type}", web.clearDbHandler)
	mux.HandleFunc("POST /setup/db/restore", web.restoreDbHandler)
	mux.HandleFunc("GET /setup/db/save", web.saveDbHandler)
	mux.HandleFunc("GET /setup/displays", web.displaysGetHandler)
	mux.HandleFunc("GET /setup/displays/websocket", web.displaysWebsocketHandler)
	mux.HandleFunc("GET /setup/field_testing", web.fieldTestingGetHandler)
	mux.HandleFunc("GET /setup/field_testing/websocket", web.fieldTestingWebsocketHandler)
	mux.HandleFunc("GET /setup/lower_thirds", web.lowerThirdsGetHandler)
	mux.HandleFunc("GET /setup/lower_thirds/websocket", web.lowerThirdsWebsocketHandler)
	mux.HandleFunc("GET /setup/schedule", web.scheduleGetHandler)
	mux.HandleFunc("POST /setup/schedule/generate", web.scheduleGeneratePostHandler)
	mux.HandleFunc("POST /setup/schedule/save", web.scheduleSavePostHandler)
	mux.HandleFunc("GET /setup/settings", web.settingsGetHandler)
	mux.HandleFunc("POST /setup/settings", web.settingsPostHandler)
	mux.HandleFunc("GET /setup/settings/publish_alliances", web.settingsPublishAlliancesHandler)
	mux.HandleFunc("GET /setup/settings/publish_awards", web.settingsPublishAwardsHandler)
	mux.HandleFunc("GET /setup/settings/publish_matches", web.settingsPublishMatchesHandler)
	mux.HandleFunc("GET /setup/settings/publish_rankings", web.settingsPublishRankingsHandler)
	mux.HandleFunc("GET /setup/settings/publish_teams", web.settingsPublishTeamsHandler)
	mux.HandleFunc("GET /setup/sponsor_slides", web.sponsorSlidesGetHandler)
	mux.HandleFunc("POST /setup/sponsor_slides", web.sponsorSlidesPostHandler)
	mux.HandleFunc("GET /setup/teams", web.teamsGetHandler)
	mux.HandleFunc("POST /setup/teams", web.teamsPostHandler)
	mux.HandleFunc("POST /setup/teams/{id}/delete", web.teamDeletePostHandler)
	mux.HandleFunc("GET /setup/teams/{id}/edit", web.teamEditGetHandler)
	mux.HandleFunc("POST /setup/teams/{id}/edit", web.teamEditPostHandler)
	mux.HandleFunc("POST /setup/teams/clear", web.teamsClearHandler)
	mux.HandleFunc("GET /setup/teams/generate_wpa_keys", web.teamsGenerateWpaKeysHandler)
	mux.HandleFunc("GET /setup/teams/progress", web.teamsUpdateProgressBarHandler)
	mux.HandleFunc("GET /setup/teams/refresh", web.teamsRefreshHandler)
	return mux
}

// Writes the given error out as plain text with a status code of 500.
func handleWebErr(w http.ResponseWriter, err error) {
	log.Printf("HTTP request error: %v", err)
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
