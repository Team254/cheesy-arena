// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const testDbPath = "test.db"

func setupTest(t *testing.T) {
	os.Remove(testDbPath)
	var err error
	db, err = model.OpenDatabase(".", testDbPath)
	assert.Nil(t, err)
	eventSettings, err = db.GetEventSettings()
	assert.Nil(t, err)
	mainArena.Setup()
}
