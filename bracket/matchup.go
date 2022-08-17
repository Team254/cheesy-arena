// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Models and logic encapsulating a group of one or more matches between the same two alliances at a given point in a
// playoff tournament.

package bracket

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"strconv"
	"strings"
)

// Conveys how a given alliance should be populated -- either directly from alliance selection or based on the results
// of a prior matchup.
type allianceSource struct {
	allianceId int
	matchupKey matchupKey
	useWinner  bool
}

// Key for uniquely identifying a matchup. Round IDs are arbitrary and in descending order with "1" always representing
// the playoff finals. Group IDs are 1-indexed within a round and increasing in order of play.
type matchupKey struct {
	round int
	group int
}

// Conveys the complete generic information about a matchup required to construct it. In aggregate, the full list of
// match templates describing a bracket format can be used to construct an empty playoff bracket for a given number of
// alliances.
type matchupTemplate struct {
	matchupKey
	displayNameFormat  string
	numWinsToAdvance   int
	redAllianceSource  allianceSource
	blueAllianceSource allianceSource
}

// Encapsulates the format and state of a group of one or more matches between the same two alliances at a given point
// in a playoff tournament.
type Matchup struct {
	matchupTemplate
	RedAllianceSourceMatchup  *Matchup
	BlueAllianceSourceMatchup *Matchup
	RedAllianceId             int
	BlueAllianceId            int
	RedAllianceWins           int
	BlueAllianceWins          int
}

// Convenience method to quickly create an alliance source that points to the winner of a different matchup.
func newWinnerAllianceSource(round, group int) allianceSource {
	return allianceSource{matchupKey: newMatchupKey(round, group), useWinner: true}
}

// Convenience method to quickly create an alliance source that points to the loser of a different matchup.
func newLoserAllianceSource(round, group int) allianceSource {
	return allianceSource{matchupKey: newMatchupKey(round, group), useWinner: false}
}

// Convenience method to quickly create a matchup key.
func newMatchupKey(round, group int) matchupKey {
	return matchupKey{round: round, group: group}
}

// Returns the display name for a specific match within a matchup.
func (matchupTemplate *matchupTemplate) displayName(instance int) string {
	displayName := matchupTemplate.displayNameFormat
	displayName = strings.Replace(displayName, "${group}", strconv.Itoa(matchupTemplate.group), -1)
	if strings.Contains(displayName, "${instance}") {
		displayName = strings.Replace(displayName, "${instance}", strconv.Itoa(instance), -1)
	} else if instance > 1 {
		// Special case to handle matchups that only have more than one instance under exceptional circumstances (like
		// ties in double-elimination unresolved by tiebreakers).
		displayName += fmt.Sprintf("-%d", instance)
	}
	return displayName
}

// Returns the winning alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) winner() int {
	if matchup.RedAllianceWins >= matchup.numWinsToAdvance {
		return matchup.RedAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.numWinsToAdvance {
		return matchup.BlueAllianceId
	}
	return 0
}

// Returns the losing alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) loser() int {
	if matchup.RedAllianceWins >= matchup.numWinsToAdvance {
		return matchup.BlueAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.numWinsToAdvance {
		return matchup.RedAllianceId
	}
	return 0
}

// Returns true if the matchup has been won, and false if it is still to be determined.
func (matchup *Matchup) isComplete() bool {
	return matchup.winner() > 0
}

// Recursively traverses the matchup graph to update the state of this matchup and all of its children based on match
// results, counting wins and creating or deleting matches as required.
func (matchup *Matchup) update(database *model.Database) error {
	// Update child matchups first. Only recurse down winner links to avoid visiting a node twice.
	if matchup.RedAllianceSourceMatchup != nil && matchup.redAllianceSource.useWinner {
		if err := matchup.RedAllianceSourceMatchup.update(database); err != nil {
			return err
		}
	}
	if matchup.BlueAllianceSourceMatchup != nil && matchup.blueAllianceSource.useWinner {
		if err := matchup.BlueAllianceSourceMatchup.update(database); err != nil {
			return err
		}
	}

	// Populate the alliance IDs from the lower matchups (or with a zero value if they are not yet complete).
	if matchup.RedAllianceSourceMatchup != nil {
		if matchup.redAllianceSource.useWinner {
			matchup.RedAllianceId = matchup.RedAllianceSourceMatchup.winner()
		} else {
			matchup.RedAllianceId = matchup.RedAllianceSourceMatchup.loser()
		}
	}
	if matchup.BlueAllianceSourceMatchup != nil {
		if matchup.blueAllianceSource.useWinner {
			matchup.BlueAllianceId = matchup.BlueAllianceSourceMatchup.winner()
		} else {
			matchup.BlueAllianceId = matchup.BlueAllianceSourceMatchup.loser()
		}
	}

	matches, err := database.GetMatchesByElimRoundGroup(matchup.round, matchup.group)
	if err != nil {
		return err
	}

	// Bail if we do not yet know both alliances.
	if matchup.RedAllianceId == 0 || matchup.BlueAllianceId == 0 {
		// Ensure the current state is reset; it may have previously been populated if a match result was edited.
		matchup.RedAllianceWins = 0
		matchup.BlueAllianceWins = 0

		// Delete any previously created matches.
		for _, match := range matches {
			if err = database.DeleteMatch(match.Id); err != nil {
				return err
			}
		}

		return nil
	}

	// Create, update, and/or delete unplayed matches as required.
	redAlliance, err := database.GetAllianceById(matchup.RedAllianceId)
	if err != nil {
		return err
	}
	blueAlliance, err := database.GetAllianceById(matchup.BlueAllianceId)
	if err != nil {
		return err
	}
	matchup.RedAllianceWins = 0
	matchup.BlueAllianceWins = 0
	var unplayedMatches []model.Match
	for _, match := range matches {
		if !match.IsComplete() {
			// Update the teams in the match if they are not yet set or are incorrect.
			changed := false
			if match.Red1 != redAlliance.Lineup[0] || match.Red2 != redAlliance.Lineup[1] ||
				match.Red3 != redAlliance.Lineup[2] {
				positionRedTeams(&match, redAlliance)
				match.ElimRedAlliance = redAlliance.Id
				changed = true
				if err = database.UpdateMatch(&match); err != nil {
					return err
				}
			}
			if match.Blue1 != blueAlliance.Lineup[0] || match.Blue2 != blueAlliance.Lineup[1] ||
				match.Blue3 != blueAlliance.Lineup[2] {
				positionBlueTeams(&match, blueAlliance)
				match.ElimBlueAlliance = blueAlliance.Id
				changed = true
			}
			if changed {
				if err = database.UpdateMatch(&match); err != nil {
					return err
				}
			}

			unplayedMatches = append(unplayedMatches, match)
			continue
		}

		// Check who won.
		if match.Status == model.RedWonMatch {
			matchup.RedAllianceWins++
		} else if match.Status == model.BlueWonMatch {
			matchup.BlueAllianceWins++
		}
	}

	maxWins := matchup.RedAllianceWins
	if matchup.BlueAllianceWins > maxWins {
		maxWins = matchup.BlueAllianceWins
	}
	numUnplayedMatchesNeeded := matchup.numWinsToAdvance - maxWins
	if len(unplayedMatches) > numUnplayedMatchesNeeded {
		// Delete any superfluous matches off the end of the list.
		for i := 0; i < len(unplayedMatches)-numUnplayedMatchesNeeded; i++ {
			if err = database.DeleteMatch(unplayedMatches[len(unplayedMatches)-i-1].Id); err != nil {
				return err
			}
		}
	} else if len(unplayedMatches) < numUnplayedMatchesNeeded {
		// Create initial set of matches or any additional required matches due to tie matches or ties in the round.
		for i := 0; i < numUnplayedMatchesNeeded-len(unplayedMatches); i++ {
			instance := len(matches) + i + 1
			match := model.Match{
				Type:             "elimination",
				DisplayName:      matchup.displayName(instance),
				ElimRound:        matchup.round,
				ElimGroup:        matchup.group,
				ElimInstance:     instance,
				ElimRedAlliance:  redAlliance.Id,
				ElimBlueAlliance: blueAlliance.Id,
			}
			positionRedTeams(&match, redAlliance)
			positionBlueTeams(&match, blueAlliance)
			if err = database.CreateMatch(&match); err != nil {
				return err
			}
		}
	}

	return nil
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
