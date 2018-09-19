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
	assert.Equal(t, Queued, powerUp.GetState(timeAfterStart(25)))
	assert.Equal(t, Queued, powerUp.GetState(timeAfterStart(29.9)))
	assert.Equal(t, Active, powerUp.GetState(timeAfterStart(30.1)))
	assert.Equal(t, Active, powerUp.GetState(timeAfterStart(39.9)))
	assert.Equal(t, Expired, powerUp.GetState(timeAfterStart(40.1)))
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

	assert.Nil(t, GetActivePowerUp(timeAfterStart(29.9)))
	assert.Equal(t, powerUp1, GetActivePowerUp(timeAfterStart(30.1)))
	assert.Equal(t, powerUp1, GetActivePowerUp(timeAfterStart(39.9)))
	assert.Nil(t, GetActivePowerUp(timeAfterStart(42)))
	assert.Equal(t, powerUp2, GetActivePowerUp(timeAfterStart(45.1)))
	assert.Equal(t, powerUp2, GetActivePowerUp(timeAfterStart(54.9)))
	assert.Nil(t, GetActivePowerUp(timeAfterStart(55.1)))
}

func TestPowerUpQueue(t *testing.T) {
	ResetPowerUps()

	powerUp1 := &PowerUp{Alliance: RedAlliance}
	assert.NotNil(t, maybeActivatePowerUp(powerUp1, timeAfterStart(60)))

	powerUp2 := &PowerUp{Alliance: RedAlliance}
	assert.Nil(t, maybeActivatePowerUp(powerUp2, timeAfterStart(65)))
	powerUp2.Alliance = BlueAlliance
	if assert.NotNil(t, maybeActivatePowerUp(powerUp2, timeAfterStart(65))) {
		assert.Equal(t, timeAfterStart(70), powerUp2.startTime)
	}

	powerUp3 := &PowerUp{Alliance: RedAlliance}
	assert.NotNil(t, maybeActivatePowerUp(powerUp3, timeAfterStart(81)))

	assert.Equal(t, powerUp1, GetActivePowerUp(timeAfterStart(69.9)))
	assert.Equal(t, powerUp2, GetActivePowerUp(timeAfterStart(70.1)))
}

func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}

func timeAfterEnd(sec float32) time.Time {
	matchDuration := time.Duration(MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
	return matchStartTime.Add(matchDuration).Add(time.Duration(1000*sec) * time.Millisecond)
}
