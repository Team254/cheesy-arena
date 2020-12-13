// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{17, 25, 150},
		{18, 1868, 0},
		{19, 25, 25.2},
	}
	return &Score{
		ExitedInitiationLine: [3]bool{true, true, false},
		AutoCellsBottom:      [2]int{2, 1},
		AutoCellsOuter:       [2]int{6, 0},
		AutoCellsInner:       [2]int{4, 5},
		TeleopCellsBottom:    [4]int{0, 11, 2, 0},
		TeleopCellsOuter:     [4]int{0, 5, 0, 0},
		TeleopCellsInner:     [4]int{0, 5, 0, 0},
		ControlPanelStatus:   ControlPanelRotation,
		EndgameStatuses:      [3]EndgameStatus{EndgameHang, EndgameHang, EndgameHang},
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
		TeleopCellsBottom:    [4]int{2, 0, 2, 0},
		TeleopCellsOuter:     [4]int{2, 14, 0, 1},
		TeleopCellsInner:     [4]int{2, 6, 20, 0},
		ControlPanelStatus:   ControlPanelPosition,
		EndgameStatuses:      [3]EndgameStatus{EndgamePark, EndgamePark, EndgameHang},
		RungIsLevel:          true,
		Fouls:                []Foul{},
		ElimDq:               false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 0.1114, 1, 3, 2, 0, 10}}
}
