// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func TestAddScoreSummary(t *testing.T) {
	rand.Seed(0)
	redScore := TestScore1()
	blueScore := TestScore2()
	redSummary := redScore.Summarize(blueScore.Fouls)
	blueSummary := blueScore.Summarize(redScore.Fouls)
	rankingFields := RankingFields{}

	// Add a loss.
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	assert.Equal(t, RankingFields{1, 30, 20, 12, 9, 0.9451961492941164, 0, 1, 0, 0, 1}, rankingFields)

	// Add a win.
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	assert.Equal(t, RankingFields{4, 48, 24, 33, 15, 0.24496508529377975, 1, 1, 0, 0, 2}, rankingFields)

	// Add a tie.
	rankingFields.AddScoreSummary(redSummary, redSummary, false)
	assert.Equal(t, RankingFields{6, 78, 44, 45, 24, 0.6559562651954052, 1, 1, 1, 0, 3}, rankingFields)

	// Add a disqualification.
	rankingFields.AddScoreSummary(blueSummary, redSummary, true)
	assert.Equal(t, RankingFields{6, 78, 44, 45, 24, 0.6559562651954052, 1, 1, 1, 1, 4}, rankingFields)
}

func TestSortRankings(t *testing.T) {
	// Check tiebreakers.
	rankings := make(Rankings, 12)
	rankings[0] = &Ranking{1, 0, RankingFields{50, 50, 50, 50, 50, 0.49, 3, 2, 1, 0, 10}}
	rankings[1] = &Ranking{2, 0, RankingFields{50, 50, 50, 50, 50, 0.51, 3, 2, 1, 0, 10}}
	rankings[2] = &Ranking{3, 0, RankingFields{50, 50, 50, 50, 49, 0.50, 3, 2, 1, 0, 10}}
	rankings[3] = &Ranking{4, 0, RankingFields{50, 50, 50, 50, 51, 0.50, 3, 2, 1, 0, 10}}
	rankings[4] = &Ranking{5, 0, RankingFields{50, 50, 50, 49, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[5] = &Ranking{6, 0, RankingFields{50, 50, 50, 51, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[6] = &Ranking{7, 0, RankingFields{50, 50, 49, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[7] = &Ranking{8, 0, RankingFields{50, 50, 51, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[8] = &Ranking{9, 0, RankingFields{50, 49, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[9] = &Ranking{10, 0, RankingFields{50, 51, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[10] = &Ranking{11, 0, RankingFields{49, 50, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[11] = &Ranking{12, 0, RankingFields{51, 50, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 12, rankings[0].TeamId)
	assert.Equal(t, 10, rankings[1].TeamId)
	assert.Equal(t, 8, rankings[2].TeamId)
	assert.Equal(t, 6, rankings[3].TeamId)
	assert.Equal(t, 4, rankings[4].TeamId)
	assert.Equal(t, 2, rankings[5].TeamId)
	assert.Equal(t, 1, rankings[6].TeamId)
	assert.Equal(t, 3, rankings[7].TeamId)
	assert.Equal(t, 5, rankings[8].TeamId)
	assert.Equal(t, 7, rankings[9].TeamId)
	assert.Equal(t, 9, rankings[10].TeamId)
	assert.Equal(t, 11, rankings[11].TeamId)

	// Check with unequal number of matches played.
	rankings = make(Rankings, 3)
	rankings[0] = &Ranking{1, 0, RankingFields{10, 25, 25, 25, 25, 0.49, 3, 2, 1, 0, 5}}
	rankings[1] = &Ranking{2, 0, RankingFields{19, 50, 50, 50, 50, 0.51, 3, 2, 1, 0, 9}}
	rankings[2] = &Ranking{3, 0, RankingFields{20, 50, 50, 50, 50, 0.51, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 3, rankings[1].TeamId)
	assert.Equal(t, 1, rankings[2].TeamId)
}
