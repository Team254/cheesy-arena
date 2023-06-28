// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package playoff

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupTestDb(t *testing.T) *model.Database {
	return model.SetupTestDb(t, "playoff")
}

type expectedMatchSpec struct {
	longName            string
	shortName           string
	nameDetail          string
	order               int
	matchGroupId        string
	useTiebreakCriteria bool
	isHidden            bool
	tbaCompLevel        string
	tbaSetNumber        int
	tbaMatchNumber      int
}

func assertMatchSpecs(
	t *testing.T,
	matchSpecs []*matchSpec,
	expected []expectedMatchSpec,
) {
	if assert.Equal(t, len(expected), len(matchSpecs)) {
		for i, expectedValue := range expected {
			assert.Equal(t, expectedValue.longName, matchSpecs[i].longName)
			assert.Equal(t, expectedValue.shortName, matchSpecs[i].shortName)
			assert.Equal(t, expectedValue.nameDetail, matchSpecs[i].nameDetail)
			assert.Equal(t, expectedValue.matchGroupId, matchSpecs[i].matchGroupId)
			assert.Equal(t, expectedValue.order, matchSpecs[i].order)
			assert.Equal(t, expectedValue.useTiebreakCriteria, matchSpecs[i].useTiebreakCriteria)
			assert.Equal(t, expectedValue.isHidden, matchSpecs[i].isHidden)
			assert.Equal(t, expectedValue.tbaCompLevel, matchSpecs[i].tbaMatchKey.CompLevel)
			assert.Equal(t, expectedValue.tbaSetNumber, matchSpecs[i].tbaMatchKey.SetNumber)
			assert.Equal(t, expectedValue.tbaMatchNumber, matchSpecs[i].tbaMatchKey.MatchNumber)
			assert.Equal(t, 0, matchSpecs[i].redAllianceId)
			assert.Equal(t, 0, matchSpecs[i].blueAllianceId)
		}
	}
}

type expectedAlliances struct {
	redAllianceId  int
	blueAllianceId int
}

func assertMatchSpecAlliances(
	t *testing.T,
	matchSpecs []*matchSpec,
	expected []expectedAlliances,
) {
	if assert.Equal(t, len(expected), len(matchSpecs)) {
		for i, alliance := range expected {
			assert.Equal(t, alliance.redAllianceId, matchSpecs[i].redAllianceId)
			assert.Equal(t, alliance.blueAllianceId, matchSpecs[i].blueAllianceId)
		}
	}
}

func assertMatchGroups(
	t *testing.T,
	matchGroups map[string]MatchGroup,
	expectedMatchGroupIds ...string,
) {
	assert.Equal(t, len(expectedMatchGroupIds), len(matchGroups))
	for _, expectedMatchGroupId := range expectedMatchGroupIds {
		assert.Contains(t, matchGroups, expectedMatchGroupId)
	}
}

func assertMatch(
	t *testing.T,
	match model.Match,
	typeOrder int,
	timeSec int64,
	longName string,
	shortName string,
	nameDetail string,
	matchGroupId string,
	redAlliance int,
	blueAlliance int,
	useTiebreakCriteria bool,
	tbaCompLevel string,
	tbaSetNumber int,
	tbaMatchNumber int,
) {
	assert.Equal(t, model.Playoff, match.Type)
	assert.Equal(t, typeOrder, match.TypeOrder)
	assert.Equal(t, timeSec, match.Time.Unix())
	assert.Equal(t, longName, match.LongName)
	assert.Equal(t, shortName, match.ShortName)
	assert.Equal(t, nameDetail, match.NameDetail)
	assert.Equal(t, matchGroupId, match.PlayoffMatchGroupId)
	assert.Equal(t, redAlliance, match.PlayoffRedAlliance)
	assert.Equal(t, blueAlliance, match.PlayoffBlueAlliance)
	if redAlliance == 0 {
		assert.Equal(t, 0, match.Red1)
		assert.Equal(t, 0, match.Red2)
		assert.Equal(t, 0, match.Red3)
	} else {
		assert.Equal(t, 100*redAlliance+2, match.Red1)
		assert.Equal(t, 100*redAlliance+1, match.Red2)
		assert.Equal(t, 100*redAlliance+3, match.Red3)
	}
	if blueAlliance == 0 {
		assert.Equal(t, 0, match.Blue1)
		assert.Equal(t, 0, match.Blue2)
		assert.Equal(t, 0, match.Blue3)
	} else {
		assert.Equal(t, 100*blueAlliance+2, match.Blue1)
		assert.Equal(t, 100*blueAlliance+1, match.Blue2)
		assert.Equal(t, 100*blueAlliance+3, match.Blue3)
	}
	assert.Equal(t, useTiebreakCriteria, match.UseTiebreakCriteria)
	assert.Equal(t, tbaCompLevel, match.TbaMatchKey.CompLevel)
	assert.Equal(t, tbaSetNumber, match.TbaMatchKey.SetNumber)
	assert.Equal(t, tbaMatchNumber, match.TbaMatchKey.MatchNumber)
}

func assertBreak(
	t *testing.T,
	scheduledBreak model.ScheduledBreak,
	typeOrderBefore int,
	timeSec int64,
	durationSec int,
	description string,
) {
	assert.Equal(t, model.Playoff, scheduledBreak.MatchType)
	assert.Equal(t, typeOrderBefore, scheduledBreak.TypeOrderBefore)
	assert.Equal(t, timeSec, scheduledBreak.Time.Unix())
	assert.Equal(t, durationSec, scheduledBreak.DurationSec)
	assert.Equal(t, description, scheduledBreak.Description)
}

func assertMatchupOutcome(t *testing.T, matchGroup MatchGroup, redDestination string, blueDestination string) {
	matchup, ok := matchGroup.(*Matchup)
	if assert.True(t, ok) {
		assert.Equal(t, redDestination, matchup.RedAllianceDestination())
		assert.Equal(t, blueDestination, matchup.BlueAllianceDestination())
	}
}
