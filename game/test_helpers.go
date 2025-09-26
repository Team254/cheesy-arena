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
		RobotsBypassed: [3]bool{false, false, true},
		LeaveStatuses:  [3]bool{true, true, false},
		Reef: Reef{
			AutoBranches:   [3][12]bool{{true}},
			Branches:       [3][12]bool{{true, true}, {true, true, true}},
			AutoTroughNear: 0,
			AutoTroughFar:  1,
			TroughNear:     3,
			TroughFar:      4,
		},
		BargeAlgae:      7,
		ProcessorAlgae:  2,
		EndgameStatuses: [3]EndgameStatus{EndgameParked, EndgameNone, EndgameDeepCage},
		Fouls:           fouls,
		PlayoffDq:       false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotsBypassed: [3]bool{false, false, false},
		LeaveStatuses:  [3]bool{false, true, false},
		Reef: Reef{
			AutoBranches:   [3][12]bool{{}, {}, {true, true, true, true}},
			Branches:       [3][12]bool{{true, true, true}, {true, true, true, true, true}, {true, true, true}},
			AutoTroughNear: 2,
			AutoTroughFar:  1,
			TroughNear:     10,
			TroughFar:      5,
		},
		BargeAlgae:      9,
		ProcessorAlgae:  1,
		EndgameStatuses: [3]EndgameStatus{EndgameDeepCage, EndgameShallowCage, EndgameShallowCage},
		Fouls:           []Foul{},
		PlayoffDq:       false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}
