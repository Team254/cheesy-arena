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
func (touchpad *Touchpad) UpdateState(triggered bool, matchStartTime, currentTime time.Time) {
	matchEndTime := GetMatchEndTime(matchStartTime)

	if triggered && !touchpad.lastTriggered && currentTime.Before(matchEndTime) {
		touchpad.triggeredTime = &currentTime
		touchpad.untriggeredTime = nil
	} else if !triggered && touchpad.lastTriggered {
		if currentTime.Before(matchEndTime) || touchpad.GetState(currentTime) == Triggered {
			touchpad.triggeredTime = nil
		}
		touchpad.untriggeredTime = &currentTime
	}
	touchpad.lastTriggered = triggered
}

// Determines the scoring status of the touchpad. Returns 0 if not triggered, 1 if triggered but not yet for a full
// second, and 2 if triggered and counting for points.
func (touchpad *Touchpad) GetState(currentTime time.Time) int {
	if touchpad.triggeredTime != nil {
		if touchpad.untriggeredTime != nil {
			if touchpad.untriggeredTime.Sub(*touchpad.triggeredTime) >= time.Second {
				return Held
			} else {
				return NotTriggered
			}
		}
		if currentTime.Sub(*touchpad.triggeredTime) >= time.Second {
			return Held
		} else {
			return Triggered
		}
	}

	return NotTriggered
}

func CountTouchpads(touchpads *[3]Touchpad, currentTime time.Time) int {
	count := 0
	for _, touchpad := range touchpads {
		if touchpad.GetState(currentTime) == Held {
			count++
		}
	}

	return count
}
