// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOwnership(t *testing.T) {
	ownership := Ownership{nil, redAlliance, timeAfterStart(1), nil}
	assert.Equal(t, 0, ownership.getSeconds(timeAfterStart(0), timeAfterStart(0), true))
	assert.Equal(t, 0.5, ownership.getSeconds(timeAfterStart(0), timeAfterStart(1.5), true))
	assert.Equal(t, 8.75, ownership.getSeconds(timeAfterStart(0), timeAfterStart(9.75), true))

	// Check with truncated start.
	assert.Equal(t, 2.5, ownership.getSeconds(timeAfterStart(1.5), timeAfterStart(4), true))
	assert.Equal(t, 5, ownership.getSeconds(timeAfterStart(5), timeAfterStart(10), true))

	// Check with end time.
	endTime := timeAfterStart(13.5)
	ownership.endTime = &endTime
	assert.Equal(t, 12.5, ownership.getSeconds(timeAfterStart(0), timeAfterStart(15), true))
	assert.Equal(t, 4, ownership.getSeconds(timeAfterStart(9.5), timeAfterStart(20), true))

	// Check invalid/corner cases.
	assert.Equal(t, 0, ownership.getSeconds(timeAfterStart(2), timeAfterStart(1), true))
}

func TestSecondCounting(t *testing.T) {
	ResetPowerUps()

	redSwitch := &Seesaw{kind: redAlliance}
	redSwitch.SetRandomization(true)

	// Test that there is no accumulation before the start of the match.
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(-20))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(-12))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(-9))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(-3))
	assert.Equal(t, 0, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(0)))
	assert.Equal(t, 0, redSwitch.GetBlueSeconds(timeAfterStart(0), timeAfterStart(0)))

	// Test autonomous.
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(1))
	assert.Equal(t, 1, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(2)))
	assert.Equal(t, 5.5, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(6.5)))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(8.1))
	assert.Equal(t, 7.1, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(8.5)))
	assert.Equal(t, 7.1, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(10)))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(10))
	assert.Equal(t, 7.1, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(13)))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(13.5))
	redSwitch.UpdateState([2]bool{true, false}, timeAfterStart(13.9))
	assert.Equal(t, 8.2, redSwitch.GetRedSeconds(timeAfterStart(0), timeAfterStart(15)))

	// Test teleop.
	assert.Equal(t, 3, redSwitch.GetRedSeconds(timeAfterStart(17), timeAfterStart(20)))
	redSwitch.UpdateState([2]bool{false, false}, timeAfterStart(30.8))
	assert.Equal(t, 13.8, redSwitch.GetRedSeconds(timeAfterStart(17), timeAfterStart(34)))
	redSwitch.UpdateState([2]bool{false, true}, timeAfterStart(35))
	assert.Equal(t, 13.8, redSwitch.GetRedSeconds(timeAfterStart(17), timeAfterEnd(-10)))
	redSwitch.UpdateState([2]bool{true, false}, timeAfterEnd(-5.1))
	assert.Equal(t, 18.9, redSwitch.GetRedSeconds(timeAfterStart(17), timeAfterEnd(0)))
	assert.Equal(t, 111.9, redSwitch.GetBlueSeconds(timeAfterStart(17), timeAfterEnd(0)))
}

func TestForce(t *testing.T) {
	ResetPowerUps()

	blueSwitch := &Seesaw{kind: blueAlliance}
	blueSwitch.SetRandomization(true)
	scale := &Seesaw{kind: neitherAlliance}
	scale.SetRandomization(true)

	// Force switch only.
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(0))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(0))
	powerUp := &PowerUp{alliance: blueAlliance, kind: force, level: 1}
	maybeActivatePowerUp(powerUp, timeAfterStart(2.5))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(2.5))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(2.5))
	assert.Equal(t, 2.5, blueSwitch.GetBlueSeconds(timeAfterStart(0), timeAfterStart(5)))
	assert.Equal(t, 0, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(5)))
	assert.Equal(t, 10, blueSwitch.GetBlueSeconds(timeAfterStart(0), timeAfterStart(12.5)))
	assert.Equal(t, 0, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(12.5)))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(12.5))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(12.5))
	assert.Equal(t, 10, blueSwitch.GetBlueSeconds(timeAfterStart(0), timeAfterStart(15)))
	assert.Equal(t, 0, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(15)))

	// Force scale only.
	powerUp = &PowerUp{alliance: blueAlliance, kind: force, level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(20))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(20))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(20))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(30))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(30))
	assert.Equal(t, 0, blueSwitch.GetBlueSeconds(timeAfterStart(20), timeAfterStart(40)))
	assert.Equal(t, 10, scale.GetBlueSeconds(timeAfterStart(20), timeAfterStart(40)))

	// Force both switch and scale.
	powerUp = &PowerUp{alliance: blueAlliance, kind: force, level: 3}
	maybeActivatePowerUp(powerUp, timeAfterStart(50))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(50))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(50))
	blueSwitch.UpdateState([2]bool{true, false}, timeAfterStart(60))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(60))
	assert.Equal(t, 10, blueSwitch.GetBlueSeconds(timeAfterStart(50), timeAfterStart(70)))
	assert.Equal(t, 10, scale.GetBlueSeconds(timeAfterStart(50), timeAfterStart(70)))
}

func TestBoost(t *testing.T) {
	ResetPowerUps()

	blueSwitch := &Seesaw{kind: blueAlliance}
	blueSwitch.SetRandomization(true)
	scale := &Seesaw{kind: neitherAlliance}
	scale.SetRandomization(false)

	// Test within continuous ownership period.
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(20))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(20))
	powerUp := &PowerUp{alliance: blueAlliance, kind: boost, level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(25))
	assert.Equal(t, 5, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(25)))
	assert.Equal(t, 6, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(25.5)))
	assert.Equal(t, 7.5, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(26.25)))
	assert.Equal(t, 15, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(30)))
	assert.Equal(t, 25, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(35)))
	assert.Equal(t, 30, scale.GetBlueSeconds(timeAfterStart(0), timeAfterStart(40)))
	assert.Equal(t, 20, blueSwitch.GetBlueSeconds(timeAfterStart(0), timeAfterStart(40)))

	// Test with no ownership at the start.
	ResetPowerUps()
	blueSwitch.UpdateState([2]bool{false, false}, timeAfterStart(44))
	scale.UpdateState([2]bool{false, false}, timeAfterStart(44))
	powerUp = &PowerUp{alliance: blueAlliance, kind: boost, level: 3}
	maybeActivatePowerUp(powerUp, timeAfterStart(45))
	assert.Equal(t, 0, blueSwitch.GetBlueSeconds(timeAfterStart(45), timeAfterStart(50)))
	assert.Equal(t, 0, scale.GetBlueSeconds(timeAfterStart(45), timeAfterStart(50)))
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(50))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(50))
	assert.Equal(t, 10, blueSwitch.GetBlueSeconds(timeAfterStart(45), timeAfterStart(55)))
	assert.Equal(t, 15, blueSwitch.GetBlueSeconds(timeAfterStart(45), timeAfterStart(60)))
	assert.Equal(t, 10, scale.GetBlueSeconds(timeAfterStart(45), timeAfterStart(55)))
	assert.Equal(t, 15, scale.GetBlueSeconds(timeAfterStart(45), timeAfterStart(60)))

	// Test with interrupted ownership.
	ResetPowerUps()
	scale.UpdateState([2]bool{false, true}, timeAfterStart(65))
	assert.Equal(t, 5, scale.GetRedSeconds(timeAfterStart(65), timeAfterStart(70)))
	powerUp = &PowerUp{alliance: redAlliance, kind: boost, level: 2}
	maybeActivatePowerUp(powerUp, timeAfterStart(70))
	scale.UpdateState([2]bool{false, false}, timeAfterStart(72.5))
	assert.Equal(t, 10, scale.GetRedSeconds(timeAfterStart(65), timeAfterStart(72.5)))
	assert.Equal(t, 10, scale.GetRedSeconds(timeAfterStart(65), timeAfterStart(77.5)))
	scale.UpdateState([2]bool{false, true}, timeAfterStart(77.5))
	assert.Equal(t, 15, scale.GetRedSeconds(timeAfterStart(65), timeAfterStart(80)))
	assert.Equal(t, 20, scale.GetRedSeconds(timeAfterStart(65), timeAfterStart(85)))

	// Test with just the switch.
	blueSwitch.UpdateState([2]bool{false, true}, timeAfterStart(100))
	scale.UpdateState([2]bool{true, false}, timeAfterStart(100))
	powerUp = &PowerUp{alliance: blueAlliance, kind: boost, level: 1}
	maybeActivatePowerUp(powerUp, timeAfterStart(100))
	assert.Equal(t, 20, blueSwitch.GetBlueSeconds(timeAfterStart(100), timeAfterStart(110)))
	assert.Equal(t, 10, scale.GetBlueSeconds(timeAfterStart(100), timeAfterStart(110)))
}
