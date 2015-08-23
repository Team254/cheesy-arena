// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetupLowerThirds(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	db.CreateLowerThird(&LowerThird{0, "Top Text 1", "Bottom Text 1", 0})
	db.CreateLowerThird(&LowerThird{0, "Top Text 2", "Bottom Text 2", 1})
	db.CreateLowerThird(&LowerThird{0, "Top Text 3", "Bottom Text 3", 2})

	recorder := getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Top Text 1")
	assert.Contains(t, recorder.Body.String(), "Bottom Text 2")

	server, wsUrl := startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/setup/lower_thirds/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn}

	ws.Write("saveLowerThird", LowerThird{1, "Top Text 4", "Bottom Text 1", 0})
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	lowerThird, _ := db.GetLowerThirdById(1)
	assert.Equal(t, "Top Text 4", lowerThird.TopText)

	ws.Write("deleteLowerThird", LowerThird{1, "Top Text 4", "Bottom Text 1", 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = db.GetLowerThirdById(1)
	assert.Nil(t, lowerThird)

	assert.Equal(t, "blank", mainArena.audienceDisplayScreen)
	ws.Write("showLowerThird", LowerThird{2, "Top Text 5", "Bottom Text 1", 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = db.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 5", lowerThird.TopText)
	assert.Equal(t, "lowerThird", mainArena.audienceDisplayScreen)

	ws.Write("hideLowerThird", LowerThird{2, "Top Text 6", "Bottom Text 1", 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = db.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 6", lowerThird.TopText)
	assert.Equal(t, "blank", mainArena.audienceDisplayScreen)

	ws.Write("reorderLowerThird", map[string]interface{}{"Id": 2, "moveUp": false})
	time.Sleep(time.Millisecond * 100)
	lowerThirds, _ := db.GetAllLowerThirds()
	assert.Equal(t, 3, lowerThirds[0].Id)
	assert.Equal(t, 2, lowerThirds[1].Id)
}
