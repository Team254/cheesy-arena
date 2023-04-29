// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2023 Grid element.

package game

type Grid struct {
	Nodes [3][9]Node
}

type Node struct {
	AutoCones   int
	AutoCubes   int
	TeleopCones int
	TeleopCubes int
}

type Link struct {
	Row         Row
	StartColumn int
}

type Row int

const (
	rowBottom Row = iota
	rowMiddle
	rowTop
	rowCount
)

var autoPoints = map[Row]int{
	rowBottom: 3,
	rowMiddle: 4,
	rowTop:    6,
}

var teleopPoints = map[Row]int{
	rowBottom: 2,
	rowMiddle: 3,
	rowTop:    5,
}

func (grid *Grid) AutoGamePiecePoints() int {
	points := 0
	for row := rowBottom; row < rowCount; row++ {
		for column := 0; column < 9; column++ {
			autoPieces, _ := grid.numScoredAutoTeleopGamePieces(row, column)
			if autoPieces > 0 {
				points += autoPoints[row]
			}
		}
	}
	return points
}

func (grid *Grid) TeleopGamePiecePoints() int {
	points := 0
	for row := rowBottom; row < rowCount; row++ {
		for column := 0; column < 9; column++ {
			autoPieces, teleopPieces := grid.numScoredAutoTeleopGamePieces(row, column)
			if autoPieces == 0 && teleopPieces > 0 {
				points += teleopPoints[row]
			}
		}
	}
	return points
}

func (grid *Grid) SuperchargedPoints() int {
	return 3 * grid.NumSuperchargedNodes()
}

func (grid *Grid) NumSuperchargedNodes() int {
	if !grid.IsFull() {
		return 0
	}

	numSuperchargedNodes := 0
	for row := rowBottom; row < rowCount; row++ {
		for column := 0; column < 9; column++ {
			if grid.numScoredGamePieces(row, column) > 1 {
				numSuperchargedNodes++
			}
		}
	}
	return numSuperchargedNodes
}

func (grid *Grid) LinkPoints() int {
	return 5 * len(grid.Links())
}

func (grid *Grid) Links() []Link {
	var links []Link
	for row := rowBottom; row < rowCount; row++ {
		startColumn := 0
		for startColumn < 7 {
			isValidLink := true
			for column := startColumn; column < startColumn+3; column++ {
				if grid.numScoredGamePieces(row, column) == 0 {
					isValidLink = false
					break
				}
			}

			if isValidLink {
				link := Link{Row: row, StartColumn: startColumn}
				links = append(links, link)
				startColumn += 3
			} else {
				startColumn++
			}
		}
	}
	return links
}

// Returns true if this grid contains enough scored nodes to activate the coopertition bonus (both alliances' grids must
// meet this condition for the bonus to be awarded).
func (grid *Grid) IsCoopertitionThresholdAchieved() bool {
	pieces := 0
	for row := rowBottom; row < rowCount; row++ {
		for column := 3; column < 6; column++ {
			pieces += grid.numScoredGamePieces(row, column)
		}
	}

	return pieces >= 3
}

func (grid *Grid) IsFull() bool {
	for row := rowBottom; row < rowCount; row++ {
		for column := 0; column < 9; column++ {
			if grid.numScoredGamePieces(row, column) == 0 {
				return false
			}
		}
	}
	return true
}

// Returns the separate counts of scored auto and teleop game pieces in the given node, limiting them to valid values.
func (grid *Grid) numScoredAutoTeleopGamePieces(row Row, column int) (int, int) {
	if row < rowBottom || row > rowTop || column < 0 || column > 8 {
		// This is not a valid node.
		return 0, 0
	}

	node := grid.Nodes[row][column]
	autoPieces := 0
	teleopPieces := 0
	if row == rowBottom {
		autoPieces = node.AutoCones + node.AutoCubes
		teleopPieces = node.TeleopCones + node.TeleopCubes
	} else {
		// Don't count game pieces of the wrong type for this node.
		if column == 1 || column == 4 || column == 7 {
			autoPieces = node.AutoCubes
			teleopPieces = node.TeleopCubes
		} else {
			autoPieces = node.AutoCones
			teleopPieces = node.TeleopCones
		}
	}

	// Don't report more than two pieces scored in a node.
	if autoPieces > 2 {
		autoPieces = 2
	}
	if autoPieces+teleopPieces > 2 {
		teleopPieces = 2 - autoPieces
	}

	return autoPieces, teleopPieces
}

// Returns the total number of game pieces in the given node, limiting it to valid values.
func (grid *Grid) numScoredGamePieces(row Row, column int) int {
	autoPieces, teleopPieces := grid.numScoredAutoTeleopGamePieces(row, column)
	return autoPieces + teleopPieces
}
