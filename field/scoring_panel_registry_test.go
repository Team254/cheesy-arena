// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoringPanelRegistry(t *testing.T) {
	var registry ScoringPanelRegistry
	registry.initialize()
	assert.Equal(t, 0, registry.GetNumPanels("red"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, registry.GetNumPanels("blue"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("blue"))

	ws1 := new(websocket.Websocket)
	ws2 := new(websocket.Websocket)
	ws3 := new(websocket.Websocket)
	registry.RegisterPanel("red", ws1)
	registry.RegisterPanel("blue", ws2)
	registry.RegisterPanel("red", ws3)
	assert.Equal(t, 2, registry.GetNumPanels("red"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, registry.GetNumPanels("blue"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("blue"))

	registry.SetScoreCommitted("red", ws3)
	registry.SetScoreCommitted("blue", ws2)
	registry.SetScoreCommitted("blue", ws2)
	assert.Equal(t, 2, registry.GetNumPanels("red"))
	assert.Equal(t, 1, registry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, registry.GetNumPanels("blue"))
	assert.Equal(t, 1, registry.GetNumScoreCommitted("blue"))

	registry.UnregisterPanel("red", ws1)
	registry.UnregisterPanel("blue", ws2)
	assert.Equal(t, 1, registry.GetNumPanels("red"))
	assert.Equal(t, 1, registry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, registry.GetNumPanels("blue"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("blue"))

	registry.resetScoreCommitted()
	assert.Equal(t, 1, registry.GetNumPanels("red"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, registry.GetNumPanels("blue"))
	assert.Equal(t, 0, registry.GetNumScoreCommitted("blue"))
}
