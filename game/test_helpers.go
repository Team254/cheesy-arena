// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{true, 25, 13},
		{false, 1868, 14},
		{false, 1868, 14},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
	}
	return &Score{
		LeaveStatuses: [3]bool{true, true, false},
		AmpSpeaker: AmpSpeaker{
			CoopActivated:                 true,
			AutoAmpNotes:                  1,
			TeleopAmpNotes:                4,
			AutoSpeakerNotes:              6,
			TeleopUnamplifiedSpeakerNotes: 1,
			TeleopAmplifiedSpeakerNotes:   5,
		},
		EndgameStatuses:    [3]EndgameStatus{EndgameParked, EndgameNone, EndgameStageLeft},
		MicrophoneStatuses: [3]bool{false, true, true},
		TrapStatuses:       [3]bool{true, true, false},
		Fouls:              fouls,
		PlayoffDq:          false,
	}
}

func TestScore2() *Score {
	return &Score{
		LeaveStatuses: [3]bool{false, true, false},
		AmpSpeaker: AmpSpeaker{
			CoopActivated:                 false,
			AutoAmpNotes:                  0,
			TeleopAmpNotes:                51,
			AutoSpeakerNotes:              8,
			TeleopUnamplifiedSpeakerNotes: 3,
			TeleopAmplifiedSpeakerNotes:   23,
		},
		EndgameStatuses:    [3]EndgameStatus{EndgameStageLeft, EndgameCenterStage, EndgameCenterStage},
		MicrophoneStatuses: [3]bool{false, true, true},
		TrapStatuses:       [3]bool{false, false, false},
		Fouls:              []Foul{},
		PlayoffDq:          false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}
