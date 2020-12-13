// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldAssessRung(t *testing.T) {
	assert.Equal(t, false, ShouldAssessRung(matchStartTime, timeAfterStart(0)))
	assert.Equal(t, false, ShouldAssessRung(matchStartTime, timeAfterStart(121.9)))
	assert.Equal(t, true, ShouldAssessRung(matchStartTime, timeAfterStart(122.1)))
	assert.Equal(t, true, ShouldAssessRung(matchStartTime, timeAfterStart(152.1)))
	assert.Equal(t, true, ShouldAssessRung(matchStartTime, timeAfterStart(156.9)))
	assert.Equal(t, false, ShouldAssessRung(matchStartTime, timeAfterStart(157.1)))
}
