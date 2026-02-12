// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Logic for creating judging schedules.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"math/rand"
	"time"
)

// JudgingScheduleParams contains configuration parameters for the judging schedule generation.
type JudgingScheduleParams struct {
	// NumJudges is the number of judge teams operating in parallel.
	NumJudges int

	// DurationMinutes is the duration of each judging slot in minutes.
	DurationMinutes int

	// PreviousSpacingMinutes is the minimum buffer time in minutes between the start of a team's match and when they
	// can be scheduled for judging.
	PreviousSpacingMinutes int

	// NextSpacingMinutes is the minimum buffer time in minutes between the end of a team's judging slot and their next
	// scheduled match.
	NextSpacingMinutes int
}

// judgeSchedule represents the schedule of a judge team, with a list of judging slots and the ending time of the last
// slot.
type judgeSchedule struct {
	judgeNumber int
	endTime     time.Time
	slots       []*model.JudgingSlot
}

// BuildJudgingSchedule generates a judging schedule based on the given parameters and qualification match schedule.
func BuildJudgingSchedule(database *model.Database, params JudgingScheduleParams) error {
	slots, err := database.GetAllJudgingSlots()
	if err != nil {
		return fmt.Errorf("error getting judging slots: %v", err)
	}
	if len(slots) > 0 {
		return fmt.Errorf("cannot generate judging schedule: existing judging slots found")
	}

	teams, err := database.GetAllTeams()
	if err != nil {
		return fmt.Errorf("error getting teams: %v", err)
	}
	if len(teams) == 0 {
		return fmt.Errorf("cannot generate judging schedule: no teams present")
	}

	matches, err := database.GetMatchesByType(model.Qualification, true)
	if err != nil {
		return fmt.Errorf("error getting qualification matches: %v", err)
	}
	if len(matches) < 2 {
		return fmt.Errorf("cannot generate judging schedule: no qualification matches found")
	}

	scheduleBlocks, err := database.GetScheduleBlocksByMatchType(model.Qualification)
	if err != nil {
		return fmt.Errorf("error getting schedule blocks: %v", err)
	}

	// Create a map of teams to their matches.
	teamMatches := createTeamMatchMap(teams, matches)

	// Assume that the second match is the start time for the judging schedule.
	startTime := matches[1].Time

	// Initialize judging team schedules.
	judgeSchedules := make([]*judgeSchedule, params.NumJudges)
	for i := 0; i < params.NumJudges; i++ {
		judgeSchedules[i] = &judgeSchedule{
			judgeNumber: i + 1,
			endTime:     startTime,
			slots:       []*model.JudgingSlot{},
		}
	}

	if params.NumJudges <= 0 {
		return fmt.Errorf("cannot generate judging schedule: no judges available")
	}

	// Randomly shuffle the teams to avoid bias in the scheduling.
	rand.Shuffle(
		len(teams), func(i, j int) {
			teams[i], teams[j] = teams[j], teams[i]
		},
	)

	// Loop until all teams have been scheduled.
	scheduledTeams := make(map[int]struct{})
	noProgressCount := 0
	maxNoProgress := len(teams) * params.NumJudges * 5
	for len(scheduledTeams) < len(teams) {
		// Select the judge with fewest scheduled visits (or first if there are multiple).
		var selectedJudge *judgeSchedule
		for _, judge := range judgeSchedules {
			if selectedJudge == nil || len(judge.slots) < len(selectedJudge.slots) {
				selectedJudge = judge
			}
		}
		if selectedJudge == nil {
			return fmt.Errorf("no available judges to schedule")
		}

		candidateTime := selectedJudge.endTime
		var selectedSlot *model.JudgingSlot
		for _, team := range teams {
			if _, ok := scheduledTeams[team.Id]; ok {
				continue
			}

			slot, err := getNextSlotForTeam(team, candidateTime, teamMatches[team.Id], params)
			if err != nil {
				return fmt.Errorf("error finding next slot for team %d: %v", team.Id, err)
			}
			if selectedSlot == nil || slot.Time.Before(selectedSlot.Time) {
				selectedSlot = slot
			}
			if slot.Time == candidateTime {
				// The slot perfectly matches the candidate time; no need to evaluate the remaining teams.
				break
			}
		}
		if selectedSlot == nil {
			return fmt.Errorf("no available judging slot found")
		}

		// Check the validity of the selected slot with respect to the scheduled breaks.
		slotEndTime := selectedSlot.Time.Add(time.Duration(params.DurationMinutes) * time.Minute)
		validAssignment := true
		for i, block := range scheduleBlocks {
			blockEndTime := block.StartTime.Add(time.Duration(block.NumMatches*block.MatchSpacingSec) * time.Second)
			if selectedSlot.Time.Before(block.StartTime) {
				// The slot time falls between blocks; advance the judge's end time to the start of the next block.
				selectedJudge.endTime = block.StartTime
				// Don't allow a slot to start during a break, but do allow one to end during a break.
				validAssignment = false
				break
			}
			if selectedSlot.Time.Before(blockEndTime) {
				// The slot starts within the block.
				if slotEndTime.After(blockEndTime) && i+1 < len(scheduleBlocks) {
					nextBlockStart := scheduleBlocks[i+1].StartTime
					if slotEndTime.After(nextBlockStart) {
						// The slot runs into the next block; advance to the next block start.
						selectedJudge.endTime = nextBlockStart
						validAssignment = false
					}
				}
				break
			}
		}
		if !validAssignment {
			// The slot time is invalid; try the next judge.
			noProgressCount++
			if noProgressCount >= maxNoProgress {
				judgeEndTimes := make([]time.Time, len(judgeSchedules))
				for i, judge := range judgeSchedules {
					judgeEndTimes[i] = judge.endTime
				}
				return fmt.Errorf(
					"cannot generate judging schedule: no progress after %d attempts (scheduled %d/%d, candidate %s, judgeEndTimes %v, params %+v)",
					noProgressCount,
					len(scheduledTeams),
					len(teams),
					candidateTime.Format(time.RFC3339),
					judgeEndTimes,
					params,
				)
			}
			continue
		}

		// Update the schedule.
		selectedSlot.JudgeNumber = selectedJudge.judgeNumber
		selectedJudge.slots = append(selectedJudge.slots, selectedSlot)
		selectedJudge.endTime = selectedSlot.Time.Add(time.Duration(params.DurationMinutes) * time.Minute)
		scheduledTeams[selectedSlot.TeamId] = struct{}{}

		if err := database.CreateJudgingSlot(selectedSlot); err != nil {
			return fmt.Errorf("error saving judging slot for team %d: %v", selectedSlot.TeamId, err)
		}
		noProgressCount = 0
	}

	return nil
}

// createTeamMatchMap creates a map of team IDs to their scheduled qualification matches.
func createTeamMatchMap(teams []model.Team, matches []model.Match) map[int][]model.Match {
	teamMatches := make(map[int][]model.Match)
	for _, team := range teams {
		teamMatches[team.Id] = []model.Match{}
	}

	for _, match := range matches {
		teamMatches[match.Red1] = append(teamMatches[match.Red1], match)
		teamMatches[match.Red2] = append(teamMatches[match.Red2], match)
		teamMatches[match.Red3] = append(teamMatches[match.Red3], match)
		teamMatches[match.Blue1] = append(teamMatches[match.Blue1], match)
		teamMatches[match.Blue2] = append(teamMatches[match.Blue2], match)
		teamMatches[match.Blue3] = append(teamMatches[match.Blue3], match)
	}

	return teamMatches
}

// getNextSlotForTeam finds the next available judging slot for a team at or after the given candidate time.
func getNextSlotForTeam(
	team model.Team,
	candidateTime time.Time,
	matches []model.Match,
	params JudgingScheduleParams,
) (*model.JudgingSlot, error) {
	if len(matches) == 0 {
		return nil, fmt.Errorf("no qualification matches for team")
	}

	var previousMatch *model.Match
	for i := range matches {
		match := matches[i]
		if match.Time.After(candidateTime) {
			// Calculate the spacing between the candidate time and the previous match.
			previousSpacingMinutes := float64(params.PreviousSpacingMinutes)
			if previousMatch != nil {
				previousSpacingMinutes = candidateTime.Sub(previousMatch.Time).Minutes()
			}
			if previousSpacingMinutes < float64(params.PreviousSpacingMinutes) {
				// The candidate time is too close to the previous match; adjust it minimally.
				candidateTime = previousMatch.Time.Add(time.Duration(params.PreviousSpacingMinutes) * time.Minute)
			}

			nextSpacingMinutes := match.Time.Sub(candidateTime).Minutes() - float64(params.DurationMinutes)
			if nextSpacingMinutes >= float64(params.NextSpacingMinutes) {
				// The candidate time is far enough from the next match; schedule the judging slot.
				slot := model.JudgingSlot{
					Time:            candidateTime,
					TeamId:          team.Id,
					NextMatchNumber: match.TypeOrder,
					NextMatchTime:   match.Time,
				}
				if previousMatch != nil {
					slot.PreviousMatchNumber = previousMatch.TypeOrder
					slot.PreviousMatchTime = previousMatch.Time
				}
				return &slot, nil
			}

			// The candidate time is too close to the next match; continue searching.
		}
		previousMatch = &matches[i]
	}

	// If we get here, the team can only be scheduled once all matches are complete.
	if previousMatch == nil {
		return nil, fmt.Errorf("no previous match found for team")
	}
	minCandidateTime := previousMatch.Time.Add(time.Duration(params.PreviousSpacingMinutes) * time.Minute)
	if candidateTime.Before(minCandidateTime) {
		candidateTime = minCandidateTime
	}
	slot := model.JudgingSlot{
		Time:                candidateTime,
		TeamId:              team.Id,
		PreviousMatchNumber: previousMatch.TypeOrder,
		PreviousMatchTime:   previousMatch.Time,
	}
	return &slot, nil
}
