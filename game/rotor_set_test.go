// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRotorsBeforeMatch(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState([4]bool{true, true, true, false}, matchStartTime, timeAfterStart(-1))
	checkRotorCounts(t, 0, 0, &rotorSet)
}

func TestAutoRotors(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState([4]bool{false, false, false, false}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{false, true, true, true}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, false, false, false}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, false}, matchStartTime, timeAfterStart(11))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, false}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 2, 1, &rotorSet)

	// Check going straight to two.
	rotorSet = RotorSet{}
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 2, 0, &rotorSet)

	// Check timing threshold.
	rotorSet = RotorSet{}
	rotorSet.UpdateState([4]bool{true, false, false, false}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterStart(15.1))
	checkRotorCounts(t, 1, 1, &rotorSet)
}

func TestTeleopRotors(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState([4]bool{false, false, false, false}, matchStartTime, timeAfterStart(14))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, false, false, false}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 0, 1, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterStart(30))
	checkRotorCounts(t, 0, 2, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, false}, matchStartTime, timeAfterStart(100))
	checkRotorCounts(t, 0, 3, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, true}, matchStartTime, timeAfterStart(120))
	checkRotorCounts(t, 0, 4, &rotorSet)
}

func TestRotorsAfterMatch(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState([4]bool{true, false, false, false}, matchStartTime, timeAfterEnd(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterEnd(2))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, false}, matchStartTime, timeAfterEnd(3))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, true, true}, matchStartTime, timeAfterEnd(4))
	checkRotorCounts(t, 0, 0, &rotorSet)
}

func TestRotorLatching(t *testing.T) {
	rotorSet := RotorSet{}

	rotorSet.UpdateState([4]bool{false, true, false, false}, matchStartTime, timeAfterStart(1))
	checkRotorCounts(t, 0, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, false, false, false}, matchStartTime, timeAfterStart(2))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, false, false}, matchStartTime, timeAfterStart(5))
	checkRotorCounts(t, 1, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{false, true, false, false}, matchStartTime, timeAfterStart(10))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{true, true, false, false}, matchStartTime, timeAfterStart(10))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, false, true}, matchStartTime, timeAfterStart(20))
	checkRotorCounts(t, 2, 0, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, true, false}, matchStartTime, timeAfterStart(30))
	checkRotorCounts(t, 2, 1, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, false, true}, matchStartTime, timeAfterStart(50))
	checkRotorCounts(t, 2, 2, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, false, false}, matchStartTime, timeAfterEnd(-1))
	checkRotorCounts(t, 2, 2, &rotorSet)
	rotorSet.UpdateState([4]bool{false, false, false, false}, matchStartTime, timeAfterEnd(1))
	checkRotorCounts(t, 2, 2, &rotorSet)
}

func checkRotorCounts(t *testing.T, autoRotors, rotors int, rotorSet *RotorSet) {
	assert.Equal(t, autoRotors, rotorSet.AutoRotors)
	assert.Equal(t, rotors, rotorSet.Rotors)
}
