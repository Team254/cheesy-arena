// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
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
		RobotsBypassed:      [3]bool{false, false, true},
		AutoFuel:            3,
		ActiveFuel:          50,
		InactiveFuel:        25,
		AutoClimbStatuses:   [3]EndgameStatus{EndgameLevel1, EndgameNone, EndgameNone},
		TeleopClimbStatuses: [3]EndgameStatus{EndgameLevel1, EndgameNone, EndgameLevel3},
		Fouls:               fouls,
		PlayoffDq:           false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotsBypassed:      [3]bool{false, false, false},
		AutoFuel:            5,
		ActiveFuel:          120,
		InactiveFuel:        80,
		AutoClimbStatuses:   [3]EndgameStatus{EndgameNone, EndgameLevel1, EndgameNone},
		TeleopClimbStatuses: [3]EndgameStatus{EndgameLevel3, EndgameLevel2, EndgameLevel1},
		Fouls:               []Foul{},
		PlayoffDq:           false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}
