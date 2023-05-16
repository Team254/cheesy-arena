// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestFieldMonitorDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/field_monitor?displayId=1&fta=true&reversed=false")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Monitor - Untitled Event - Cheesy Arena")
}

func TestFieldMonitorDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	assert.Nil(t, web.arena.SubstituteTeam(254, "B1"))

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/field_monitor/websocket?displayId=1&fta=false",
		nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "displayConfiguration")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "matchLoad")

	// Should not be able to update team notes.
	ws.Write("updateTeamNotes", map[string]any{"station": "B1", "notes": "Bypassed in M1"})
	assert.Contains(t, readWebsocketError(t, ws), "Must be in FTA mode to update team notes")
	assert.Equal(t, "", web.arena.AllianceStations["B1"].Team.FtaNotes)
}

func TestFieldMonitorFtaDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.Database.CreateTeam(&model.Team{Id: 254})
	assert.Nil(t, web.arena.SubstituteTeam(254, "B1"))

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/field_monitor/websocket?displayId=1&fta=true",
		nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "displayConfiguration")
	readWebsocketType(t, ws, "arenaStatus")
	readWebsocketType(t, ws, "eventStatus")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "matchLoad")

	// Should not be able to update team notes.
	ws.Write("updateTeamNotes", map[string]any{"station": "B1", "notes": "Bypassed in M1"})
	readWebsocketType(t, ws, "arenaStatus")
	assert.Equal(t, "Bypassed in M1", web.arena.AllianceStations["B1"].Team.FtaNotes)

	// Check error scenarios.
	ws.Write("updateTeamNotes", map[string]any{"station": "N", "notes": "Bypassed in M2"})
	assert.Contains(t, readWebsocketError(t, ws), "Invalid alliance station")
	ws.Write("updateTeamNotes", map[string]any{"station": "R3", "notes": "Bypassed in M3"})
	assert.Contains(t, readWebsocketError(t, ws), "No team present")
}
