// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaceholderDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/audience")
	assert.Equal(t, 302, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Location"), "displayId=100")

	recorder = web.getHttpResponse("/display?displayId=1")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Placeholder Display - Untitled Event - Cheesy Arena")
}

func TestPlaceholderDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/display/websocket?displayId=123&nickname=blop&a=b", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "displayConfiguration")

	if assert.Contains(t, web.arena.Displays, "123") {
		assert.Equal(t, "blop", web.arena.Displays["123"].DisplayConfiguration.Nickname)
		if assert.Equal(t, 1, len(web.arena.Displays["123"].DisplayConfiguration.Configuration)) {
			assert.Equal(t, "b", web.arena.Displays["123"].DisplayConfiguration.Configuration["a"])
		}
	}

	// Reconfigure the display and verify that the new configuration is received.
	displayConfig := field.DisplayConfiguration{Id: "123", Nickname: "Alliance", Type: field.AllianceStationDisplay,
		Configuration: map[string]string{"station": "B2"}}
	web.arena.UpdateDisplay(displayConfig)
	readWebsocketType(t, ws, "displayConfiguration")
}
