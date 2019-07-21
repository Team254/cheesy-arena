// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/mitchellh/mapstructure"
)

func TestSetupLedPlcWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/led_plc/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readLedModes(t, ws)
	readWebsocketType(t, ws, "plcIoChange")
}

func readLedModes(t *testing.T, ws *websocket.Websocket) *field.LedModeMessage {
	message := readWebsocketType(t, ws, "ledMode")
	var ledModeMessage field.LedModeMessage
	err := mapstructure.Decode(message, &ledModeMessage)
	assert.Nil(t, err)
	return &ledModeMessage
}
