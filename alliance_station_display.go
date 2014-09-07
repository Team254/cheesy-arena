// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the alliance station display.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

// Renders the team number and status display shown above each alliance station.
func AllianceStationDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/alliance_station_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	displayId := ""
	if _, ok := r.URL.Query()["displayId"]; ok {
		// Register the display in memory by its ID so that it can be configured to a certain station.
		displayId = r.URL.Query()["displayId"][0]
	}

	data := struct {
		*EventSettings
		DisplayId string
	}{eventSettings, displayId}
	err = template.ExecuteTemplate(w, "alliance_station_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the alliance station display client to receive status updates.
func AllianceStationDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	displayId := r.URL.Query()["displayId"][0]
	station, ok := mainArena.allianceStationDisplays[displayId]
	if !ok {
		station = ""
		mainArena.allianceStationDisplays[displayId] = station
	}

	allianceStationDisplayListener := mainArena.allianceStationDisplayNotifier.Listen()
	defer close(allianceStationDisplayListener)
	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	robotStatusListener := mainArena.robotStatusNotifier.Listen()
	defer close(robotStatusListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	hotGoalLightListener := mainArena.hotGoalLightNotifier.Listen()
	defer close(hotGoalLightListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	err = websocket.Write("setAllianceStationDisplay", mainArena.allianceStationDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		AllianceStation string
		Teams           map[string]*Team
	}{station, map[string]*Team{"R1": mainArena.AllianceStations["R1"].team,
		"R2": mainArena.AllianceStations["R2"].team, "R3": mainArena.AllianceStations["R3"].team,
		"B1": mainArena.AllianceStations["B1"].team, "B2": mainArena.AllianceStations["B2"].team,
		"B3": mainArena.AllianceStations["B3"].team}}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		RedCycle  Cycle
		BlueScore int
		BlueCycle Cycle
	}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
		mainArena.redRealtimeScore.CurrentCycle,
		mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
		mainArena.blueRealtimeScore.CurrentCycle}
	err = websocket.Write("realtimeScore", data)
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
			case _, ok := <-allianceStationDisplayListener:
				if !ok {
					return
				}
				websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
				messageType = "setAllianceStationDisplay"
				message = mainArena.allianceStationDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				station = mainArena.allianceStationDisplays[displayId]
				message = struct {
					AllianceStation string
					Teams           map[string]*Team
				}{station, map[string]*Team{station: mainArena.AllianceStations["R1"].team,
					"R2": mainArena.AllianceStations["R2"].team, "R3": mainArena.AllianceStations["R3"].team,
					"B1": mainArena.AllianceStations["B1"].team, "B2": mainArena.AllianceStations["B2"].team,
					"B3": mainArena.AllianceStations["B3"].team}}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = mainArena
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{mainArena.MatchState, matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					RedCycle  Cycle
					BlueScore int
					BlueCycle Cycle
				}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
					mainArena.redRealtimeScore.CurrentCycle,
					mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
					mainArena.blueRealtimeScore.CurrentCycle}
			case side, ok := <-hotGoalLightListener:
				if !ok {
					return
				}
				if !eventSettings.AllianceDisplayHotGoals {
					continue
				}
				messageType = "hotGoalLight"
				message = side
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
		messageType, data, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}

		switch messageType {
		case "setAllianceStation":
			// The client knows what station it is (e.g. across a server restart) and is informing the server.
			station, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.allianceStationDisplays[displayId] = station
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
