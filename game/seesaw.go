// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2018 scale and switch elements.

package game

import (
	"time"
)

const (
	neitherAlliance = iota
	redAlliance
	blueAlliance
)

type Seesaw struct {
	kind       int
	nearIsRed  bool
	ownerships []*Ownership
}

type Ownership struct {
	seesaw    *Seesaw
	ownedBy   int
	startTime time.Time
	endTime   *time.Time
}

// Sets which side of the scale or switch belongs to which alliance. A value of true indicates that the side nearest the
// scoring table is red.
func (seesaw *Seesaw) SetRandomization(nearIsRed bool) {
	seesaw.nearIsRed = nearIsRed
}

// Updates the internal timing state of the scale or switch given the current state of the sensors.
func (seesaw *Seesaw) UpdateState(state [2]bool, currentTime time.Time) {
	ownedBy := neitherAlliance

	// Check if there is an active force power up for this seesaw.
	currentPowerUp := getActivePowerUp(currentTime)
	if currentPowerUp != nil && currentPowerUp.kind == force &&
		(seesaw.kind == neitherAlliance && currentPowerUp.level >= 2 ||
			(seesaw.kind == currentPowerUp.alliance && (currentPowerUp.level == 1 || currentPowerUp.level == 3))) {
		ownedBy = currentPowerUp.alliance
	} else {
		// Determine current ownership from sensor state.
		if state[0] && !state[1] && seesaw.nearIsRed || state[1] && !state[0] && !seesaw.nearIsRed {
			ownedBy = redAlliance
		} else if state[0] && !state[1] && !seesaw.nearIsRed || state[1] && !state[0] && seesaw.nearIsRed {
			ownedBy = blueAlliance
		}
	}

	// Update data if ownership has changed since last cycle.
	currentOwnership := seesaw.getCurrentOwnership()
	if currentOwnership != nil && ownedBy != currentOwnership.ownedBy ||
		currentOwnership == nil && ownedBy != neitherAlliance {
		if currentOwnership != nil {
			currentOwnership.endTime = &currentTime
		}

		if ownedBy != neitherAlliance {
			newOwnership := &Ownership{seesaw: seesaw, ownedBy: ownedBy, startTime: currentTime}
			seesaw.ownerships = append(seesaw.ownerships, newOwnership)
		}
	}
}

// Returns the auto and teleop period scores for the red alliance.
func (seesaw *Seesaw) GetRedSeconds(startTime, endTime time.Time) float64 {
	return seesaw.getAllianceSeconds(redAlliance, startTime, endTime)
}

// Returns the auto and teleop period scores for the blue alliance.
func (seesaw *Seesaw) GetBlueSeconds(startTime, endTime time.Time) float64 {
	return seesaw.getAllianceSeconds(blueAlliance, startTime, endTime)
}

func (seesaw *Seesaw) getCurrentOwnership() *Ownership {
	if len(seesaw.ownerships) > 0 {
		lastOwnership := seesaw.ownerships[len(seesaw.ownerships)-1]
		if lastOwnership.endTime == nil {
			return lastOwnership
		}
	}
	return nil
}

func (seesaw *Seesaw) getAllianceSeconds(ownedBy int, startTime, endTime time.Time) float64 {
	var seconds float64
	for _, ownership := range seesaw.ownerships {
		if ownership.ownedBy == ownedBy {
			seconds += ownership.getSeconds(startTime, endTime, false)
		}
	}
	return seconds
}

// Returns the scoring value for the ownership period, whether it is past or current.
func (ownership *Ownership) getSeconds(startTime, endTime time.Time, ignoreBoost bool) float64 {
	var ownershipStartTime, ownershipEndTime time.Time
	if ownership.startTime.Before(startTime) {
		ownershipStartTime = startTime
	} else {
		ownershipStartTime = ownership.startTime
	}
	if ownership.endTime == nil || ownership.endTime.After(endTime) {
		ownershipEndTime = endTime
	} else {
		ownershipEndTime = *ownership.endTime
	}

	if ownershipStartTime.After(ownershipEndTime) {
		return 0
	}
	ownershipSeconds := ownershipEndTime.Sub(ownershipStartTime).Seconds()

	// Find the boost power up applicable to this seesaw and alliance, if it exists.
	var boostPowerUp *PowerUp
	for _, powerUp := range powerUpUses {
		if powerUp.kind == boost && ownership.ownedBy == powerUp.alliance {
			if ownership.seesaw.kind == neitherAlliance && powerUp.level >= 2 ||
				ownership.seesaw.kind != neitherAlliance && (powerUp.level == 1 || powerUp.level == 3) {
				boostPowerUp = powerUp
				break
			}
		}
	}

	if boostPowerUp != nil {
		// Adjust for the boost.
		var boostStartTime, boostEndTime time.Time
		if boostPowerUp.startTime.Before(ownershipStartTime) {
			boostStartTime = ownershipStartTime
		} else {
			boostStartTime = boostPowerUp.startTime
		}
		if boostPowerUp.getEndTime().After(ownershipEndTime) {
			boostEndTime = ownershipEndTime
		} else {
			boostEndTime = boostPowerUp.getEndTime()
		}
		if boostEndTime.After(boostStartTime) {
			ownershipSeconds += boostEndTime.Sub(boostStartTime).Seconds()
		}
	}
	return ownershipSeconds
}
