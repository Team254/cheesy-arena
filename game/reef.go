// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2025 Reef element.

package game

type Reef struct {
	AutoBranches   [3][12]bool
	Branches       [3][12]bool
	AutoTroughNear int
	AutoTroughFar  int
	TroughNear     int
	TroughFar      int
}

type Level int

const (
	Level1 Level = iota - 1
	Level2
	Level3
	Level4
	LevelCount
)

var autoPoints = map[Level]int{
	Level1: 3,
	Level2: 4,
	Level3: 6,
	Level4: 7,
}

var teleopPoints = map[Level]int{
	Level1: 2,
	Level2: 3,
	Level3: 4,
	Level4: 5,
}

// CountTotalCoralByLevel calculates the total number of Coral scored at a specific level across both auto and teleop.
func (reef *Reef) CountTotalCoralByLevel(level Level) int {
	return reef.CountCoralByLevelAndPeriod(level, true) + reef.CountCoralByLevelAndPeriod(level, false)
}

// AutoCoralCount calculates the total number of Coral scored during the autonomous period across all levels.
func (reef *Reef) AutoCoralCount() int {
	coral := 0
	for level := Level1; level < LevelCount; level++ {
		coral += reef.CountCoralByLevelAndPeriod(level, true)
	}
	return coral
}

// AutoCoralPoints calculates the total points scored during the autonomous period based on the Coral scored at each
// level.
func (reef *Reef) AutoCoralPoints() int {
	points := 0
	for level := Level1; level < LevelCount; level++ {
		points += reef.CountCoralByLevelAndPeriod(level, true) * autoPoints[level]
	}
	return points
}

// TeleopCoralCount calculates the total number of Coral scored during the teleoperated period across all levels.
func (reef *Reef) TeleopCoralCount() int {
	coral := 0
	for level := Level1; level < LevelCount; level++ {
		coral += reef.CountCoralByLevelAndPeriod(level, false)
	}
	return coral
}

// TeleopCoralPoints calculates the total points scored during the teleoperated period based on the Coral scored at each
// level.
func (reef *Reef) TeleopCoralPoints() int {
	points := 0
	for level := Level1; level < LevelCount; level++ {
		points += reef.CountCoralByLevelAndPeriod(level, false) * teleopPoints[level]
	}
	return points
}

// CountCoralByLevelAndPeriod calculates the number of Coral scored at a specific level and period (auto or teleop).
func (reef *Reef) CountCoralByLevelAndPeriod(level Level, isAuto bool) int {
	if level < Level1 || level >= LevelCount {
		return 0
	}

	if level == Level1 {
		troughTotal := reef.TroughNear + reef.TroughFar
		autoTroughTotal := reef.AutoTroughNear + reef.AutoTroughFar

		// Coral must stay scored in teleop to count for auto points, but L1 Coral is not tracked by specific location;
		// it's assumed that lowest-scoring Coral is removed first and highest-scoring Coral re-added first.
		autoCoral := min(autoTroughTotal, troughTotal)
		if isAuto {
			return autoCoral
		}
		return troughTotal - autoCoral
	}

	coral := 0
	for i, branch := range reef.Branches[level] {
		// Coral must stay scored in teleop to count for auto points. Coral initially scored in auto, de-scored in
		// teleop, then re-scored in the same location does count for auto points.
		if branch && isAuto == reef.AutoBranches[level][i] {
			coral++
		}
	}
	return coral
}

// isAutoBonusCoralThresholdMet returns true if the alliance has scored enough Coral in auto to meet that half of the
// bonus RP criteria.
func (reef *Reef) isAutoBonusCoralThresholdMet() bool {
	// Unlike for auto points, de-scoring a Coral in teleop does not invalidate the auto bonus.
	autoCoral := reef.AutoTroughNear + reef.AutoTroughFar
	for _, level := range reef.AutoBranches {
		for _, branch := range level {
			if branch {
				autoCoral++
			}
		}
	}
	return autoCoral >= AutoBonusCoralThreshold
}

// countCoralBonusSatisfiedLevels counts the number of levels that have enough Coral scored on them to satisfy the Coral
// bonus RP.
func (reef *Reef) countCoralBonusSatisfiedLevels() int {
	satisfiedLevels := 0
	for level := Level1; level < LevelCount; level++ {
		autoCoral := reef.CountCoralByLevelAndPeriod(level, true)
		teleopCoral := reef.CountCoralByLevelAndPeriod(level, false)
		if autoCoral+teleopCoral >= CoralBonusPerLevelThreshold {
			satisfiedLevels++
		}
	}
	return satisfiedLevels
}
