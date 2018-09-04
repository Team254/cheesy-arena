// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package plc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByteToBool(t *testing.T) {
	bytes := []byte{7, 254, 3}
	bools := byteToBool(bytes, 17)
	if assert.Equal(t, 17, len(bools)) {
		expectedBools := []bool{true, true, true, false, false, false, false, false, false, true, true, true, true,
			true, true, true, true}
		assert.Equal(t, expectedBools, bools)
	}
}

func TestByteToUint(t *testing.T) {
	bytes := []byte{1, 77, 2, 253, 21, 179}
	uints := byteToUint(bytes, 3)
	if assert.Equal(t, 3, len(uints)) {
		assert.Equal(t, []uint16{333, 765, 5555}, uints)
	}
}

func TestBoolToByte(t *testing.T) {
	bools := []bool{true, true, false, false, true, false, false, false, false, true}
	bytes := boolToByte(bools)
	if assert.Equal(t, 2, len(bytes)) {
		assert.Equal(t, []byte{19, 2}, bytes)
		assert.Equal(t, bools, byteToBool(bytes, len(bools)))
	}
}
