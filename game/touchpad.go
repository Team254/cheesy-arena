// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2017 touchpad element.

package game

import (
	"time"
)

const (
	NotTriggered = iota
	Triggered
	Held
)

type Touchpad struct {
	lastTriggered   bool
	triggeredTime   *time.Time
	untriggeredTime *time.Time
}

// Updates the internal timing state of the touchpad given the current state of the sensor.
func (touchpad *Touchpad) UpdateState(triggered bool, currentTime time.Time) {
	if triggered && !touchpad.lastTriggered {
		touchpad.triggeredTime = &currentTime
		touchpad.untriggeredTime = nil
	} else if !triggered && touchpad.lastTriggered {
		touchpad.untriggeredTime = &currentTime
	}
	touchpad.lastTriggered = triggered
}

// Determines the scoring status of the touchpad. Returns 0 if not triggered, 1 if triggered but not yet for a full
// second, and 2 if triggered and counting for points.
func (touchpad *Touchpad) GetState(matchStartTime, currentTime time.Time) int {
	matchEndTime := GetMatchEndTime(matchStartTime)

	if touchpad.triggeredTime != nil && touchpad.triggeredTime.Before(matchEndTime) {
		if touchpad.untriggeredTime == nil {
			if currentTime.Sub(*touchpad.triggeredTime) >= time.Second {
				return Held
			} else {
				return Triggered
			}
		} else if touchpad.untriggeredTime.Sub(*touchpad.triggeredTime) >= time.Second &&
			touchpad.untriggeredTime.After(matchEndTime) {
			return Held
		}
	}

	return NotTriggered
}

func CountTouchpads(touchpads *[3]Touchpad, matchStartTime, currentTime time.Time) int {
	matchEndTime := GetMatchEndTime(matchStartTime)

	count := 0
	for _, touchpad := range touchpads {
		if touchpad.GetState(matchEndTime, currentTime) == 2 {
			count++
		}
	}

	return count
}
