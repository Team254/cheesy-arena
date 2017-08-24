// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Data for use in tests in this package and others.

package model

import "github.com/Team254/cheesy-arena/game"

func BuildTestMatchResult(matchId int, playNumber int) *MatchResult {
	matchResult := &MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = game.TestScore1()
	matchResult.BlueScore = game.TestScore2()
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}

func BuildTestAlliances(db *Database) {
	db.CreateAllianceTeam(&AllianceTeam{0, 2, 0, 1718})
	db.CreateAllianceTeam(&AllianceTeam{0, 1, 3, 74})
	db.CreateAllianceTeam(&AllianceTeam{0, 1, 1, 469})
	db.CreateAllianceTeam(&AllianceTeam{0, 1, 0, 254})
	db.CreateAllianceTeam(&AllianceTeam{0, 1, 2, 2848})
	db.CreateAllianceTeam(&AllianceTeam{0, 2, 1, 2451})
}
