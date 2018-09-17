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
	assert.Empty(t, message.Displays)
	assert.Empty(t, message.DisplayUrls)

	// Connect a couple of displays and verify the resulting configuration messages.
	displayConn1, _, _ := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/display/websocket?displayId=1", nil)
	defer displayConn1.Close()
	readDisplayConfiguration(t, ws)
	displayConn2, _, _ := gorillawebsocket.DefaultDialer.Dial(wsUrl+
		"/displays/alliance_station/websocket?displayId=2&station=R2", nil)
	defer displayConn2.Close()
	expectedDisplay1 := &field.Display{Id: "1", Type: field.PlaceholderDisplay, Configuration: map[string]string{},
		ConnectionCount: 1, IpAddress: "127.0.0.1"}
	expectedDisplay2 := &field.Display{Id: "2", Type: field.AllianceStationDisplay,
		Configuration: map[string]string{"station": "R2"}, ConnectionCount: 1, IpAddress: "127.0.0.1"}
	message = readDisplayConfiguration(t, ws)
	if assert.Equal(t, 2, len(message.Displays)) {
		assert.Equal(t, expectedDisplay1, message.Displays["1"])
		assert.Equal(t, expectedDisplay2, message.Displays["2"])
		assert.Equal(t, expectedDisplay1.ToUrl(), message.DisplayUrls["1"])
		assert.Equal(t, expectedDisplay2.ToUrl(), message.DisplayUrls["2"])
	}

	// Reconfigure a display and verify the result.
	expectedDisplay1.Nickname = "Audience Display"
	expectedDisplay1.Type = field.AudienceDisplay
	expectedDisplay1.Configuration["background"] = "#00f"
	expectedDisplay1.Configuration["reversed"] = "true"
	ws.Write("configureDisplay", expectedDisplay1)
	message = readDisplayConfiguration(t, ws)
	assert.Equal(t, expectedDisplay1, message.Displays["1"])
	assert.Equal(t, expectedDisplay1.ToUrl(), message.DisplayUrls["1"])
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
	readDisplayConfiguration(t, displayWs)
	readDisplayConfiguration(t, ws)

	// Reset a display selectively and verify the resulting message.
	ws.Write("reloadDisplay", "1")
	assert.Equal(t, "1", readWebsocketType(t, displayWs, "reload"))
	ws.Write("reloadAllDisplays", nil)
	assert.Equal(t, nil, readWebsocketType(t, displayWs, "reload"))
}

func readDisplayConfiguration(t *testing.T, ws *websocket.Websocket) *field.DisplayConfigurationMessage {
	message := readWebsocketType(t, ws, "displayConfiguration")
	var displayConfigurationMessage field.DisplayConfigurationMessage
	err := mapstructure.Decode(message, &displayConfigurationMessage)
	assert.Nil(t, err)
	return &displayConfigurationMessage
}
