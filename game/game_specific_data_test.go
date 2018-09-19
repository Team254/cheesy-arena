// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestGenerateGameSpecificData(t *testing.T) {
	rand.Seed(0)

	// Make sure all possibilities are hit at least twice.
	assert.Equal(t, "RLR", GenerateGameSpecificData())
	assert.Equal(t, "RLR", GenerateGameSpecificData())
	assert.Equal(t, "LLL", GenerateGameSpecificData())
	assert.Equal(t, "RLR", GenerateGameSpecificData())
	assert.Equal(t, "LRL", GenerateGameSpecificData())
	assert.Equal(t, "RRR", GenerateGameSpecificData())
	assert.Equal(t, "LRL", GenerateGameSpecificData())
	assert.Equal(t, "LLL", GenerateGameSpecificData())
	assert.Equal(t, "RRR", GenerateGameSpecificData())
	assert.Equal(t, "RRR", GenerateGameSpecificData())
}

func TestIsValidGameSpecificData(t *testing.T) {
	for _, data := range validGameSpecificDatas {
		assert.True(t, IsValidGameSpecificData(data))
	}

	assert.False(t, IsValidGameSpecificData(""))
	assert.False(t, IsValidGameSpecificData("R"))
	assert.False(t, IsValidGameSpecificData("RL"))
	assert.False(t, IsValidGameSpecificData("RRL"))
	assert.False(t, IsValidGameSpecificData("RRRL"))
}
