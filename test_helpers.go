// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupTest(t *testing.T) {
	db = model.SetupTestDb(t, "main", ".")
	var err error
	eventSettings, err = db.GetEventSettings()
	assert.Nil(t, err)
	tbaClient = partner.NewTbaClient(eventSettings.TbaEventCode, eventSettings.TbaSecretId, eventSettings.TbaSecret)
	stemTvClient = partner.NewStemTvClient(eventSettings.StemTvEventCode)
	mainArena.Setup()
}
