// Copyright 2014 Team 254. All Rights Reserved.
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

func TestSetupLowerThirds(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateLowerThird(&model.LowerThird{0, "Top Text 1", "Bottom Text 1", 0, 0})
	web.arena.Database.CreateLowerThird(&model.LowerThird{0, "Top Text 2", "Bottom Text 2", 1, 0})
	web.arena.Database.CreateLowerThird(&model.LowerThird{0, "Top Text 3", "Bottom Text 3", 2, 0})

	recorder := web.getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Top Text 1")
	assert.Contains(t, recorder.Body.String(), "Bottom Text 2")

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/lower_thirds/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	ws.Write("saveLowerThird", model.LowerThird{1, "Top Text 4", "Bottom Text 1", 0, 0})
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	lowerThird, _ := web.arena.Database.GetLowerThirdById(1)
	assert.Equal(t, "Top Text 4", lowerThird.TopText)

	ws.Write("deleteLowerThird", model.LowerThird{1, "Top Text 4", "Bottom Text 1", 0, 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(1)
	assert.Nil(t, lowerThird)

	assert.Equal(t, "blank", web.arena.AudienceDisplayMode)
	ws.Write("showLowerThird", model.LowerThird{2, "Top Text 5", "Bottom Text 1", 0, 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 5", lowerThird.TopText)
	assert.Equal(t, true, web.arena.ShowLowerThird)

	ws.Write("hideLowerThird", model.LowerThird{2, "Top Text 6", "Bottom Text 1", 0, 0})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 6", lowerThird.TopText)
	assert.Equal(t, false, web.arena.ShowLowerThird)

	ws.Write("reorderLowerThird", map[string]any{"Id": 2, "moveUp": false})
	time.Sleep(time.Millisecond * 100)
	lowerThirds, _ := web.arena.Database.GetAllLowerThirds()
	assert.Equal(t, 3, lowerThirds[0].Id)
	assert.Equal(t, 2, lowerThirds[1].Id)
}
