// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

type gridScoringAction struct {
	Row       Row
	Column    int
	nodeState NodeState
	isAuto    bool
}

func TestScore1() *Score {
	fouls := []Foul{
		{true, 25, 13},
		{false, 1868, 14},
		{true, 25, 15},
	}
	return &Score{
		MobilityStatuses:          [3]bool{true, true, false},
		Grid:                      testGrid1(),
		AutoDockStatuses:          [3]bool{false, true, false},
		AutoChargeStationLevel:    false,
		EndgameStatuses:           [3]EndgameStatus{EndgameParked, EndgameNone, EndgameDocked},
		EndgameChargeStationLevel: true,
		Fouls:                     fouls,
		PlayoffDq:                 false,
	}
}

func TestScore2() *Score {
	return &Score{
		MobilityStatuses:          [3]bool{false, true, false},
		Grid:                      testGrid2(),
		AutoDockStatuses:          [3]bool{true, true, true},
		AutoChargeStationLevel:    true,
		EndgameStatuses:           [3]EndgameStatus{EndgameDocked, EndgameDocked, EndgameDocked},
		EndgameChargeStationLevel: false,
		Fouls:                     []Foul{},
		PlayoffDq:                 false,
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
		{rowBottom, 0, Cone, true},
		{rowBottom, 1, Cone, false},
		{rowBottom, 3, Cube, true},
		{rowBottom, 6, Cube, false},
		{rowBottom, 8, Cone, false},
		{rowMiddle, 1, Cube, false},
		{rowMiddle, 2, Cone, true},
		{rowMiddle, 6, Cone, false},
		{rowMiddle, 7, Cube, false},
		{rowTop, 0, Cone, false},
		{rowTop, 2, Cone, false},
		{rowTop, 4, Cube, false},
		{rowTop, 7, Cube, true},
		{rowTop, 8, Cone, true},
	})
}

func testGrid2() Grid {
	// Full grid with supercharging.
	return buildTestGrid([]gridScoringAction{
		{rowBottom, 0, ConeThenCube, true},
		{rowBottom, 1, Cone, false},
		{rowBottom, 2, Cube, false},
		{rowBottom, 3, Cube, true},
		{rowBottom, 4, CubeThenCone, false},
		{rowBottom, 5, Cube, true},
		{rowBottom, 6, Cube, false},
		{rowBottom, 7, Cone, true},
		{rowBottom, 8, Cone, false},
		{rowMiddle, 0, Cone, false},
		{rowMiddle, 1, Cube, false},
		{rowMiddle, 2, TwoCones, true},
		{rowMiddle, 3, Cone, false},
		{rowMiddle, 4, Cube, false},
		{rowMiddle, 5, Cone, false},
		{rowMiddle, 6, Cone, false},
		{rowMiddle, 7, Cube, false},
		{rowMiddle, 8, Cone, false},
		{rowTop, 0, Cone, false},
		{rowTop, 1, Cube, false},
		{rowTop, 2, Cone, false},
		{rowTop, 3, Cone, false},
		{rowTop, 4, Cube, false},
		{rowTop, 5, Cone, false},
		{rowTop, 6, Cone, false},
		{rowTop, 7, TwoCubes, true},
		{rowTop, 8, Cone, true},
	})
}

func buildTestGrid(gridScoringActions []gridScoringAction) Grid {
	var grid Grid

	// Apply the scoring actions to the grid to get it into the expected state.
	for _, action := range gridScoringActions {
		grid.AutoScoring[action.Row][action.Column] = action.isAuto
		grid.Nodes[action.Row][action.Column] = action.nodeState
	}

	return grid
}
