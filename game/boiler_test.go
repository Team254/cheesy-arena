// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var matchStartTime = time.Unix(10, 0)

func TestFuelBeforeMatch(t *testing.T) {
	boiler := Boiler{}

	boiler.UpdateState(1, 2, matchStartTime, timeAfterStart(-1))
	checkBoilerCounts(t, 1, 2, 0, 0, &boiler)
}

func TestAutoFuel(t *testing.T) {
	boiler := Boiler{}

	boiler.UpdateState(3, 4, matchStartTime, timeAfterStart(1))
	checkBoilerCounts(t, 3, 4, 0, 0, &boiler)
	boiler.UpdateState(5, 6, matchStartTime, timeAfterStart(10))
	checkBoilerCounts(t, 5, 6, 0, 0, &boiler)
	boiler.UpdateState(7, 8, matchStartTime, timeAfterStart(19.9))
	checkBoilerCounts(t, 7, 8, 0, 0, &boiler)
	boiler.UpdateState(9, 10, matchStartTime, timeAfterStart(20.1))
	checkBoilerCounts(t, 7, 8, 9, 10, &boiler)
}

func TestTeleopFuel(t *testing.T) {
	boiler := Boiler{}

	boiler.UpdateState(1, 2, matchStartTime, timeAfterStart(1))
	boiler.UpdateState(3, 4, matchStartTime, timeAfterStart(21))
	checkBoilerCounts(t, 1, 2, 3, 4, &boiler)
	boiler.UpdateState(5, 6, matchStartTime, timeAfterStart(120))
	checkBoilerCounts(t, 1, 2, 5, 6, &boiler)
	boiler.UpdateState(7, 8, matchStartTime, timeAfterEnd(-1))
	checkBoilerCounts(t, 1, 2, 7, 8, &boiler)
	boiler.UpdateState(9, 10, matchStartTime, timeAfterEnd(4.9))
	checkBoilerCounts(t, 1, 2, 9, 10, &boiler)
	boiler.UpdateState(11, 12, matchStartTime, timeAfterEnd(5.1))
	checkBoilerCounts(t, 1, 2, 9, 10, &boiler)
}

func checkBoilerCounts(t *testing.T, autoLow, autoHigh, low, high int, boiler *Boiler) {
	assert.Equal(t, autoLow, boiler.AutoFuelLow)
	assert.Equal(t, autoHigh, boiler.AutoFuelHigh)
	assert.Equal(t, low, boiler.FuelLow)
	assert.Equal(t, high, boiler.FuelHigh)
}
func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}

func timeAfterEnd(sec float32) time.Time {
	matchDuration := time.Duration(MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
	return matchStartTime.Add(matchDuration).Add(time.Duration(1000*sec) * time.Millisecond)
}
