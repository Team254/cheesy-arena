// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOwnership(t *testing.T) {
	ownership := &Ownership{nil, RedAlliance, timeAfterStart(1), nil}
	assertSeconds(t, 0.0, 0.0, ownership, timeAfterStart(0), timeAfterStart(0))
	assertSeconds(t, 0.0, 0.0, ownership, timeAfterStart(0), timeAfterStart(0))
	assertSeconds(t, 0.5, 0.0, ownership, timeAfterStart(0), timeAfterStart(1.5))
	assertSeconds(t, 8.75, 0.0, ownership, timeAfterStart(0), timeAfterStart(9.75))

	// Check with truncated start.
	assertSeconds(t, 2.5, 0.0, ownership, timeAfterStart(1.5), timeAfterStart(4))
	assertSeconds(t, 5.0, 0.0, ownership, timeAfterStart(5), timeAfterStart(10))

	// Check with end time.
	endTime := timeAfterStart(13.5)
	ownership.endTime = &endTime
	assertSeconds(t, 12.5, 0.0, ownership, timeAfterStart(0), timeAfterStart(15))
	assertSeconds(t, 4.0, 0.0, ownership, timeAfterStart(9.5), timeAfterStart(20))

	// Check invalid/corner cases.
	assertSeconds(t, 0.0, 0.0, ownership, timeAfterStart(2), timeAfterStart(1))
}

func TestSecondCounting(t *testing.T) {
	ResetPowerUps()

	redSwitch := &Seesaw{Kind: RedAlliance}
	redSwitch.NearIsRed = true

	// Test that there is no accumulation before the start of the match.
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(-20))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(-12))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(-9))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(-3))
	assertRedSeconds(t, 0.0, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(0))
	assertBlueSeconds(t, 0.0, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(0))

	// Test autonomous.
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(1))
	assertRedSeconds(t, 1.0, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(2))
	assertRedSeconds(t, 5.5, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(6.5))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(8.1))
	assertRedSeconds(t, 7.1, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(8.5))
	assertRedSeconds(t, 7.1, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(10))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(10))
	assertRedSeconds(t, 7.1, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(13))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(13.5))
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(13.9))
	assertRedSeconds(t, 8.2, 0.0, redSwitch, timeAfterStart(0), timeAfterStart(15))

	// Test teleop.
	assertRedSeconds(t, 3.0, 0.0, redSwitch, timeAfterStart(17), timeAfterStart(20))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(30.8))
	assertRedSeconds(t, 13.8, 0.0, redSwitch, timeAfterStart(17), timeAfterStart(34))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(35))
	assertRedSeconds(t, 13.8, 0.0, redSwitch, timeAfterStart(17), timeAfterEnd(-10))
	redSwitch.UpdateState([2]bool{true, false}, timeAfterEnd(-5.1))
	assertRedSeconds(t, 18.9, 0.0, redSwitch, timeAfterStart(17), timeAfterEnd(0))
	assertBlueSeconds(t, 109.9, 0.0, redSwitch, timeAfterStart(17), timeAfterEnd(0))
}

func TestForce(t *testing.T) {
	ResetPowerUps()

	blueSwitch := &Seesaw{Kind: BlueAlliance}
	blueSwitch.NearIsRed = true
	scale := &Seesaw{Kind: NeitherAlliance}
	scale.NearIsRed = true

	// Force switch only.
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(0))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(0))
	powerUp := &PowerUp{Alliance: BlueAlliance, Effect: Force, Level: 1}
	maybeActivatePowerUp(powerUp, timeAfterStart(2.5))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(2.5))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(2.5))
	assertBlueSeconds(t, 2.5, 0.0, blueSwitch, timeAfterStart(0), timeAfterStart(5))
	assertBlueSeconds(t, 0.0, 0.0, scale, timeAfterStart(0), timeAfterStart(5))
	assertBlueSeconds(t, 10.0, 0.0, blueSwitch, timeAfterStart(0), timeAfterStart(12.5))
	assertBlueSeconds(t, 0.0, 0.0, scale, timeAfterStart(0), timeAfterStart(12.5))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(12.5))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(12.5))
	assertBlueSeconds(t, 10.0, 0.0, blueSwitch, timeAfterStart(0), timeAfterStart(15))
	assertBlueSeconds(t, 0.0, 0.0, scale, timeAfterStart(0), timeAfterStart(15))

	// Force scale only.
	powerUp = &PowerUp{Alliance: BlueAlliance, Effect: Force, Level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(20))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(20))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(20))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(30))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(30))
	assertBlueSeconds(t, 0.0, 0.0, blueSwitch, timeAfterStart(20), timeAfterStart(40))
	assertBlueSeconds(t, 10.0, 0.0, scale, timeAfterStart(20), timeAfterStart(40))

	// Force both switch and scale.
	powerUp = &PowerUp{Alliance: BlueAlliance, Effect: Force, Level: 3}
	maybeActivatePowerUp(powerUp, timeAfterStart(50))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(50))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(50))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(60))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(60))
	assertBlueSeconds(t, 10.0, 0.0, blueSwitch, timeAfterStart(50), timeAfterStart(70))
	assertBlueSeconds(t, 10.0, 0.0, scale, timeAfterStart(50), timeAfterStart(70))
}

func TestBoost(t *testing.T) {
	ResetPowerUps()

	blueSwitch := &Seesaw{Kind: BlueAlliance}
	blueSwitch.NearIsRed = true
	scale := &Seesaw{Kind: NeitherAlliance}
	scale.NearIsRed = false

	// Test within continuous ownership period.
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(20))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(20))
	powerUp := &PowerUp{Alliance: BlueAlliance, Effect: Boost, Level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(25))
	assertBlueSeconds(t, 5.0, 0.0, scale, timeAfterStart(0), timeAfterStart(25))
	assertBlueSeconds(t, 5.5, 0.5, scale, timeAfterStart(0), timeAfterStart(25.5))
	assertBlueSeconds(t, 6.25, 1.25, scale, timeAfterStart(0), timeAfterStart(26.25))
	assertBlueSeconds(t, 10.0, 5.0, scale, timeAfterStart(0), timeAfterStart(30))
	assertBlueSeconds(t, 15.0, 10.0, scale, timeAfterStart(0), timeAfterStart(35))
	assertBlueSeconds(t, 20.0, 10.0, scale, timeAfterStart(0), timeAfterStart(40))
	assertBlueSeconds(t, 20.0, 0.0, blueSwitch, timeAfterStart(0), timeAfterStart(40))

	// Test with no ownership at the start.
	ResetPowerUps()
	blueSwitch.UpdateState([2]bool{false, false}, timeAfterStart(44))
	scale.UpdateState([2]bool{false, false}, timeAfterStart(44))
	powerUp = &PowerUp{Alliance: BlueAlliance, Effect: Boost, Level: 3}
	maybeActivatePowerUp(powerUp, timeAfterStart(45))
	assertBlueSeconds(t, 0.0, 0.0, blueSwitch, timeAfterStart(45), timeAfterStart(50))
	assertBlueSeconds(t, 0.0, 0.0, scale, timeAfterStart(45), timeAfterStart(50))
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(50))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(50))
	assertBlueSeconds(t, 5.0, 5.0, blueSwitch, timeAfterStart(45), timeAfterStart(55))
	assertBlueSeconds(t, 10.0, 5.0, blueSwitch, timeAfterStart(45), timeAfterStart(60))
	assertBlueSeconds(t, 5.0, 5.0, scale, timeAfterStart(45), timeAfterStart(55))
	assertBlueSeconds(t, 10.0, 5.0, scale, timeAfterStart(45), timeAfterStart(60))

	// Test with interrupted ownership.
	ResetPowerUps()
	scale.UpdateState([2]bool{false, true}, timeAfterStart(65))
	assertRedSeconds(t, 5.0, 0.0, scale, timeAfterStart(65), timeAfterStart(70))
	powerUp = &PowerUp{Alliance: RedAlliance, Effect: Boost, Level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(70))
	scale.UpdateState([2]bool{false, false}, timeAfterStart(72.5))
	assertRedSeconds(t, 7.5, 2.5, scale, timeAfterStart(65), timeAfterStart(72.5))
	assertRedSeconds(t, 7.5, 2.5, scale, timeAfterStart(65), timeAfterStart(77.5))
	scale.UpdateState([2]bool{false, true}, timeAfterStart(77.5))
	assertRedSeconds(t, 10.0, 5.0, scale, timeAfterStart(65), timeAfterStart(80))
	assertRedSeconds(t, 15.0, 5.0, scale, timeAfterStart(65), timeAfterStart(85))

	// Test with just the switch.
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(100))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(100))
	powerUp = &PowerUp{Alliance: BlueAlliance, Effect: Boost, Level: 1}
	maybeActivatePowerUp(powerUp, timeAfterStart(100))
	assertBlueSeconds(t, 10.0, 10.0, blueSwitch, timeAfterStart(100), timeAfterStart(110))
	assertBlueSeconds(t, 10.0, 0.0, scale, timeAfterStart(100), timeAfterStart(110))
}

func assertSeconds(t *testing.T, expectedOwnership, expectedBoost float64, ownership *Ownership, startTime,
	endTime time.Time) {
	actualOwnership, actualBoost := ownership.getSeconds(startTime, endTime)
	assert.Equal(t, expectedOwnership, actualOwnership)
	assert.Equal(t, expectedBoost, actualBoost)
}

func assertRedSeconds(t *testing.T, expectedOwnership, expectedBoost float64, seesaw *Seesaw, startTime,
	endTime time.Time) {
	actualOwnership, actualBoost := seesaw.GetRedSeconds(startTime, endTime)
	assert.Equal(t, expectedOwnership, actualOwnership)
	assert.Equal(t, expectedBoost, actualBoost)
}

func assertBlueSeconds(t *testing.T, expectedOwnership, expectedBoost float64, seesaw *Seesaw, startTime,
	endTime time.Time) {
	actualOwnership, actualBoost := seesaw.GetBlueSeconds(startTime, endTime)
	assert.Equal(t, expectedOwnership, actualOwnership)
	assert.Equal(t, expectedBoost, actualBoost)
}
