// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{Rule{"G18", true, false, ""}, 25, 150},
		{Rule{"G20", true, false, ""}, 1868, 0},
		{Rule{"G22", false, false, ""}, 25, 25.2},
	}
	return &Score{
		RobotStartLevels: [3]int{2, 1, 2},
		SandstormBonuses: [3]bool{true, true, false},
		CargoBaysPreMatch: [8]BayStatus{BayHatch, BayEmpty, BayEmpty, BayCargo, BayHatch, BayCargo, BayHatch,
			BayHatch},
		CargoBays: [8]BayStatus{BayHatchCargo, BayHatch, BayEmpty, BayHatchCargo, BayHatchCargo, BayEmpty,
			BayHatch, BayHatchCargo},
		RocketNearLeftBays:  [3]BayStatus{BayHatchCargo, BayEmpty, BayHatchCargo},
		RocketNearRightBays: [3]BayStatus{BayHatchCargo, BayHatch, BayHatchCargo},
		RocketFarLeftBays:   [3]BayStatus{BayEmpty, BayHatchCargo, BayHatch},
		RocketFarRightBays:  [3]BayStatus{BayEmpty, BayHatchCargo, BayEmpty},
		RobotEndLevels:      [3]int{0, 0, 3},
		Fouls:               fouls,
		ElimDq:              false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotStartLevels: [3]int{1, 2, 1},
		SandstormBonuses: [3]bool{false, true, false},
		CargoBaysPreMatch: [8]BayStatus{BayEmpty, BayEmpty, BayHatch, BayHatch, BayHatch, BayHatch, BayHatch,
			BayHatch},
		CargoBays: [8]BayStatus{BayEmpty, BayEmpty, BayHatchCargo, BayHatchCargo, BayHatchCargo, BayHatch, BayHatch,
			BayHatchCargo},
		RocketNearLeftBays:  [3]BayStatus{BayEmpty, BayEmpty, BayEmpty},
		RocketNearRightBays: [3]BayStatus{BayEmpty, BayEmpty, BayEmpty},
		RocketFarLeftBays:   [3]BayStatus{BayHatchCargo, BayEmpty, BayEmpty},
		RocketFarRightBays:  [3]BayStatus{BayEmpty, BayEmpty, BayHatchCargo},
		RobotEndLevels:      [3]int{1, 3, 2},
		Fouls:               []Foul{},
		ElimDq:              false,
	}
}

func TestScoreValidPreMatch() *Score {
	return &Score{
		RobotStartLevels:  [3]int{1, 2, 3},
		CargoBaysPreMatch: [8]BayStatus{1, 3, 3, 0, 0, 1, 1, 3},
		CargoBays:         [8]BayStatus{1, 3, 3, 0, 0, 1, 1, 3},
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, RankingFields{20, 625, 90, 554, 10, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, RankingFields{18, 700, 625, 90, 554, 0.1114, 1, 3, 2, 0, 10}}
}
