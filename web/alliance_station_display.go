// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the alliance station display.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
	"strconv"
)

// Renders the team number and status display shown above each alliance station.
func (web *Web) allianceStationDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	template, err := web.parseFiles("templates/alliance_station_display.html")
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
		*model.EventSettings
		DisplayId string
	}{web.arena.EventSettings, displayId}
	err = template.ExecuteTemplate(w, "alliance_station_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the alliance station display client to receive status updates.
func (web *Web) allianceStationDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	displayId := r.URL.Query()["displayId"][0]
	station, ok := web.arena.AllianceStationDisplays[displayId]
	if !ok {
		station = ""
		web.arena.AllianceStationDisplays[displayId] = station
	}
	rankings := make(map[string]*game.Ranking)
	for _, allianceStation := range web.arena.AllianceStations {
		if allianceStation.Team != nil {
			rankings[strconv.Itoa(allianceStation.Team.Id)], _ =
				web.arena.Database.GetRankingForTeam(allianceStation.Team.Id)
		}
	}

	allianceStationDisplayListener := web.arena.AllianceStationDisplayNotifier.Listen()
	defer close(allianceStationDisplayListener)
	matchLoadTeamsListener := web.arena.MatchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	robotStatusListener := web.arena.RobotStatusNotifier.Listen()
	defer close(robotStatusListener)
	matchTimeListener := web.arena.MatchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := web.arena.RealtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	reloadDisplaysListener := web.arena.ReloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	err = websocket.Write("setAllianceStationDisplay", web.arena.AllianceStationDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", game.MatchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{int(web.arena.MatchState), int(web.arena.LastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		AllianceStation string
		Teams           map[string]*model.Team
		Rankings        map[string]*game.Ranking
	}{station, map[string]*model.Team{"R1": web.arena.AllianceStations["R1"].Team,
		"R2": web.arena.AllianceStations["R2"].Team, "R3": web.arena.AllianceStations["R3"].Team,
		"B1": web.arena.AllianceStations["B1"].Team, "B2": web.arena.AllianceStations["B2"].Team,
		"B3": web.arena.AllianceStations["B3"].Team}, rankings}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
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
				websocket.Write("matchTime",
					MatchTimeMessage{int(web.arena.MatchState), int(web.arena.LastMatchTimeSec)})
				messageType = "setAllianceStationDisplay"
				message = web.arena.AllianceStationDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				station = web.arena.AllianceStationDisplays[displayId]
				rankings := make(map[string]*game.Ranking)
				for _, allianceStation := range web.arena.AllianceStations {
					if allianceStation.Team != nil {
						rankings[strconv.Itoa(allianceStation.Team.Id)], _ =
							web.arena.Database.GetRankingForTeam(allianceStation.Team.Id)
					}
				}
				message = struct {
					AllianceStation string
					Teams           map[string]*model.Team
					Rankings        map[string]*game.Ranking
					MatchType       string
				}{station, map[string]*model.Team{"R1": web.arena.AllianceStations["R1"].Team,
					"R2": web.arena.AllianceStations["R2"].Team, "R3": web.arena.AllianceStations["R3"].Team,
					"B1": web.arena.AllianceStations["B1"].Team, "B2": web.arena.AllianceStations["B2"].Team,
					"B3": web.arena.AllianceStations["B3"].Team}, rankings, web.arena.CurrentMatch.Type}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = web.arena.GetStatus()
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{int(web.arena.MatchState), matchTimeSec.(int)}
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					BlueScore int
				}{web.arena.RedScoreSummary().Score, web.arena.BlueScoreSummary().Score}
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
			web.arena.AllianceStationDisplays[displayId] = station
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
