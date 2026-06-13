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
		AutoTowerStatuses: [3]TowerStatus{TowerNone, TowerLevel2, TowerNone},
		Hub: Hub{
			WonAuto:     false,
			ShiftCounts: [ShiftCount]int{18, 10, 20, 30, 25, 40, 15},
		},
		EndgameTowerStatuses: [3]TowerStatus{TowerLevel1, TowerLevel2, TowerNone},
		Fouls:                fouls,
		PlayoffDq:            false,
	}
}

func TestScore2() *Score {
	return &Score{
		AutoTowerStatuses: [3]TowerStatus{TowerLevel1, TowerNone, TowerLevel3},
		Hub: Hub{
			WonAuto:     true,
			ShiftCounts: [ShiftCount]int{35, 12, 40, 30, 50, 28, 9},
		},
		EndgameTowerStatuses: [3]TowerStatus{TowerLevel3, TowerLevel2, TowerLevel1},
		Fouls:                []Foul{},
		PlayoffDq:            false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 0.1114, 1, 3, 2, 0, 10}}
}
