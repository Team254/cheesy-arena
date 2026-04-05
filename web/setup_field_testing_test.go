// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"testing"
)

type plcIoChangeMessage struct {
	Inputs        []bool
	Registers     []uint16
	Coils         []bool
	CoilOverrides []string
}

func TestSetupFieldTesting(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/setup/field_testing")
	assert.Equal(t, 200, recorder.Code)
}

func TestSetupFieldTestingWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/field_testing/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	messages := readWebsocketMultiple(t, ws, 2)
	_, hasPlcIoChange := messages["plcIoChange"]
	_, hasArenaStatus := messages["arenaStatus"]
	assert.True(t, hasPlcIoChange)
	assert.True(t, hasArenaStatus)

	// Also create a websocket to the audience display to check that it plays the requested game sound.
	audienceConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/audience/websocket?displayId=1", nil)
	assert.Nil(t, err)
	defer audienceConn.Close()
	audienceWs := websocket.NewTestWebsocket(audienceConn)
	readWebsocketMultiple(t, audienceWs, 9)

	ws.Write("playSound", "resume")
	assert.Equal(t, "resume", readWebsocketType(t, audienceWs, "playSound"))
}

func TestSetupFieldTestingWebsocketSetPlcCoilOverride(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/field_testing/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	readWebsocketMultiple(t, ws, 2)

	ws.Write("setPlcCoilOverride", map[string]any{"Index": 7, "Override": "on"})
	plcIoChange := readPlcIoChangeMessage(t, ws)
	assert.Equal(t, true, plcIoChange.Coils[7])
	assert.Equal(t, "on", plcIoChange.CoilOverrides[7])

	ws.Write("setPlcCoilOverride", map[string]any{"Index": 7, "Override": "off"})
	plcIoChange = readPlcIoChangeMessage(t, ws)
	assert.Equal(t, false, plcIoChange.Coils[7])
	assert.Equal(t, "off", plcIoChange.CoilOverrides[7])

	ws.Write("setPlcCoilOverride", map[string]any{"Index": 7, "Override": "auto"})
	plcIoChange = readPlcIoChangeMessage(t, ws)
	assert.Equal(t, false, plcIoChange.Coils[7])
	assert.Equal(t, "auto", plcIoChange.CoilOverrides[7])
}

func TestSetupFieldTestingWebsocketSetPlcCoilOverrideAllowedStates(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/field_testing/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	readWebsocketMultiple(t, ws, 2)

	for _, matchState := range []field.MatchState{
		field.PreMatch,
		field.PostMatch,
		field.TimeoutActive,
		field.PostTimeout,
	} {
		web.arena.MatchState = matchState
		ws.Write("setPlcCoilOverride", map[string]any{"Index": 8, "Override": "on"})
		plcIoChange := readPlcIoChangeMessage(t, ws)
		assert.Equal(t, true, plcIoChange.Coils[8])
		assert.Equal(t, "on", plcIoChange.CoilOverrides[8])

		ws.Write("setPlcCoilOverride", map[string]any{"Index": 8, "Override": "auto"})
		plcIoChange = readPlcIoChangeMessage(t, ws)
		assert.Equal(t, "auto", plcIoChange.CoilOverrides[8])
	}
}

func TestSetupFieldTestingWebsocketSetPlcCoilOverrideDisallowedStates(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/field_testing/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	readWebsocketMultiple(t, ws, 2)

	for _, matchState := range []field.MatchState{
		field.StartMatch,
		field.AutoPeriod,
		field.PausePeriod,
		field.TeleopPeriod,
	} {
		web.arena.MatchState = matchState
		ws.Write("setPlcCoilOverride", map[string]any{"Index": 8, "Override": "on"})
		assert.Equal(t, fieldTestingOverrideDisabledMessage, readWebsocketError(t, ws))
	}
}

func TestSetupFieldTestingWebsocketSetPlcCoilOverrideInvalidArgs(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/field_testing/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	readWebsocketMultiple(t, ws, 2)

	ws.Write("setPlcCoilOverride", map[string]any{"Index": 8, "Override": "invalid"})
	assert.Equal(t, "Invalid coil override state 'invalid'.", readWebsocketError(t, ws))
}

func readPlcIoChangeMessage(t *testing.T, ws *websocket.Websocket) plcIoChangeMessage {
	var message plcIoChangeMessage
	assert.Nil(t, mapstructure.Decode(readWebsocketType(t, ws, "plcIoChange"), &message))
	return message
}
