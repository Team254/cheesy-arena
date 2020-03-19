// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{18, 25, 150},
		{20, 1868, 0},
		{21, 25, 25.2},
	}
	return &Score{
		ExitedInitiationLine: [3]bool{true, true, false},
		AutoCellsBottom:      [2]int{2, 1},
		AutoCellsOuter:       [2]int{6, 0},
		AutoCellsInner:       [2]int{4, 5},
		TeleopPeriodStarted:  true,
		TeleopCellsBottom:    [4]int{0, 11, 2, 0},
		TeleopCellsOuter:     [4]int{0, 5, 0, 0},
		TeleopCellsInner:     [4]int{0, 5, 0, 0},
		RotationControl:      true,
		PositionControl:      false,
		EndgameStatuses:      [3]EndgameStatus{Hang, Hang, Hang},
		RungIsLevel:          false,
		Fouls:                fouls,
		ElimDq:               false,
	}
}

func TestScore2() *Score {
	return &Score{
		ExitedInitiationLine: [3]bool{false, true, false},
		AutoCellsBottom:      [2]int{0, 0},
		AutoCellsOuter:       [2]int{3, 0},
		AutoCellsInner:       [2]int{0, 0},
		TeleopPeriodStarted:  true,
		TeleopCellsBottom:    [4]int{2, 0, 2, 0},
		TeleopCellsOuter:     [4]int{2, 14, 0, 1},
		TeleopCellsInner:     [4]int{2, 6, 20, 0},
		RotationControl:      true,
		PositionControl:      true,
		EndgameStatuses:      [3]EndgameStatus{Park, Park, Hang},
		RungIsLevel:          true,
		Fouls:                []Foul{},
		ElimDq:               false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, RankingFields{20, 625, 90, 554, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, RankingFields{18, 700, 625, 90, 0.1114, 1, 3, 2, 0, 10}}
}
