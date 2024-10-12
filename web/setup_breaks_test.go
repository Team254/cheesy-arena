// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetupBreaks(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateScheduledBreak(
		&model.ScheduledBreak{0, model.Playoff, 4, time.Unix(500, 0).UTC(), 900, "Field Break 1"},
	)
	web.arena.Database.CreateScheduledBreak(
		&model.ScheduledBreak{0, model.Playoff, 4, time.Unix(500, 0).UTC(), 900, "Field Break 2"},
	)

	recorder := web.getHttpResponse("/setup/breaks")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Break 1")
	assert.Contains(t, recorder.Body.String(), "Field Break 2")

	recorder = web.postHttpResponse("/setup/breaks", "id=2&description=Award Break 3")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/breaks")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Break 1")
	assert.NotContains(t, recorder.Body.String(), "Field Break 2")
	assert.Contains(t, recorder.Body.String(), "Award Break 3")
}
