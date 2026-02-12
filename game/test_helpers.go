// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{1, true, 25, 16},
		{2, false, 1868, 13},
		{3, false, 1868, 13},
		{4, true, 25, 15},
		{5, true, 25, 15},
		{6, true, 25, 15},
		{7, true, 25, 15},
	}
	return &Score{
		RobotsBypassed: [3]bool{false, false, true},
		// 模擬：前兩台機器人 Auto 達成 Tower Level 1
		AutoTowerLevel1: [3]bool{true, true, false},

		// 模擬：Auto 投 5 顆，Teleop 投 20 顆
		AutoFuelCount:   5,
		TeleopFuelCount: 20,

		// 模擬：Endgame 狀態 (一台 Level 3，一台 None，一台 Level 2)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameNone, EndgameLevel2},
		Fouls:           fouls,
		PlayoffDq:       false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotsBypassed: [3]bool{false, false, false},
		// 模擬：只有第二台機器人 Auto 達成 Tower Level 1
		AutoTowerLevel1: [3]bool{false, true, false},

		// 模擬：Auto 投 10 顆，Teleop 投 40 顆
		AutoFuelCount:   10,
		TeleopFuelCount: 40,

		// 模擬：Endgame 狀態 (Level 3, Level 2, Level 2)
		EndgameStatuses: [3]EndgameStatus{EndgameLevel3, EndgameLevel2, EndgameLevel2},
		Fouls:           []Foul{},
		PlayoffDq:       false,
	}
}

// 對應新的 RankingFields 結構：RankingPoints, MatchPoints, AutoPoints, TowerPoints, Random, W, L, T, DQ, Played
func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 100, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 120, 0.1114, 1, 3, 2, 0, 10}}
}
