// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollectMatchGroupsErrors(t *testing.T) {
	// Duplicate match group ID.
	matchGroup1 := Matchup{
		id:                 "M1",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSelectionSource{2},
		blueAllianceSource: allianceSelectionSource{3},
		matchSpecs:         newDoubleEliminationMatch(1, "", 300),
	}
	matchGroup2 := Matchup{
		id:                 "M1",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSelectionSource{1},
		blueAllianceSource: matchupSource{&matchGroup1, true},
		matchSpecs:         newDoubleEliminationMatch(2, "", 300),
	}

	_, err := collectMatchGroups(&matchGroup2)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "ID \"M1\" defined more than once")
	}
}

func TestCollectMatchSpecsErrors(t *testing.T) {
	match1 := matchSpec{
		longName:    "Final 1",
		shortName:   "F1",
		order:       1,
		tbaMatchKey: model.TbaMatchKey{"f", 1, 1},
	}
	match2 := matchSpec{
		longName:    "Final 2",
		shortName:   "F2",
		order:       2,
		tbaMatchKey: model.TbaMatchKey{"f", 1, 2},
	}
	match3 := matchSpec{
		longName:    "Final 3",
		shortName:   "F3",
		order:       3,
		tbaMatchKey: model.TbaMatchKey{"f", 1, 3},
	}

	// No errors to start.
	matchGroup1 := Matchup{
		id:                 "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  allianceSelectionSource{1},
		blueAllianceSource: allianceSelectionSource{2},
		matchSpecs:         []*matchSpec{&match3, &match2, &match1},
	}
	matchSpecs, err := collectMatchSpecs(&matchGroup1)
	assert.Nil(t, err)
	assert.Equal(t, []*matchSpec{&match1, &match2, &match3}, matchSpecs)

	// Duplicate long name.
	match3.longName = "Final 1"
	_, err = collectMatchSpecs(&matchGroup1)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "long name \"Final 1\" defined more than once")
	}

	// Duplicate short name.
	match3.longName = "Final 3"
	match3.shortName = "F1"
	_, err = collectMatchSpecs(&matchGroup1)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "short name \"F1\" defined more than once")
	}

	// Duplicate order.
	match3.shortName = "F3"
	match3.order = 1
	_, err = collectMatchSpecs(&matchGroup1)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "order 1 defined more than once")
	}

	// Duplicate TBA match key.
	match3.order = 3
	match3.tbaMatchKey = model.TbaMatchKey{"f", 1, 1}
	_, err = collectMatchSpecs(&matchGroup1)
	if assert.NotNil(t, err) {
		assert.Regexp(t, "TBA key .* defined more than once", err.Error())
	}
}
