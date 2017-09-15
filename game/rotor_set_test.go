// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRotorsBeforeMatch(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(true, [3]int{15, 15, 0}, matchStartTime, timeAfterStart(-1))
	checkRotorCounts(t, 0, 0, &rotorSet)
}

func TestRotorCountThreshold(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState(true, [3]int{10, 0, 0}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState(true, [3]int{14, 0, 0}, matchStartTime, timeAfterStart(21))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(22))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 13, 0}, matchStartTime, timeAfterStart(23))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 16, 0}, matchStartTime, timeAfterStart(24))
	checkRotorCounts(t, 0, 3, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 16, 50}, matchStartTime, timeAfterStart(25))
	checkRotorCounts(t, 0, 4, &rotorSet)
}

func TestAutoRotors(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(false, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(false, [3]int{15, 15, 15}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 0}, matchStartTime, timeAfterStart(11))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 0}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 2, 1, &rotorSet)

	// Check timing threshold.
	rotorSet = RotorSet{}
	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(15.1))
	checkRotorCounts(t, 1, 1, &rotorSet)
}

func TestTeleopRotors(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(false, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(14))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(30))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 0}, matchStartTime, timeAfterStart(100))
	checkRotorCounts(t, 0, 3, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 15}, matchStartTime, timeAfterStart(120))
	checkRotorCounts(t, 0, 4, &rotorSet)
}

func TestRotorsAfterMatch(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterEnd(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterEnd(2))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 0}, matchStartTime, timeAfterEnd(3))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 15, 15}, matchStartTime, timeAfterEnd(4))
	checkRotorCounts(t, 0, 0, &rotorSet)
}

func TestRotorLatching(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(false, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(2))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 0, 0}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState(false, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(10))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState(true, [3]int{15, 0, 0}, matchStartTime, timeAfterStart(10))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 0, 15}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 15, 0}, matchStartTime, timeAfterStart(30))
	checkRotorCounts(t, 2, 1, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 0, 15}, matchStartTime, timeAfterStart(50))
	checkRotorCounts(t, 2, 2, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 0, 0}, matchStartTime, timeAfterEnd(-1))
	checkRotorCounts(t, 2, 2, &rotorSet)
	rotorSet.UpdateState(false, [3]int{0, 0, 0}, matchStartTime, timeAfterEnd(1))
	checkRotorCounts(t, 2, 2, &rotorSet)
}

func TestRotorActivationOrder(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState(true, [3]int{0, 25, 50}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState(true, [3]int{20, 25, 50}, matchStartTime, timeAfterStart(21))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState(true, [3]int{20, 39, 50}, matchStartTime, timeAfterStart(22))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState(true, [3]int{20, 40, 70}, matchStartTime, timeAfterStart(23))
	checkRotorCounts(t, 0, 3, &rotorSet)
	rotorSet.UpdateState(true, [3]int{20, 40, 84}, matchStartTime, timeAfterStart(24))
	checkRotorCounts(t, 0, 3, &rotorSet)
	rotorSet.UpdateState(true, [3]int{20, 40, 85}, matchStartTime, timeAfterStart(25))
	checkRotorCounts(t, 0, 4, &rotorSet)
}

func checkRotorCounts(t *testing.T, autoRotors, rotors int, rotorSet *RotorSet) {
	assert.Equal(t, autoRotors, rotorSet.AutoRotors)
	assert.Equal(t, rotors, rotorSet.Rotors)
}
