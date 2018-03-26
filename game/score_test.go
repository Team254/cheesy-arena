// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore.Fouls)
	assert.Equal(t, 5, redSummary.AutoRunPoints)
	assert.Equal(t, 17, redSummary.AutoPoints)
	assert.Equal(t, 59, redSummary.OwnershipPoints)
	assert.Equal(t, 15, redSummary.VaultPoints)
	assert.Equal(t, 90, redSummary.ParkClimbPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 169, redSummary.Score)
	assert.Equal(t, false, redSummary.AutoQuest)
	assert.Equal(t, true, redSummary.FaceTheBoss)

	blueSummary := blueScore.Summarize(redScore.Fouls)
	assert.Equal(t, 15, blueSummary.AutoRunPoints)
	assert.Equal(t, 35, blueSummary.AutoPoints)
	assert.Equal(t, 93, blueSummary.OwnershipPoints)
	assert.Equal(t, 30, blueSummary.VaultPoints)
	assert.Equal(t, 35, blueSummary.ParkClimbPoints)
	assert.Equal(t, 55, blueSummary.FoulPoints)
	assert.Equal(t, 228, blueSummary.Score)
	assert.Equal(t, true, blueSummary.AutoQuest)
	assert.Equal(t, false, blueSummary.FaceTheBoss)

	// Test Auto Quest boundary conditions.
	blueScore.AutoEndSwitchOwnership = false
	assert.Equal(t, false, blueScore.Summarize(redScore.Fouls).AutoQuest)
	blueScore.AutoEndSwitchOwnership = true
	blueScore.AutoRuns = 2
	assert.Equal(t, false, blueScore.Summarize(redScore.Fouls).AutoQuest)

	// Test Face the Boss boundary conditions.
	redScore.Levitate = false
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls).FaceTheBoss)
	redScore.Climbs = 3
	assert.Equal(t, true, redScore.Summarize(blueScore.Fouls).FaceTheBoss)
	redScore.Climbs = 1
	redScore.Parks = 2
	assert.Equal(t, false, redScore.Summarize(blueScore.Fouls).FaceTheBoss)

	// Test elimination disqualification.
	redScore.ElimDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore.Fouls).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore.Fouls).Score)
	blueScore.ElimDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore.Fouls).Score)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2.AutoRuns += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoEndSwitchOwnership = !score2.AutoEndSwitchOwnership
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.AutoOwnershipPoints += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TeleopOwnershipPoints += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.VaultCubes += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Levitate = !score2.Levitate
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Parks += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Climbs += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleNumber = "G1000"
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsTechnical = !score2.Fouls[0].IsTechnical
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TimeInMatchSec += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.ElimDq = !score2.ElimDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
