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

func TestSetupDisplays(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/setup/displays")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Display Configuration - Untitled Event - Cheesy Arena")
}

func TestSetupDisplaysWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/displays/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	message := readDisplayConfiguration(t, ws)
	assert.Empty(t, message)

	// Connect a couple of displays and verify the resulting configuration messages.
	displayConn1, _, _ := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/display/websocket?displayId=1", nil)
	defer displayConn1.Close()
	displayWs1 := websocket.NewTestWebsocket(displayConn1)
	assert.Equal(t, "/display?displayId=1", readWebsocketType(t, displayWs1, "displayConfiguration"))
	readDisplayConfiguration(t, ws)
	displayConn2, _, _ := gorillawebsocket.DefaultDialer.Dial(wsUrl+
		"/displays/alliance_station/websocket?displayId=2&station=R2", nil)
	defer displayConn2.Close()
	message = readDisplayConfiguration(t, ws)
	if assert.Equal(t, 2, len(message)) {
		assert.Equal(t, field.DisplayConfiguration{"1", "", field.PlaceholderDisplay, map[string]string{}},
			message["1"].DisplayConfiguration)
		assert.Equal(t, 1, message["1"].ConnectionCount)
		assert.Equal(t, "127.0.0.1", message["1"].IpAddress)
		assert.Equal(t, field.DisplayConfiguration{"2", "", field.AllianceStationDisplay,
			map[string]string{"station": "R2"}}, message["2"].DisplayConfiguration)
		assert.Equal(t, 1, message["2"].ConnectionCount)
		assert.Equal(t, "127.0.0.1", message["2"].IpAddress)
	}

	// Reconfigure a display and verify the result.
	displayConfig := field.DisplayConfiguration{Id: "1", Nickname: "Audience Display", Type: field.AudienceDisplay,
		Configuration: map[string]string{"background": "#00f", "reversed": "true"}}
	ws.Write("configureDisplay", displayConfig)
	message = readDisplayConfiguration(t, ws)
	assert.Equal(t, displayConfig, message["1"].DisplayConfiguration)
	assert.Equal(t, "/displays/audience?displayId=1&nickname=Audience+Display&background=%2300f&reversed=true",
		readWebsocketType(t, displayWs1, "displayConfiguration"))
}

func TestSetupDisplaysWebsocketReloadDisplays(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/displays/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readDisplayConfiguration(t, ws)

	// Connect a display and verify the resulting configuration messages.
	displayConn, _, _ := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/display/websocket?displayId=1", nil)
	defer displayConn.Close()
	displayWs := websocket.NewTestWebsocket(displayConn)
	assert.Equal(t, "/display?displayId=1", readWebsocketType(t, displayWs, "displayConfiguration"))
	readDisplayConfiguration(t, ws)

	// Reset a display selectively and verify the resulting message.
	ws.Write("reloadDisplay", "1")
	assert.Equal(t, "1", readWebsocketType(t, displayWs, "reload"))
	ws.Write("reloadAllDisplays", nil)
	assert.Equal(t, nil, readWebsocketType(t, displayWs, "reload"))
}

func readDisplayConfiguration(t *testing.T, ws *websocket.Websocket) map[string]field.Display {
	message := readWebsocketType(t, ws, "displayConfiguration")
	var displayConfigurationMessage map[string]field.Display
	err := mapstructure.Decode(message, &displayConfigurationMessage)
	assert.Nil(t, err)
	return displayConfigurationMessage
}
