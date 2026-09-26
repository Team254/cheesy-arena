// Copyright 2026 Team 254. All Rights Reserved.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
)

// Rebuilds carried alliance yellow cards from the latest results of completed playoff matches.
func CalculateAllianceCards(database *model.Database) error {
	alliances, err := database.GetAllAlliances()
	if err != nil {
		return err
	}
	yellowCards := make(map[int]bool)
	matches, err := database.GetMatchesByType(model.Playoff, false)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if !match.IsComplete() {
			continue
		}
		result, err := database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("found no match result for match %d", match.Id)
		}
		if result.PlayoffRedAllianceCard == "yellow" || result.PlayoffRedAllianceCard == "red" {
			yellowCards[match.PlayoffRedAlliance] = true
		}
		if result.PlayoffBlueAllianceCard == "yellow" || result.PlayoffBlueAllianceCard == "red" {
			yellowCards[match.PlayoffBlueAlliance] = true
		}
	}
	for _, alliance := range alliances {
		alliance.YellowCard = yellowCards[alliance.Id]
		if err := database.UpdateAlliance(&alliance); err != nil {
			return err
		}
	}
	return nil
}
