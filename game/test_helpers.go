// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{{Rule{"G22", false, ""}, 25, 25.2}, {Rule{"G18", true, ""}, 25, 150},
		{Rule{"G20", true, ""}, 1868, 0}}
	return &Score{1, 1.5, 4.5, true, 25.4, 0, 21.6, 0, 0, 0, 3, true, 0, 0, 2, 0, fouls, false}
}

func TestScore2() *Score {
	return &Score{3, 4, 6, true, 33, 10, 20, 10, 3, 3, 0, false, 3, 3, 1, 1, []Foul{}, false}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, RankingFields{20, 625, 90, 554, 10, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, RankingFields{18, 700, 625, 90, 554, 0.1114, 1, 3, 2, 0, 10}}
}
