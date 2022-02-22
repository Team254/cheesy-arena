// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var matchStartTime = time.Unix(10, 0)

func TestHub(t *testing.T) {
	var hub Hub
	assertHub(t, &hub, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})

	// Check before match start and during the autonomous period.
	hub.UpdateState([4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8}, matchStartTime, timeAfterStart(-1))
	assertHub(t, &hub, [4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	hub.UpdateState([4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, matchStartTime, timeAfterStart(1))
	assertHub(t, &hub, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	hub.UpdateState([4]int{5, 6, 7, 8}, [4]int{1, 2, 3, 4}, matchStartTime, timeAfterStart(2))
	assertHub(t, &hub, [4]int{5, 6, 7, 8}, [4]int{1, 2, 3, 4}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	hub.UpdateState([4]int{6, 7, 8, 9}, [4]int{2, 3, 4, 5}, matchStartTime, timeAfterStart(5))
	assertHub(t, &hub, [4]int{6, 7, 8, 9}, [4]int{2, 3, 4, 5}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})

	// Check boundary conditions around the auto end grace period.
	hub.UpdateState([4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, matchStartTime, timeAfterStart(19.9))
	assertHub(t, &hub, [4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	hub.UpdateState([4]int{8, 9, 9, 9}, [4]int{4, 5, 6, 7}, matchStartTime, timeAfterStart(20.1))
	assertHub(t, &hub, [4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, [4]int{1, 1, 0, 0}, [4]int{1, 1, 1, 1})

	// Check during the teleoperated period.
	hub.UpdateState([4]int{9, 9, 9, 9}, [4]int{9, 8, 7, 7}, matchStartTime, timeAfterStart(25))
	assertHub(t, &hub, [4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, [4]int{2, 1, 0, 0}, [4]int{6, 4, 2, 1})

	// Check boundary conditions around the teleop end grace period.
	hub.UpdateState([4]int{10, 11, 12, 13}, [4]int{14, 15, 16, 17}, matchStartTime, timeAfterStart(161.9))
	assertHub(t, &hub, [4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, [4]int{3, 3, 3, 4}, [4]int{11, 11, 11, 11})
	hub.UpdateState([4]int{11, 12, 13, 14}, [4]int{15, 16, 17, 18}, matchStartTime, timeAfterStart(162.1))
	assertHub(t, &hub, [4]int{7, 8, 9, 9}, [4]int{3, 4, 5, 6}, [4]int{3, 3, 3, 4}, [4]int{11, 11, 11, 11})
}

func assertHub(
	t *testing.T,
	hub *Hub,
	expectedAutoCargoLower [4]int,
	expectedAutoCargoUpper [4]int,
	expectedTeleopCargoLower [4]int,
	expectedTeleopCargoUpper [4]int,
) {
	for i := 0; i < 4; i++ {
		assert.Equal(t, expectedAutoCargoLower[i], hub.AutoCargoLower[i])
		assert.Equal(t, expectedAutoCargoUpper[i], hub.AutoCargoUpper[i])
		assert.Equal(t, expectedTeleopCargoLower[i], hub.TeleopCargoLower[i])
		assert.Equal(t, expectedTeleopCargoUpper[i], hub.TeleopCargoUpper[i])
	}
}

func timeAfterStart(sec float32) time.Time {
	return matchStartTime.Add(time.Duration(1000*sec) * time.Millisecond)
}
