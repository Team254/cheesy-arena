// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2018 scale and switch elements.

package game

import (
	"time"
)

type Alliance int

const (
	NeitherAlliance Alliance = iota
	RedAlliance
	BlueAlliance
)

type Seesaw struct {
	Kind       Alliance // Red or blue indicates that it is a switch; neither indicates the scale.
	NearIsRed  bool
	ownerships []*Ownership
}

type Ownership struct {
	seesaw    *Seesaw
	ownedBy   Alliance
	startTime time.Time
	endTime   *time.Time
}

// Updates the internal timing state of the scale or switch given the current state of the sensors. Returns true if
// ownership has changed since the last cycle.
func (seesaw *Seesaw) UpdateState(state [2]bool, currentTime time.Time) bool {
	ownedBy := NeitherAlliance

	// Check if there is an active force power up for this seesaw.
	currentPowerUp := GetActivePowerUp(currentTime)
	if currentPowerUp != nil && currentPowerUp.Effect == Force &&
		(seesaw.Kind == NeitherAlliance && currentPowerUp.Level >= 2 ||
			(seesaw.Kind == currentPowerUp.Alliance && (currentPowerUp.Level == 1 || currentPowerUp.Level == 3))) {
		ownedBy = currentPowerUp.Alliance
	} else {
		// Determine current ownership from sensor state.
		if state[0] && !state[1] && seesaw.NearIsRed || state[1] && !state[0] && !seesaw.NearIsRed {
			ownedBy = RedAlliance
		} else if state[0] && !state[1] && !seesaw.NearIsRed || state[1] && !state[0] && seesaw.NearIsRed {
			ownedBy = BlueAlliance
		}
	}

	// Update data if ownership has changed since last cycle.
	currentOwnership := seesaw.getCurrentOwnership()
	if currentOwnership != nil && ownedBy != currentOwnership.ownedBy ||
		currentOwnership == nil && ownedBy != NeitherAlliance {
		if currentOwnership != nil {
			currentOwnership.endTime = &currentTime
		}

		if ownedBy != NeitherAlliance {
			newOwnership := &Ownership{seesaw: seesaw, ownedBy: ownedBy, startTime: currentTime}
			seesaw.ownerships = append(seesaw.ownerships, newOwnership)
		}
		return true
	}
	return false
}

func (seesaw *Seesaw) GetOwnedBy() Alliance {
	ownership := seesaw.getCurrentOwnership()
	if ownership == nil {
		return NeitherAlliance
	} else {
		return ownership.ownedBy
	}
}

// GetRedSeconds returns the total seconds of ownership and boost score accumulation for the red alliance.
func (seesaw *Seesaw) GetRedSeconds(startTime, endTime time.Time) (float64, float64) {
	return seesaw.getAllianceSeconds(RedAlliance, startTime, endTime)
}

// GetBlueSeconds returns the total seconds of ownership and boost score accumulation for the blue alliance.
func (seesaw *Seesaw) GetBlueSeconds(startTime, endTime time.Time) (float64, float64) {
	return seesaw.getAllianceSeconds(BlueAlliance, startTime, endTime)
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

func (seesaw *Seesaw) getAllianceSeconds(ownedBy Alliance, startTime, endTime time.Time) (float64, float64) {
	var ownershipSec, boostSec float64
	for _, ownership := range seesaw.ownerships {
		if ownership.ownedBy == ownedBy {
			ownership, boost := ownership.getSeconds(startTime, endTime)
			ownershipSec += ownership
			boostSec += boost
		}
	}
	return ownershipSec, boostSec
}

// Returns the regular and boost scoring values for the ownership period, whether it is past or current.
func (ownership *Ownership) getSeconds(startTime, endTime time.Time) (float64, float64) {
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
		return 0, 0
	}
	ownershipSec := ownershipEndTime.Sub(ownershipStartTime).Seconds()

	// Find the boost power up applicable to this seesaw and alliance, if it exists.
	var boostPowerUp *PowerUp
	for _, powerUp := range powerUpUses {
		if powerUp.Effect == Boost && ownership.ownedBy == powerUp.Alliance {
			if ownership.seesaw.Kind == NeitherAlliance && powerUp.Level >= 2 ||
				ownership.seesaw.Kind != NeitherAlliance && (powerUp.Level == 1 || powerUp.Level == 3) {
				boostPowerUp = powerUp
				break
			}
		}
	}

	var boostSec float64
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
			boostSec = boostEndTime.Sub(boostStartTime).Seconds()
		}
	}

	return ownershipSec, boostSec
}
