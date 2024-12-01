// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package partner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBlackmagicClient(t *testing.T) {
	// Test with an empty address.
	client := NewBlackmagicClient("")
	assert.Equal(t, 0, len(client.deviceAddresses))

	// Test with whitespace in the address.
	client = NewBlackmagicClient("  ")
	assert.Equal(t, 0, len(client.deviceAddresses))

	// Test with a single address.
	client = NewBlackmagicClient("1.2.3.4")
	if assert.Equal(t, 1, len(client.deviceAddresses)) {
		assert.Equal(t, "1.2.3.4", client.deviceAddresses[0])
	}

	// Test with multiple addresses.
	client = NewBlackmagicClient(" 1.2.3.4 ,  5.6.7.8   ")
	if assert.Equal(t, 2, len(client.deviceAddresses)) {
		assert.Equal(t, "1.2.3.4", client.deviceAddresses[0])
		assert.Equal(t, "5.6.7.8", client.deviceAddresses[1])
	}
}
