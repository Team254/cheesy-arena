// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package tournament

import (
	"github.com/Team254/cheesy-arena/model"
	"testing"
)

func CreateTestAlliances(database *model.Database, allianceCount int) {
	for i := 1; i <= allianceCount; i++ {
		database.CreateAllianceTeam(&model.AllianceTeam{0, i, 0, i})
		database.CreateAllianceTeam(&model.AllianceTeam{0, i, 1, 10 * i})
		database.CreateAllianceTeam(&model.AllianceTeam{0, i, 2, 100 * i})
	}
}

func setupTestDb(t *testing.T) *model.Database {
	return model.SetupTestDb(t, "tournament")
}
