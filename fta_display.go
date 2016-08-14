// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the FTA diagnostic display.

package main

import (
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the FTA diagnostic display.
func FtaDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	// Retrieve the next few matches to show which defenses they will require.
	numUpcomingMatches := 3
	matches, err := db.GetMatchesByType(mainArena.currentMatch.Type)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	var upcomingMatches []Match
	for _, match := range matches {
		if match.Status != "complete" {
			upcomingMatches = append(upcomingMatches, match)
			if len(upcomingMatches) == numUpcomingMatches {
				break
			}
		}
	}

	template := template.New("").Funcs(templateHelpers)
	_, err = template.ParseFiles("templates/fta_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		UpcomingMatches []Match
		DefenseNames    map[string]string
	}{eventSettings, upcomingMatches, defenseNames}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the FTA display client to receive status updates.
func FtaDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(patrick): Enable authentication once Safari (for iPad) supports it over Websocket.

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	robotStatusListener := mainArena.robotStatusNotifier.Listen()
	defer close(robotStatusListener)
	defenseSelectionListener := mainArena.defenseSelectionNotifier.Listen()
	defer close(defenseSelectionListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	err = websocket.Write("status", mainArena)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = mainArena
			case _, ok := <-defenseSelectionListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			}
			err = websocket.Write(messageType, message)
			if err != nil {
				// The client has probably closed the connection; nothing to do here.
				return
			}
		}
	}()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		_, _, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}
