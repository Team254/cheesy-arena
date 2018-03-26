// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var matchStartTime = time.Unix(10, 0)

func TestPowerUpGetState(t *testing.T) {
	powerUp := PowerUp{startTime: timeAfterStart(30)}
	assert.Equal(t, queued, powerUp.GetState(timeAfterStart(25)))
	assert.Equal(t, queued, powerUp.GetState(timeAfterStart(29.9)))
	assert.Equal(t, active, powerUp.GetState(timeAfterStart(30.1)))
	assert.Equal(t, active, powerUp.GetState(timeAfterStart(39.9)))
	assert.Equal(t, expired, powerUp.GetState(timeAfterStart(40.1)))
}

func TestPowerUpActivate(t *testing.T) {
	powerUp1 := new(PowerUp)
	if assert.NotNil(t, maybeActivatePowerUp(powerUp1, timeAfterStart(30))) {
		assert.Equal(t, timeAfterStart(30), powerUp1.startTime)
	}

	powerUp2 := new(PowerUp)
	if assert.NotNil(t, maybeActivatePowerUp(powerUp2, timeAfterStart(45))) {
		assert.Equal(t, timeAfterStart(45), powerUp2.startTime)
	}

	assert.Nil(t, getActivePowerUp(timeAfterStart(29.9)))
	assert.Equal(t, powerUp1, getActivePowerUp(timeAfterStart(30.1)))
	assert.Equal(t, powerUp1, getActivePowerUp(timeAfterStart(39.9)))
	assert.Nil(t, getActivePowerUp(timeAfterStart(42)))
	assert.Equal(t, powerUp2, getActivePowerUp(timeAfterStart(45.1)))
	assert.Equal(t, powerUp2, getActivePowerUp(timeAfterStart(54.9)))
	assert.Nil(t, getActivePowerUp(timeAfterStart(55.1)))
}

func TestPowerUpQueue(t *testing.T) {
	powerUp1 := &PowerUp{alliance: redAlliance}
	maybeActivatePowerUp(powerUp1, timeAfterStart(60))

	powerUp2 := &PowerUp{alliance: redAlliance}
	assert.Nil(t, maybeActivatePowerUp(powerUp2, timeAfterStart(65)))
	powerUp2.alliance = blueAlliance
	if assert.NotNil(t, maybeActivatePowerUp(powerUp2, timeAfterStart(65))) {
		assert.Equal(t, timeAfterStart(70), powerUp2.startTime)
	}

	assert.Equal(t, powerUp1, getActivePowerUp(timeAfterStart(69.9)))
	assert.Equal(t, powerUp2, getActivePowerUp(timeAfterStart(70.1)))
}

func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}

func timeAfterEnd(sec float32) time.Time {
	matchDuration := time.Duration(MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
	return matchStartTime.Add(matchDuration).Add(time.Duration(1000*sec) * time.Millisecond)
}
