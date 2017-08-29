// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the pit rankings display.

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the pit display which shows scrolling rankings.
func (web *Web) pitDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	template, err := template.ParseFiles("templates/pit_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.Execute(w, data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the pit display, used only to force reloads remotely.
func (web *Web) pitDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	reloadDisplaysListener := web.arena.ReloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
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
