// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package model

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func SetupTestDb(t *testing.T, uniqueName string) *Database {
	BaseDir = ".."
	dbPath := filepath.Join(BaseDir, fmt.Sprintf("%s_test.db", uniqueName))
	os.Remove(dbPath)
	database, err := OpenDatabase(dbPath)
	assert.Nil(t, err)
	return database
}

func BuildTestMatchResult(matchId int, playNumber int) *MatchResult {
	matchResult := &MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = game.TestScore1()
	matchResult.BlueScore = game.TestScore2()
	matchResult.RedCards = map[string]string{"1868": "yellow"}
	matchResult.BlueCards = map[string]string{}
	return matchResult
}

func BuildTestAlliances(database *Database) {
	database.CreateAllianceTeam(&AllianceTeam{0, 2, 0, 1718})
	database.CreateAllianceTeam(&AllianceTeam{0, 1, 3, 74})
	database.CreateAllianceTeam(&AllianceTeam{0, 1, 1, 469})
	database.CreateAllianceTeam(&AllianceTeam{0, 1, 0, 254})
	database.CreateAllianceTeam(&AllianceTeam{0, 1, 2, 2848})
	database.CreateAllianceTeam(&AllianceTeam{0, 2, 1, 2451})
}
