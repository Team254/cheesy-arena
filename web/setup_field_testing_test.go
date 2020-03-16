// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupFieldTesting(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/setup/field_testing")
	assert.Equal(t, 200, recorder.Code)
	for _, sound := range game.MatchSounds {
		assert.Contains(t, recorder.Body.String(), sound.Name)
	}
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
	readWebsocketType(t, ws, "plcIoChange")

	// Also create a websocket to the audience display to check that it plays the requested game sound.
	audienceConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/audience/websocket?displayId=1", nil)
	assert.Nil(t, err)
	defer audienceConn.Close()
	audienceWs := websocket.NewTestWebsocket(audienceConn)
	readWebsocketMultiple(t, audienceWs, 9)

	ws.Write("playSound", "resume")
	assert.Equal(t, "resume", readWebsocketType(t, audienceWs, "playSound"))
}
