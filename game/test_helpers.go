// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{1, true, 25, 16},
		{2, false, 1868, 13},
		{3, false, 1868, 13},
		{4, true, 25, 15},
		{5, true, 25, 15},
		{6, true, 25, 15},
		{7, true, 25, 15},
	}
	return &Score{
		RobotsBypassed: [3]bool{false, false, true},
		// Simulation: The first two robots Auto reach Tower Level 1
		AutoTowerLevel1: [3]bool{true, true, false},

		// Simulation: Auto scores 5 fuel, Teleop scores 20 fuel
		AutoFuelCount:   5,
		TeleopFuelCount: 20,

		// Simulation: Endgame statuses (one Level 3, one None, one Level 2)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameNone, EndgameLevel2},
		Fouls:           fouls,
		PlayoffDq:       false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotsBypassed: [3]bool{false, false, false},
		// Simulation: Only the second robot Auto reaches Tower Level 1
		AutoTowerLevel1: [3]bool{false, true, false},

		// Simulation: Auto scores 10 fuel, Teleop scores 40 fuel
		AutoFuelCount:   10,
		TeleopFuelCount: 40,

		// Simulation: Endgame statuses (Level 3, Level 2, Level 2)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameLevel2, EndgameLevel2},
		Fouls:           []Foul{},
		PlayoffDq:       false,
	}
}

// Corresponds to the new RankingFields structure: RankingPoints, MatchPoints, AutoPoints, TowerPoints, Random, W, L, T, DQ, Played
func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 100, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 120, 0.1114, 1, 3, 2, 0, 10}}
}
