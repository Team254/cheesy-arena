// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for creating and updating the elimination match schedule.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"strconv"
	"time"
)

const ElimMatchSpacingSec = 600

// Incrementally creates any elimination matches that can be created, based on the results of alliance
// selection or prior elimination rounds. Returns true if the tournament is won.
func UpdateEliminationSchedule(database *model.Database, startTime time.Time) (bool, error) {
	alliances, err := database.GetAllAlliances()
	if err != nil {
		return false, err
	}
	winner, err := buildEliminationMatchSet(database, 1, 1, len(alliances))
	if err != nil {
		return false, err
	}

	// Update the scheduled time for all matches that have yet to be run.
	matches, err := database.GetMatchesByType("elimination")
	if err != nil {
		return false, err
	}
	matchIndex := 0
	for _, match := range matches {
		if match.IsComplete() {
			continue
		}
		match.Time = startTime.Add(time.Duration(matchIndex*ElimMatchSpacingSec) * time.Second)
		if err = database.UpdateMatch(&match); err != nil {
			return false, err
		}
		matchIndex++
	}

	return winner != nil, err
}

// Updates the alliance, if necessary, to include whoever played in the match, in case there was a substitute.
func UpdateAlliance(database *model.Database, matchTeamIds [3]int, allianceId int) error {
	alliance, err := database.GetAllianceById(allianceId)
	if err != nil {
		return err
	}

	changed := false
	if matchTeamIds != alliance.Lineup {
		alliance.Lineup = matchTeamIds
		changed = true
	}

	for _, teamId := range matchTeamIds {
		found := false
		for _, allianceTeamId := range alliance.TeamIds {
			if teamId == allianceTeamId {
				found = true
				break
			}
		}
		if !found {
			alliance.TeamIds = append(alliance.TeamIds, teamId)
			changed = true
		}
	}

	if changed {
		return database.UpdateAlliance(alliance)
	}
	return nil
}

// Recursively traverses the elimination bracket downwards, creating matches as necessary. Returns the winner
// of the given round if known.
func buildEliminationMatchSet(
	database *model.Database, round int, group int, numAlliances int,
) (*model.Alliance, error) {
	if numAlliances < 2 {
		return nil, fmt.Errorf("Must have at least 2 alliances")
	}
	roundName, ok := model.ElimRoundNames[round]
	if !ok {
		return nil, fmt.Errorf("Round of depth %d is not supported", round*2)
	}
	if round != 1 {
		roundName += strconv.Itoa(group)
	}

	// Recurse to figure out who the involved alliances are.
	var redAlliance, blueAlliance *model.Alliance
	var err error
	if numAlliances < 4*round {
		// This is the first round for some or all alliances and will be at least partially populated from the
		// alliance selection results.
		matchups := []int{1, 16, 8, 9, 4, 13, 5, 12, 2, 15, 7, 10, 3, 14, 6, 11}
		factor := len(matchups) / round
		redAllianceNumber := matchups[(group-1)*factor]
		blueAllianceNumber := matchups[(group-1)*factor+factor/2]
		numDirectAlliances := 4*round - numAlliances
		if redAllianceNumber <= numDirectAlliances {
			// The red alliance has a bye or the number of alliances is a power of 2; get from alliance selection.
			redAlliance, err = database.GetAllianceById(redAllianceNumber)
			if err != nil {
				return nil, err
			}
		}
		if blueAllianceNumber <= numDirectAlliances {
			// The blue alliance has a bye or the number of alliances is a power of 2; get from alliance selection.
			blueAlliance, err = database.GetAllianceById(blueAllianceNumber)
			if err != nil {
				return nil, err
			}
		}
	}

	// If the alliances aren't known yet, get them from one round down in the bracket.
	if redAlliance == nil {
		redAlliance, err = buildEliminationMatchSet(database, round*2, group*2-1, numAlliances)
		if err != nil {
			return nil, err
		}
	}
	if blueAlliance == nil {
		blueAlliance, err = buildEliminationMatchSet(database, round*2, group*2, numAlliances)
		if err != nil {
			return nil, err
		}
	}

	// Bail if the rounds below are not yet complete and we don't know either alliance competing this round.
	if redAlliance == nil && blueAlliance == nil {
		return nil, nil
	}

	// Check if the match set exists already and if it has been won.
	var redWins, blueWins, numIncomplete int
	var ties []model.Match
	matches, err := database.GetMatchesByElimRoundGroup(round, group)
	if err != nil {
		return nil, err
	}
	var unplayedMatches []model.Match
	for _, match := range matches {
		if !match.IsComplete() {
			// Update the teams in the match if they are not yet set or are incorrect.
			if redAlliance != nil && !(match.Red1 == redAlliance.Lineup[0] && match.Red2 == redAlliance.Lineup[1] &&
				match.Red3 == redAlliance.Lineup[2]) {
				positionRedTeams(&match, redAlliance)
				match.ElimRedAlliance = redAlliance.Id
				if err = database.UpdateMatch(&match); err != nil {
					return nil, err
				}
			}
			if blueAlliance != nil && !(match.Blue1 == blueAlliance.Lineup[0] &&
				match.Blue2 == blueAlliance.Lineup[1] && match.Blue3 == blueAlliance.Lineup[2]) {
				positionBlueTeams(&match, blueAlliance)
				match.ElimBlueAlliance = blueAlliance.Id
				if err = database.UpdateMatch(&match); err != nil {
					return nil, err
				}
			}

			unplayedMatches = append(unplayedMatches, match)
			numIncomplete += 1
			continue
		}

		// Check who won.
		switch match.Status {
		case model.RedWonMatch:
			redWins += 1
		case model.BlueWonMatch:
			blueWins += 1
		case model.TieMatch:
			ties = append(ties, match)
		default:
			return nil, fmt.Errorf("Completed match %d has invalid winner '%s'", match.Id, match.Status)
		}
	}

	// Delete any superfluous matches if the round is won.
	if redWins == 2 || blueWins == 2 {
		for _, match := range unplayedMatches {
			err = database.DeleteMatch(match.Id)
			if err != nil {
				return nil, err
			}
		}

		// Bail out and announce the winner of this round.
		if redWins == 2 {
			return redAlliance, nil
		} else {
			return blueAlliance, nil
		}
	}

	// Create initial set of matches or recreate any superfluous matches that were deleted but now are needed
	// due to a revision in who won.
	if len(matches) == 0 || len(ties) == 0 && numIncomplete == 0 {
		if redAlliance != nil && len(redAlliance.TeamIds) < 3 || blueAlliance != nil && len(blueAlliance.TeamIds) < 3 {
			// Raise an error if the alliance selection process gave us less than 3 teams per alliance.
			return nil, fmt.Errorf("Alliances must consist of at least 3 teams")
		}
		if len(matches) < 1 {
			err = database.CreateMatch(createMatch(roundName, round, group, 1, redAlliance, blueAlliance))
			if err != nil {
				return nil, err
			}
		}
		if len(matches) < 2 {
			err = database.CreateMatch(createMatch(roundName, round, group, 2, redAlliance, blueAlliance))
			if err != nil {
				return nil, err
			}
		}
		if len(matches) < 3 {
			err = database.CreateMatch(createMatch(roundName, round, group, 3, redAlliance, blueAlliance))
			if err != nil {
				return nil, err
			}
		}
	}

	// Duplicate any ties if we have run out of matches.
	if numIncomplete == 0 {
		for index := range ties {
			err = database.CreateMatch(createMatch(roundName, round, group, len(matches)+index+1, redAlliance,
				blueAlliance))
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

// Creates a match at the given point in the elimination bracket and populates the teams.
func createMatch(
	roundName string,
	round int,
	group int,
	instance int,
	redAlliance,
	blueAlliance *model.Alliance,
) *model.Match {
	match := model.Match{
		Type:         "elimination",
		DisplayName:  fmt.Sprintf("%s-%d", roundName, instance),
		ElimRound:    round,
		ElimGroup:    group,
		ElimInstance: instance,
	}
	if redAlliance != nil {
		match.ElimRedAlliance = redAlliance.Id
		positionRedTeams(&match, redAlliance)
	}
	if blueAlliance != nil {
		match.ElimBlueAlliance = blueAlliance.Id
		positionBlueTeams(&match, blueAlliance)
	}
	return &match
}

// Assigns the lineup from the alliance into the red team slots for the match.
func positionRedTeams(match *model.Match, alliance *model.Alliance) {
	match.Red1 = alliance.Lineup[0]
	match.Red2 = alliance.Lineup[1]
	match.Red3 = alliance.Lineup[2]
}

// Assigns the lineup from the alliance into the blue team slots for the match.
func positionBlueTeams(match *model.Match, alliance *model.Alliance) {
	match.Blue1 = alliance.Lineup[0]
	match.Blue2 = alliance.Lineup[1]
	match.Blue3 = alliance.Lineup[2]
}
