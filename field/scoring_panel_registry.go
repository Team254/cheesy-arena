// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing and methods for tracking the state of a realtime scoring panel.

package field

import (
	"github.com/Team254/cheesy-arena/websocket"
	"sync"
)

type ScoringPanelRegistry struct {
	scoringPanels map[string]map[*websocket.Websocket]bool // The score committed state for each panel.
	mutex         sync.Mutex
}

func (registry *ScoringPanelRegistry) initialize() {
	registry.scoringPanels = map[string]map[*websocket.Websocket]bool{"red": {}, "blue": {}}
}

// Resets the score committed state for each registered panel to false.
func (registry *ScoringPanelRegistry) resetScoreCommitted() {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	for _, alliancePanels := range registry.scoringPanels {
		for key := range alliancePanels {
			alliancePanels[key] = false
		}
	}
}

// Returns the number of registered panels for the given alliance.
func (registry *ScoringPanelRegistry) GetNumPanels(alliance string) int {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	return len(registry.scoringPanels[alliance])
}

// Returns the number of registered panels whose score is committed for the given alliance.
func (registry *ScoringPanelRegistry) GetNumScoreCommitted(alliance string) int {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	numCommitted := 0
	for _, panel := range registry.scoringPanels[alliance] {
		if panel {
			numCommitted++
		}
	}
	return numCommitted
}

// Adds a panel to the registry, referenced by its websocket pointer.
func (registry *ScoringPanelRegistry) RegisterPanel(alliance string, ws *websocket.Websocket) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.scoringPanels[alliance][ws] = false
}

// Sets the score committed state to true for the given panel, referenced by its websocket pointer.
func (registry *ScoringPanelRegistry) SetScoreCommitted(alliance string, ws *websocket.Websocket) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.scoringPanels[alliance][ws] = true
}

// Removes a panel from the registry, referenced by its websocket pointer.
func (registry *ScoringPanelRegistry) UnregisterPanel(alliance string, ws *websocket.Websocket) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	delete(registry.scoringPanels[alliance], ws)
}
