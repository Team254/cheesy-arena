// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and functions for reporting on event status.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"math"
	"time"
)

const maxExpectedCycleTimeSec = 900

type EventStatus struct {
	CycleTime                   string
	EarlyLateMessage            string
	lastMatchStartTime          time.Time
	lastMatchScheduledStartTime time.Time
}

// Calculates the last cycle time and publishes an update to the displays that show it.
func (arena *Arena) updateCycleTime(matchStartTime time.Time) {
	expectedCycleTimeSec := arena.CurrentMatch.Time.Sub(arena.EventStatus.lastMatchScheduledStartTime).Seconds()
	if arena.EventStatus.lastMatchStartTime.IsZero() || expectedCycleTimeSec > maxExpectedCycleTimeSec {
		// We don't know when the previous match was started or there was a big gap that we shouldn't count.
		arena.EventStatus.CycleTime = ""
	} else {
		cycleTimeSec := int(matchStartTime.Sub(arena.EventStatus.lastMatchStartTime).Seconds())
		hours := cycleTimeSec / 3600
		minutes := cycleTimeSec % 3600 / 60
		seconds := cycleTimeSec % 60
		if hours > 0 {
			arena.EventStatus.CycleTime = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			arena.EventStatus.CycleTime = fmt.Sprintf("%d:%02d", minutes, seconds)
		}

		deltaSec := cycleTimeSec - int(expectedCycleTimeSec)
		var direction string
		if deltaSec > 0 {
			direction = "slower"
		} else {
			direction = "faster"
			deltaSec = -deltaSec
		}
		arena.EventStatus.CycleTime += fmt.Sprintf(
			" (%d:%02d %s than scheduled)", deltaSec/60, deltaSec%60, direction,
		)
	}
	arena.EventStatus.lastMatchStartTime = matchStartTime
	arena.EventStatus.lastMatchScheduledStartTime = arena.CurrentMatch.Time
	arena.EventStatusNotifier.Notify()
}

// Checks how early or late the event is running and publishes an update to the displays that show it.
func (arena *Arena) updateEarlyLateMessage() {
	newEarlyLateMessage := arena.getEarlyLateMessage()
	if newEarlyLateMessage != arena.EventStatus.EarlyLateMessage {
		arena.EventStatus.EarlyLateMessage = newEarlyLateMessage
		arena.EventStatusNotifier.Notify()
	}
}

// Updates the string that indicates how early or late the event is running.
func (arena *Arena) getEarlyLateMessage() string {
	currentMatch := arena.CurrentMatch
	if currentMatch.Type == model.Test {
		return ""
	}
	if currentMatch.IsComplete() {
		// This is a replay or otherwise unpredictable situation.
		return ""
	}

	var minutesLate float64
	if arena.MatchState > PreMatch && arena.MatchState < PostMatch {
		// The match is in progress; simply calculate lateness from its start time.
		minutesLate = currentMatch.StartedAt.Sub(currentMatch.Time).Minutes()
	} else {
		// We need to check the adjacent matches to accurately determine lateness.
		matches, _ := arena.Database.GetMatchesByType(currentMatch.Type, false)

		previousMatchIndex := -1
		nextMatchIndex := len(matches)
		for i, match := range matches {
			if match.Id == currentMatch.Id {
				previousMatchIndex = i - 1
				nextMatchIndex = i + 1
				break
			}
		}

		if arena.MatchState == PreMatch || arena.MatchState == TimeoutActive || arena.MatchState == PostTimeout {
			currentMinutesLate := time.Now().Sub(currentMatch.Time).Minutes()
			if previousMatchIndex >= 0 &&
				currentMatch.Time.Sub(matches[previousMatchIndex].Time).Minutes() <= MaxMatchGapMin {
				previousMatch := matches[previousMatchIndex]
				previousMinutesLate := previousMatch.StartedAt.Sub(previousMatch.Time).Minutes()
				minutesLate = math.Max(previousMinutesLate, currentMinutesLate)
			} else {
				minutesLate = math.Max(currentMinutesLate, 0)
			}
		} else if arena.MatchState == PostMatch {
			currentMinutesLate := currentMatch.StartedAt.Sub(currentMatch.Time).Minutes()
			if nextMatchIndex < len(matches) {
				nextMatch := matches[nextMatchIndex]
				nextMinutesLate := time.Now().Sub(nextMatch.Time).Minutes()
				minutesLate = math.Max(currentMinutesLate, nextMinutesLate)
			} else {
				minutesLate = currentMinutesLate
			}
		}
	}

	if minutesLate > earlyLateThresholdMin {
		return fmt.Sprintf("Event is running %d minutes late", int(minutesLate))
	} else if minutesLate < -earlyLateThresholdMin {
		return fmt.Sprintf("Event is running %d minutes early", int(-minutesLate))
	}
	return "Event is running on schedule"
}
