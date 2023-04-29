// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

type gridScoringAction struct {
	Row    Row
	Column int
	isCone bool
	isAuto bool
}

func TestScore1() *Score {
	fouls := []Foul{
		{13, 25, 150},
		{14, 1868, 0},
		{15, 25, 25.2},
	}
	return &Score{
		MobilityStatuses:          [3]bool{true, true, false},
		Grid:                      testGrid1(),
		AutoRobotDockStatuses:     [3]bool{false, true, false},
		AutoChargeStationLevel:    false,
		EndgameStatuses:           [3]EndgameStatus{EndgameParked, EndgameNone, EndgameDocked},
		EndgameChargeStationLevel: true,
		Fouls:                     fouls,
		ElimDq:                    false,
	}
}

func TestScore2() *Score {
	return &Score{
		MobilityStatuses:          [3]bool{false, true, false},
		Grid:                      testGrid2(),
		AutoRobotDockStatuses:     [3]bool{true, true, true},
		AutoChargeStationLevel:    true,
		EndgameStatuses:           [3]EndgameStatus{EndgameDocked, EndgameDocked, EndgameDocked},
		EndgameChargeStationLevel: false,
		Fouls:                     []Foul{},
		ElimDq:                    false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 0.1114, 1, 3, 2, 0, 10}}
}

func testGrid1() Grid {
	// Grid with many pieces but no links and no co-op bonus.
	return buildTestGrid([]gridScoringAction{
		{rowBottom, 0, true, true},
		{rowBottom, 1, true, false},
		{rowBottom, 3, false, true},
		{rowBottom, 6, false, false},
		{rowBottom, 8, true, false},
		{rowMiddle, 1, false, false},
		{rowMiddle, 2, true, true},
		{rowMiddle, 6, true, false},
		{rowMiddle, 7, false, false},
		{rowTop, 0, true, false},
		{rowTop, 2, true, false},
		{rowTop, 4, false, false},
		{rowTop, 7, false, true},
		{rowTop, 8, true, true},
	})
}

func testGrid2() Grid {
	// Full grid with supercharging.
	return buildTestGrid([]gridScoringAction{
		{rowBottom, 0, true, true},
		{rowBottom, 1, true, false},
		{rowBottom, 2, false, false},
		{rowBottom, 3, false, true},
		{rowBottom, 4, false, false},
		{rowBottom, 5, false, true},
		{rowBottom, 6, false, false},
		{rowBottom, 7, true, true},
		{rowBottom, 8, true, false},
		{rowMiddle, 0, true, false},
		{rowMiddle, 1, false, false},
		{rowMiddle, 2, true, true},
		{rowMiddle, 3, true, false},
		{rowMiddle, 4, false, false},
		{rowMiddle, 5, true, false},
		{rowMiddle, 6, true, false},
		{rowMiddle, 7, false, false},
		{rowMiddle, 8, true, false},
		{rowTop, 0, true, false},
		{rowTop, 1, false, false},
		{rowTop, 2, true, false},
		{rowTop, 3, true, false},
		{rowTop, 4, false, false},
		{rowTop, 5, true, false},
		{rowTop, 6, true, false},
		{rowTop, 7, false, true},
		{rowTop, 8, true, true},
		// Supercharging
		{rowBottom, 0, true, false},
		{rowBottom, 0, false, false},
		{rowBottom, 0, false, false},
		{rowBottom, 4, false, false},
		{rowMiddle, 2, true, false},
		{rowMiddle, 2, true, false},
		{rowTop, 7, false, true},
	})
}

func buildTestGrid(gridScoringActions []gridScoringAction) Grid {
	var grid Grid

	// Apply the scoring actions to the grid to get it into the expected state.
	for _, action := range gridScoringActions {
		if action.isCone {
			if action.isAuto {
				grid.Nodes[action.Row][action.Column].AutoCones++
			} else {
				grid.Nodes[action.Row][action.Column].TeleopCones++
			}
		} else {
			if action.isAuto {
				grid.Nodes[action.Row][action.Column].AutoCubes++
			} else {
				grid.Nodes[action.Row][action.Column].TeleopCubes++
			}
		}
	}

	return grid
}
