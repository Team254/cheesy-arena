// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const testDbPath = "test.db"

func TestOpenUnreachableDatabase(t *testing.T) {
	_, err := OpenDatabase("..", "nonexistentdir/test.db")
	assert.NotNil(t, err)
}

func setupTestDb(t *testing.T) *Database {
	os.Remove(filepath.Join("..", testDbPath))
	db, err := OpenDatabase("..", testDbPath)
	assert.Nil(t, err)
	return db
}
