// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	assertMatchStatus := func(
		expectedStatus MatchStatus,
		expectedTiebreaker string,
		redScoreSummary *ScoreSummary,
		blueScoreSummary *ScoreSummary,
		applyPlayoffTiebreakers bool,
	) {
		status, tiebreaker := DetermineMatchStatus(redScoreSummary, blueScoreSummary, applyPlayoffTiebreakers)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, expectedTiebreaker, tiebreaker)
	}

	redScoreSummary := &ScoreSummary{Score: 10}
	blueScoreSummary := &ScoreSummary{Score: 10}
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.Score = 11
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.Score = 12
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.Score = 12
	redScoreSummary.NumOpponentMajorFouls = 11
	redScoreSummary.AutoFuelPoints = 11
	redScoreSummary.AutoTowerPoints = 5
	redScoreSummary.TeleopTowerPoints = 6
	blueScoreSummary.NumOpponentMajorFouls = 10
	blueScoreSummary.AutoFuelPoints = 10
	blueScoreSummary.AutoTowerPoints = 4
	blueScoreSummary.TeleopTowerPoints = 6
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: MAJOR FOULS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.NumOpponentMajorFouls = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "TIEBREAK: MAJOR FOULS", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.NumOpponentMajorFouls = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: AUTO FUEL", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.AutoFuelPoints = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "TIEBREAK: AUTO FUEL", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.AutoFuelPoints = 12
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(RedWonMatch, "TIEBREAK: TOWER POINTS", redScoreSummary, blueScoreSummary, true)

	blueScoreSummary.TeleopTowerPoints = 8
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(BlueWonMatch, "TIEBREAK: TOWER POINTS", redScoreSummary, blueScoreSummary, true)

	redScoreSummary.TeleopTowerPoints = 7
	assertMatchStatus(TieMatch, "", redScoreSummary, blueScoreSummary, false)
	assertMatchStatus(TieMatch, "TRUE TIE", redScoreSummary, blueScoreSummary, true)
}
