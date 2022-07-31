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
		alliance := model.Alliance{
			Id:      i,
			TeamIds: []int{100*i + 1, 100*i + 2, 100*i + 3, 100*i + 4},
			Lineup:  [3]int{100*i + 2, 100*i + 1, 100*i + 3},
		}
		database.CreateAlliance(&alliance)
	}
}

func setupTestDb(t *testing.T) *model.Database {
	return model.SetupTestDb(t, "tournament")
}
