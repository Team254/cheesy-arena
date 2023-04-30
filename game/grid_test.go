// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type gridTestCase struct {
	name                                    string
	gridScoringActions                      []gridScoringAction
	expectedAutoGamePiecePoints             int
	expectedTeleopGamePiecePoints           int
	expectedSuperchargedPoints              int
	expectedLinks                           []Link
	expectedIsCoopertitionThresholdAchieved bool
	expectedIsFull                          bool
}

var gridTestCases = []gridTestCase{
	{
		name: "No scoring actions",
	},
	{
		name: "Same node scored multiple times in auto",
		gridScoringActions: []gridScoringAction{
			{rowTop, 7, TwoCubes, true},
		},
		expectedAutoGamePiecePoints:   6,
		expectedTeleopGamePiecePoints: 0,
	},
	{
		name: "Same node scored multiple times in teleop",
		gridScoringActions: []gridScoringAction{
			{rowTop, 7, TwoCubes, false},
		},
		expectedAutoGamePiecePoints:   0,
		expectedTeleopGamePiecePoints: 5,
	},
	{
		name: "Grid with many pieces but no links and no co-op bonus",
		gridScoringActions: []gridScoringAction{
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
		},
		expectedAutoGamePiecePoints:   22,
		expectedTeleopGamePiecePoints: 30,
	},
	{
		name: "Non-aligned links",
		gridScoringActions: []gridScoringAction{
			{rowMiddle, 1, Cube, false},
			{rowMiddle, 2, Cone, false},
			{rowMiddle, 3, Cone, false},
			{rowMiddle, 4, Cube, false},
			{rowMiddle, 5, Cone, false},
			{rowMiddle, 6, Cone, false},
		},
		expectedAutoGamePiecePoints:   0,
		expectedTeleopGamePiecePoints: 18,
		expectedLinks: []Link{
			{rowMiddle, 1},
			{rowMiddle, 4},
		},
		expectedIsCoopertitionThresholdAchieved: true,
	},
	{
		name: "Coopertition threshold achieved across multiple rows",
		gridScoringActions: []gridScoringAction{
			{rowBottom, 3, Cone, true},
			{rowMiddle, 4, Cube, false},
			{rowTop, 5, Cone, true},
		},
		expectedAutoGamePiecePoints:             9,
		expectedTeleopGamePiecePoints:           3,
		expectedIsCoopertitionThresholdAchieved: true,
	},
	{
		name: "Coopertition threshold not achieved due to wrong game piece",
		gridScoringActions: []gridScoringAction{
			{rowBottom, 3, Cone, true},
			{rowMiddle, 4, Cube, false},
			{rowTop, 5, Cube, true},
		},
		expectedAutoGamePiecePoints:             3,
		expectedTeleopGamePiecePoints:           3,
		expectedIsCoopertitionThresholdAchieved: false,
	},
	{
		name: "Full grid without supercharging",
		gridScoringActions: []gridScoringAction{
			{rowBottom, 0, Cone, true},
			{rowBottom, 1, Cone, false},
			{rowBottom, 2, Cube, false},
			{rowBottom, 3, Cube, true},
			{rowBottom, 4, Cube, false},
			{rowBottom, 5, Cube, true},
			{rowBottom, 6, Cube, false},
			{rowBottom, 7, Cone, true},
			{rowBottom, 8, Cone, false},
			{rowMiddle, 0, Cone, false},
			{rowMiddle, 1, Cube, false},
			{rowMiddle, 2, Cone, true},
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
			{rowTop, 7, Cube, true},
			{rowTop, 8, Cone, true},
		},
		expectedAutoGamePiecePoints:   28,
		expectedTeleopGamePiecePoints: 69,
		expectedLinks: []Link{
			{rowBottom, 0},
			{rowBottom, 3},
			{rowBottom, 6},
			{rowMiddle, 0},
			{rowMiddle, 3},
			{rowMiddle, 6},
			{rowTop, 0},
			{rowTop, 3},
			{rowTop, 6},
		},
		expectedIsCoopertitionThresholdAchieved: true,
		expectedIsFull:                          true,
	},
	{
		name: "Full grid with supercharging",
		gridScoringActions: []gridScoringAction{
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
		},
		expectedAutoGamePiecePoints:   28,
		expectedTeleopGamePiecePoints: 69,
		expectedSuperchargedPoints:    12,
		expectedLinks: []Link{
			{rowBottom, 0},
			{rowBottom, 3},
			{rowBottom, 6},
			{rowMiddle, 0},
			{rowMiddle, 3},
			{rowMiddle, 6},
			{rowTop, 0},
			{rowTop, 3},
			{rowTop, 6},
		},
		expectedIsCoopertitionThresholdAchieved: true,
		expectedIsFull:                          true,
	},
	{
		name: "Invalid scoring actions are ignored",
		gridScoringActions: []gridScoringAction{
			{rowMiddle, 0, Cube, false},
			{rowMiddle, 1, Cone, true},
			{rowMiddle, 2, TwoCubes, true},
			{rowMiddle, 3, Cube, true},
			{rowMiddle, 4, Cone, true},
			{rowMiddle, 5, CubeThenCone, false},
			{rowMiddle, 6, Cube, false},
			{rowMiddle, 7, Cone, true},
			{rowMiddle, 8, Cube, false},
			{rowTop, 0, Cube, true},
			{rowTop, 1, TwoCones, false},
			{rowTop, 2, Cube, true},
			{rowTop, 3, Cube, false},
			{rowTop, 4, Cone, true},
			{rowTop, 5, Cube, true},
			{rowTop, 6, Cube, false},
			{rowTop, 7, ConeThenCube, true},
			{rowTop, 8, Cube, true},
		},
		expectedAutoGamePiecePoints:   6,
		expectedTeleopGamePiecePoints: 3,
	},
}

func TestGrid(t *testing.T) {
	for _, testCase := range gridTestCases {
		grid := buildTestGrid(testCase.gridScoringActions)

		assert.Equal(t, testCase.expectedAutoGamePiecePoints, grid.AutoGamePiecePoints(), testCase.name)
		assert.Equal(t, testCase.expectedTeleopGamePiecePoints, grid.TeleopGamePiecePoints(), testCase.name)
		assert.Equal(t, testCase.expectedSuperchargedPoints, grid.SuperchargedPoints(), testCase.name)
		assert.Equal(t, 5*len(testCase.expectedLinks), grid.LinkPoints(), testCase.name)
		assert.Equal(t, testCase.expectedLinks, grid.Links(), testCase.name)
		assert.Equal(
			t, testCase.expectedIsCoopertitionThresholdAchieved, grid.IsCoopertitionThresholdAchieved(), testCase.name,
		)
		assert.Equal(t, testCase.expectedIsFull, grid.IsFull(), testCase.name)
	}
}
