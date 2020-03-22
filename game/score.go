// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

import "math"

type Score struct {
	ExitedInitiationLine [3]bool
	AutoCellsBottom      [2]int
	AutoCellsOuter       [2]int
	AutoCellsInner       [2]int
	TeleopCellsBottom    [4]int
	TeleopCellsOuter     [4]int
	TeleopCellsInner     [4]int
	ControlPanelStatus
	EndgameStatuses [3]EndgameStatus
	RungIsLevel     bool
	Fouls           []Foul
	ElimDq          bool
}

type ScoreSummary struct {
	InitiationLinePoints     int
	AutoPowerCellPoints      int
	AutoPoints               int
	TeleopPowerCellPoints    int
	PowerCellPoints          int
	ControlPanelPoints       int
	EndgamePoints            int
	FoulPoints               int
	Score                    int
	StagePowerCellsRemaining [3]int
	StagesActivated          [3]bool
	ControlPanelRankingPoint bool
	EndgameRankingPoint      bool
}

// Defines the number of power cells that must be scored within each Stage before it can be activated.
var StageCapacities = map[Stage]int{
	Stage1: 9,
	Stage2: 20,
	Stage3: 20,
}

// Represents a Stage towards whose capacity scored power cells are counted.
type Stage int

const (
	Stage1 Stage = iota
	Stage2
	Stage3
	StageExtra
)

type ControlPanelStatus int

const (
	ControlPanelNone ControlPanelStatus = iota
	ControlPanelRotation
	ControlPanelPosition
)

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgamePark
	EndgameHang
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul, teleopStarted bool) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, exited := range score.ExitedInitiationLine {
		if exited {
			summary.InitiationLinePoints += 5
		}
	}
	for i := 0; i < len(score.AutoCellsBottom); i++ {
		summary.AutoPowerCellPoints += 2 * score.AutoCellsBottom[i]
		summary.AutoPowerCellPoints += 4 * score.AutoCellsOuter[i]
		summary.AutoPowerCellPoints += 6 * score.AutoCellsInner[i]
	}
	summary.AutoPoints = summary.InitiationLinePoints + summary.AutoPowerCellPoints

	// Calculate teleoperated period power cell points.
	for i := 0; i < len(score.TeleopCellsBottom); i++ {
		summary.TeleopPowerCellPoints += score.TeleopCellsBottom[i]
		summary.TeleopPowerCellPoints += 2 * score.TeleopCellsOuter[i]
		summary.TeleopPowerCellPoints += 3 * score.TeleopCellsInner[i]
	}
	summary.PowerCellPoints = summary.AutoPowerCellPoints + summary.TeleopPowerCellPoints

	// Calculate control panel points and stages.
	for i := Stage1; i <= Stage3; i++ {
		summary.StagesActivated[i] = score.stageActivated(i, teleopStarted)
		summary.StagePowerCellsRemaining[i] = int(math.Max(0, float64(StageCapacities[i]-score.stagePowerCells(i))))
	}
	if summary.StagesActivated[Stage2] {
		summary.ControlPanelPoints += 10
	}
	if summary.StagesActivated[Stage3] {
		summary.ControlPanelPoints += 20
		summary.ControlPanelRankingPoint = true
	}

	// Calculate endgame points.
	anyHang := false
	for _, status := range score.EndgameStatuses {
		if status == EndgamePark {
			summary.EndgamePoints += 5
		} else if status == EndgameHang {
			summary.EndgamePoints += 25
			anyHang = true
		}
	}
	if score.RungIsLevel && anyHang {
		summary.EndgamePoints += 15
	}
	summary.EndgameRankingPoint = summary.EndgamePoints >= 65

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	// Check for the opponent fouls that automatically trigger a ranking point.
	for _, foul := range opponentFouls {
		if foul.Rule() != nil && foul.Rule().IsRankingPoint {
			summary.ControlPanelRankingPoint = true
			break
		}
	}

	summary.Score = summary.AutoPoints + summary.TeleopPowerCellPoints + summary.ControlPanelPoints +
		summary.EndgamePoints + summary.FoulPoints

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.ExitedInitiationLine != other.ExitedInitiationLine ||
		score.AutoCellsBottom != other.AutoCellsBottom ||
		score.AutoCellsOuter != other.AutoCellsOuter ||
		score.AutoCellsInner != other.AutoCellsInner ||
		score.TeleopCellsBottom != other.TeleopCellsBottom ||
		score.TeleopCellsOuter != other.TeleopCellsOuter ||
		score.TeleopCellsInner != other.TeleopCellsInner ||
		score.ControlPanelStatus != other.ControlPanelStatus ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.RungIsLevel != other.RungIsLevel ||
		score.ElimDq != other.ElimDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}

// Returns the Stage (1-3) that the score represents, in terms of which Stage scored power cells should count towards.
func (score *Score) CellCountingStage(teleopStarted bool) Stage {
	if score.stageActivated(Stage3, teleopStarted) {
		return StageExtra
	}
	if score.stageActivated(Stage2, teleopStarted) {
		return Stage3
	}
	if score.stageActivated(Stage1, teleopStarted) {
		return Stage2
	}
	return Stage1
}

// Returns true if the preconditions are satisfied for the given Stage to be activated.
func (score *Score) stageAtCapacity(stage Stage, teleopStarted bool) bool {
	if stage > Stage1 && !score.stageActivated(stage-1, teleopStarted) {
		return false
	}
	if capacity, ok := StageCapacities[stage]; ok && score.stagePowerCells(stage) >= capacity {
		return true
	}
	return false
}

// Returns true if the given Stage has been activated.
func (score *Score) stageActivated(stage Stage, teleopStarted bool) bool {
	if score.stageAtCapacity(stage, teleopStarted) {
		switch stage {
		case Stage1:
			return teleopStarted
		case Stage2:
			return score.ControlPanelStatus >= ControlPanelRotation
		case Stage3:
			return score.ControlPanelStatus == ControlPanelPosition
		}
	}
	return false
}

// Returns the total count of scored power cells within the given Stage.
func (score *Score) stagePowerCells(stage Stage) int {
	cells := score.TeleopCellsBottom[stage] + score.TeleopCellsOuter[stage] + score.TeleopCellsInner[stage]
	if stage < Stage3 {
		cells += score.AutoCellsBottom[stage] + score.AutoCellsOuter[stage] + score.AutoCellsInner[stage]
	}
	return cells
}
