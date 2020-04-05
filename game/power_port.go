// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2020 Power Port element.

package game

import (
	"time"
)

type PowerPort struct {
	AutoCellsBottom   [2]int
	AutoCellsOuter    [2]int
	AutoCellsInner    [2]int
	TeleopCellsBottom [4]int
	TeleopCellsOuter  [4]int
	TeleopCellsInner  [4]int
}

// Updates the internal counting state of the power port given the current state of the hardware counts. Allows the
// score to accumulate before the match, since the counters will be reset in hardware.
func (powerPort *PowerPort) UpdateState(portCells [3]int, stage Stage, matchStartTime, currentTime time.Time) {
	autoValidityDuration := GetDurationToAutoEnd() + powerPortAutoGracePeriodSec*time.Second
	autoValidityCutoff := matchStartTime.Add(autoValidityDuration)
	teleopValidityDuration := GetDurationToTeleopEnd() + PowerPortTeleopGracePeriodSec*time.Second
	teleopValidityCutoff := matchStartTime.Add(teleopValidityDuration)

	newBottomCells := portCells[0] - totalPortCells(powerPort.AutoCellsBottom, powerPort.TeleopCellsBottom)
	newOuterCells := portCells[1] - totalPortCells(powerPort.AutoCellsOuter, powerPort.TeleopCellsOuter)
	newInnerCells := portCells[2] - totalPortCells(powerPort.AutoCellsInner, powerPort.TeleopCellsInner)

	if currentTime.Before(autoValidityCutoff) && stage <= Stage2 {
		powerPort.AutoCellsBottom[stage] += newBottomCells
		powerPort.AutoCellsOuter[stage] += newOuterCells
		powerPort.AutoCellsInner[stage] += newInnerCells
	} else if currentTime.Before(teleopValidityCutoff) {
		powerPort.TeleopCellsBottom[stage] += newBottomCells
		powerPort.TeleopCellsOuter[stage] += newOuterCells
		powerPort.TeleopCellsInner[stage] += newInnerCells
	}
}

// Returns the total number of cells scored across all stages in a port level.
func totalPortCells(autoCells [2]int, teleopCells [4]int) int {
	var total int
	for _, stageCount := range autoCells {
		total += stageCount
	}
	for _, stageCount := range teleopCells {
		total += stageCount
	}
	return total
}
