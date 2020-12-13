// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var matchStartTime = time.Unix(10, 0)

func TestPowerPort(t *testing.T) {
	var powerPort PowerPort
	assertPowerPort(t, [3][2]int{}, [3][4]int{}, &powerPort)

	// Check before match start and during the autonomous period.
	powerPort.UpdateState([3]int{0, 1, 2}, Stage1, matchStartTime, timeAfterStart(-1))
	assertPowerPort(t, [3][2]int{{0, 0}, {1, 0}, {2, 0}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{0, 0, 0}, Stage1, matchStartTime, timeAfterStart(1))
	assertPowerPort(t, [3][2]int{{0, 0}, {0, 0}, {0, 0}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{0, 1, 2}, Stage1, matchStartTime, timeAfterStart(2))
	assertPowerPort(t, [3][2]int{{0, 0}, {1, 0}, {2, 0}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{3, 5, 2}, Stage1, matchStartTime, timeAfterStart(5))
	assertPowerPort(t, [3][2]int{{3, 0}, {5, 0}, {2, 0}}, [3][4]int{}, &powerPort)

	// Check boundary conditions around the auto end grace period.
	powerPort.UpdateState([3]int{4, 6, 3}, Stage1, matchStartTime, timeAfterStart(16.9))
	assertPowerPort(t, [3][2]int{{4, 0}, {6, 0}, {3, 0}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{5, 8, 6}, Stage2, matchStartTime, timeAfterStart(17.1))
	assertPowerPort(t, [3][2]int{{4, 1}, {6, 2}, {3, 3}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{8, 10, 7}, Stage2, matchStartTime, timeAfterStart(19.9))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{}, &powerPort)
	powerPort.UpdateState([3]int{8, 10, 8}, Stage2, matchStartTime, timeAfterStart(20.1))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 1, 0, 0}},
		&powerPort)

	// Check during the teleoperated period.
	powerPort.UpdateState([3]int{9, 10, 8}, Stage1, matchStartTime, timeAfterStart(30))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{1, 0, 0, 0}, {0, 0, 0, 0}, {0, 1, 0, 0}},
		&powerPort)
	powerPort.UpdateState([3]int{10, 12, 11}, Stage3, matchStartTime, timeAfterStart(30))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{1, 0, 1, 0}, {0, 0, 2, 0}, {0, 1, 3, 0}},
		&powerPort)
	powerPort.UpdateState([3]int{40, 32, 21}, StageExtra, matchStartTime, timeAfterStart(60))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{1, 0, 1, 30}, {0, 0, 2, 20}, {0, 1, 3, 10}},
		&powerPort)

	// Check boundary conditions around the teleop end grace period.
	powerPort.UpdateState([3]int{41, 32, 21}, StageExtra, matchStartTime, timeAfterStart(156.9))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{1, 0, 1, 31}, {0, 0, 2, 20}, {0, 1, 3, 10}},
		&powerPort)
	powerPort.UpdateState([3]int{42, 33, 22}, StageExtra, matchStartTime, timeAfterStart(157.1))
	assertPowerPort(t, [3][2]int{{4, 4}, {6, 4}, {3, 4}}, [3][4]int{{1, 0, 1, 31}, {0, 0, 2, 20}, {0, 1, 3, 10}},
		&powerPort)
}

func assertPowerPort(t *testing.T, expectedAutoCells [3][2]int, expectedTeleopCells [3][4]int, powerPort *PowerPort) {
	assert.Equal(t, expectedAutoCells[0], powerPort.AutoCellsBottom)
	assert.Equal(t, expectedAutoCells[1], powerPort.AutoCellsOuter)
	assert.Equal(t, expectedAutoCells[2], powerPort.AutoCellsInner)
	assert.Equal(t, expectedTeleopCells[0], powerPort.TeleopCellsBottom)
	assert.Equal(t, expectedTeleopCells[1], powerPort.TeleopCellsOuter)
	assert.Equal(t, expectedTeleopCells[2], powerPort.TeleopCellsInner)
}

func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}
