// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Configuration and functions for the event server web interface.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

const httpPort = 8080

var websocketUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 2014}

// Helper functions that can be used inside templates.
var templateHelpers = template.FuncMap{
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

// Wraps the Gorilla Websocket module for convenience.
type Websocket struct {
	conn *websocket.Conn
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWebsocket(w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Websocket{conn}, nil
}

func (websocket *Websocket) Close() {
	websocket.conn.Close()
}

func (websocket *Websocket) Read() (string, interface{}, error) {
	var message WebsocketMessage
	err := websocket.conn.ReadJSON(&message)
	return message.Type, message.Data, err
}

func (websocket *Websocket) Write(messageType string, data interface{}) error {
	return websocket.conn.WriteJSON(WebsocketMessage{messageType, data})
}

func (websocket *Websocket) WriteError(errorMessage string) error {
	return websocket.conn.WriteJSON(WebsocketMessage{"error", errorMessage})
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func ServeWebInterface() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/", newHandler())
	log.Printf("Serving HTTP requests on port %d", httpPort)

	// Start Server
	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
}

func newHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/setup/settings", SettingsGetHandler).Methods("GET")
	router.HandleFunc("/setup/settings", SettingsPostHandler).Methods("POST")
	router.HandleFunc("/setup/db/save", SaveDbHandler).Methods("GET")
	router.HandleFunc("/setup/db/restore", RestoreDbHandler).Methods("POST")
	router.HandleFunc("/setup/db/clear", ClearDbHandler).Methods("POST")
	router.HandleFunc("/setup/teams", TeamsGetHandler).Methods("GET")
	router.HandleFunc("/setup/teams", TeamsPostHandler).Methods("POST")
	router.HandleFunc("/setup/teams/clear", TeamsClearHandler).Methods("POST")
	router.HandleFunc("/setup/teams/{id}/edit", TeamEditGetHandler).Methods("GET")
	router.HandleFunc("/setup/teams/{id}/edit", TeamEditPostHandler).Methods("POST")
	router.HandleFunc("/setup/teams/{id}/delete", TeamDeletePostHandler).Methods("POST")
	router.HandleFunc("/setup/schedule", ScheduleGetHandler).Methods("GET")
	router.HandleFunc("/setup/schedule/generate", ScheduleGeneratePostHandler).Methods("POST")
	router.HandleFunc("/setup/schedule/save", ScheduleSavePostHandler).Methods("POST")
	router.HandleFunc("/setup/alliance_selection", AllianceSelectionGetHandler).Methods("GET")
	router.HandleFunc("/setup/alliance_selection", AllianceSelectionPostHandler).Methods("POST")
	router.HandleFunc("/setup/alliance_selection/start", AllianceSelectionStartHandler).Methods("POST")
	router.HandleFunc("/setup/alliance_selection/reset", AllianceSelectionResetHandler).Methods("POST")
	router.HandleFunc("/setup/alliance_selection/finalize", AllianceSelectionFinalizeHandler).Methods("POST")
	router.HandleFunc("/match_play", MatchPlayHandler).Methods("GET")
	router.HandleFunc("/match_play/{matchId}/load", MatchPlayLoadHandler).Methods("GET")
	router.HandleFunc("/match_play/websocket", MatchPlayWebsocketHandler).Methods("GET")
	router.HandleFunc("/match_review", MatchReviewHandler).Methods("GET")
	router.HandleFunc("/match_review/{matchId}/edit", MatchReviewEditGetHandler).Methods("GET")
	router.HandleFunc("/match_review/{matchId}/edit", MatchReviewEditPostHandler).Methods("POST")
	router.HandleFunc("/reports/csv/rankings", RankingsCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/rankings", RankingsPdfReportHandler).Methods("GET")
	router.HandleFunc("/reports/csv/schedule/{type}", ScheduleCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/schedule/{type}", SchedulePdfReportHandler).Methods("GET")
	router.HandleFunc("/reports/csv/teams", TeamsCsvReportHandler).Methods("GET")
	router.HandleFunc("/reports/pdf/teams", TeamsPdfReportHandler).Methods("GET")
	router.HandleFunc("/api/rankings", RankingsApiHandler).Methods("GET")
	router.HandleFunc("/", IndexHandler).Methods("GET")
	return router
}

func handleWebErr(w http.ResponseWriter, err error) {
	http.Error(w, "Internal server error: "+err.Error(), 500)
}
