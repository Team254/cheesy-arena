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

type gridScoringAction struct {
	Row    Row
	Column int
	isCone bool
	isAuto bool
}

var gridTestCases = []gridTestCase{
	{
		name: "No scoring actions",
	},
	{
		name: "Same node scored multiple times in auto",
		gridScoringActions: []gridScoringAction{
			{rowTop, 7, false, true},
			{rowTop, 7, false, true},
			{rowTop, 7, false, true},
		},
		expectedAutoGamePiecePoints:   6,
		expectedTeleopGamePiecePoints: 0,
	},
	{
		name: "Same node scored in auto and teleop",
		gridScoringActions: []gridScoringAction{
			{rowTop, 7, false, true},
			{rowTop, 7, false, false},
			{rowTop, 7, false, false},
		},
		expectedAutoGamePiecePoints:   6,
		expectedTeleopGamePiecePoints: 0,
	},
	{
		name: "Same node scored multiple times in teleop",
		gridScoringActions: []gridScoringAction{
			{rowTop, 7, false, false},
			{rowTop, 7, false, false},
			{rowTop, 7, false, false},
		},
		expectedAutoGamePiecePoints:   0,
		expectedTeleopGamePiecePoints: 5,
	},
	{
		name: "Grid with many pieces but no links and no co-op bonus",
		gridScoringActions: []gridScoringAction{
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
		},
		expectedAutoGamePiecePoints:   22,
		expectedTeleopGamePiecePoints: 30,
	},
	{
		name: "Non-aligned links",
		gridScoringActions: []gridScoringAction{
			{rowMiddle, 1, false, false},
			{rowMiddle, 2, true, false},
			{rowMiddle, 3, true, false},
			{rowMiddle, 4, false, false},
			{rowMiddle, 5, true, false},
			{rowMiddle, 6, true, false},
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
			{rowBottom, 3, true, true},
			{rowMiddle, 4, false, false},
			{rowTop, 5, true, true},
		},
		expectedAutoGamePiecePoints:             9,
		expectedTeleopGamePiecePoints:           3,
		expectedIsCoopertitionThresholdAchieved: true,
	},
	{
		name: "Full grid without supercharging",
		gridScoringActions: []gridScoringAction{
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
			{rowMiddle, 0, false, false},
			{rowMiddle, 1, true, true},
			{rowMiddle, 2, false, true},
			{rowMiddle, 3, false, true},
			{rowMiddle, 4, true, true},
			{rowMiddle, 5, false, true},
			{rowMiddle, 6, false, false},
			{rowMiddle, 7, true, true},
			{rowMiddle, 8, false, false},
			{rowTop, 0, false, true},
			{rowTop, 1, true, false},
			{rowTop, 2, false, true},
			{rowTop, 3, false, false},
			{rowTop, 4, true, true},
			{rowTop, 5, false, true},
			{rowTop, 6, false, false},
			{rowTop, 7, true, true},
			{rowTop, 8, false, true},
		},
	},
}

func TestGrid(t *testing.T) {
	for _, testCase := range gridTestCases {
		var grid Grid

		// Apply the scoring actions to the grid to get it into the expected state.
		for _, action := range testCase.gridScoringActions {
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
