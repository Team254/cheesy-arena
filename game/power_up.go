// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2018 power ups.

package game

import (
	"time"
)

const powerUpDurationSec = 10

// Power up type/effect enum.
type effect int

const (
	force effect = iota
	boost
)

// Power up state enum.
type PowerUpState int

const (
	Unplayed PowerUpState = iota
	Queued
	Active
	Expired
)

type PowerUp struct {
	Alliance
	effect
	level     int
	startTime time.Time
}

var powerUpUses []*PowerUp

func ResetPowerUps() {
	powerUpUses = powerUpUses[:0]
}

func (powerUp *PowerUp) GetState(currentTime time.Time) PowerUpState {
	if powerUp.startTime.After(currentTime) {
		return Queued
	}
	if powerUp.getEndTime().After(currentTime) {
		return Active
	}
	return Expired
}

func (powerUp *PowerUp) getEndTime() time.Time {
	return powerUp.startTime.Add(powerUpDurationSec * time.Second)
}

// Returns the current active power up, or nil if there isn't one.
func getActivePowerUp(currentTime time.Time) *PowerUp {
	for _, powerUp := range powerUpUses {
		if powerUp.GetState(currentTime) == Active {
			return powerUp
		}
	}
	return nil
}

// Activates the given power up if it can be played, or if not, queues it if the active power up belongs to the other
// alliance. Returns the power up if successful and nil if it cannot be played.
func maybeActivatePowerUp(powerUp *PowerUp, currentTime time.Time) *PowerUp {
	canActivate := false
	if len(powerUpUses) == 0 {
		canActivate = true
		powerUp.startTime = currentTime
	} else {
		lastPowerUp := powerUpUses[len(powerUpUses)-1]
		lastPowerUpState := lastPowerUp.GetState(currentTime)
		if lastPowerUpState == Expired {
			canActivate = true
			powerUp.startTime = currentTime
		} else if lastPowerUpState == Active && lastPowerUp.Alliance != powerUp.Alliance {
			canActivate = true
			powerUp.startTime = lastPowerUp.getEndTime()
		}
	}

	if canActivate {
		powerUpUses = append(powerUpUses, powerUp)
		return powerUp
	}

	return nil
}
