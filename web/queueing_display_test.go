// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueueingDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/queueing?displayId=1")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Queueing Display - Untitled Event - Cheesy Arena")
}

func TestQueueingDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/queueing/websocket?displayId=1", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchLoad")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "displayConfiguration")
}

func TestQueueingStatusMessage(t *testing.T) {
	assert.Equal(t, "", generateEventStatusMessage("practice", []model.Match{}))

	matches := make([]model.Match, 3)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("practice", matches))

	// Check within threshold considered to be on time.
	setMatchLateness(&matches[1], 0)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))
	setMatchLateness(&matches[1], 60)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("practice", matches))
	setMatchLateness(&matches[1], -60)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))
	setMatchLateness(&matches[1], 90)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))
	setMatchLateness(&matches[1], -90)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))
	setMatchLateness(&matches[1], 110)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("practice", matches))
	setMatchLateness(&matches[1], -110)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))

	// Check lateness.
	setMatchLateness(&matches[1], 130)
	assert.Equal(t, "Event is running 2 minutes late", generateEventStatusMessage("practice", matches))
	setMatchLateness(&matches[1], 3601)
	assert.Equal(t, "Event is running 60 minutes late", generateEventStatusMessage("qualification", matches))

	// Check earliness.
	setMatchLateness(&matches[1], -130)
	assert.Equal(t, "Event is running 2 minutes early", generateEventStatusMessage("qualification", matches))
	setMatchLateness(&matches[1], -3601)
	assert.Equal(t, "Event is running 60 minutes early", generateEventStatusMessage("practice", matches))

	// Check other match types.
	assert.Equal(t, "", generateEventStatusMessage("test", matches))
	assert.Equal(t, "", generateEventStatusMessage("elimination", matches))

	// Check that later matches supersede earlier ones.
	matches = append(matches, model.Match{})
	setMatchLateness(&matches[2], 180)
	assert.Equal(t, "Event is running 3 minutes late", generateEventStatusMessage("qualification", matches))

	// Check that a lateness before a large gap is ignored.
	matches[3].Time = time.Now().Add(time.Minute * 25)
	assert.Equal(t, "Event is running on schedule", generateEventStatusMessage("qualification", matches))
}

func setMatchLateness(match *model.Match, secondsLate int) {
	match.Time = time.Now()
	match.StartedAt = time.Now().Add(time.Second * time.Duration(secondsLate))
	match.Status = "complete"
}
