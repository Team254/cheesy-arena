// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{13, 25, 150},
		{14, 1868, 0},
		{15, 25, 25.2},
	}
	return &Score{
		TaxiStatuses:     [3]bool{true, true, false},
		AutoCargoLower:   [4]int{0, 0, 1, 0},
		AutoCargoUpper:   [4]int{3, 1, 1, 1},
		TeleopCargoLower: [4]int{0, 2, 0, 0},
		TeleopCargoUpper: [4]int{1, 5, 0, 2},
		EndgameStatuses:  [3]EndgameStatus{EndgameLow, EndgameNone, EndgameTraversal},
		Fouls:            fouls,
		ElimDq:           false,
	}
}

func TestScore2() *Score {
	return &Score{
		TaxiStatuses:     [3]bool{false, true, false},
		AutoCargoLower:   [4]int{0, 0, 0, 1},
		AutoCargoUpper:   [4]int{1, 1, 1, 0},
		TeleopCargoLower: [4]int{2, 0, 2, 7},
		TeleopCargoUpper: [4]int{2, 7, 0, 1},
		EndgameStatuses:  [3]EndgameStatus{EndgameNone, EndgameLow, EndgameHigh},
		Fouls:            []Foul{},
		ElimDq:           false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 0.1114, 1, 3, 2, 0, 10}}
}
