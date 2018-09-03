// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupDisplays(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.AllianceStationDisplays["12345"] = ""
	recorder := web.getHttpResponse("/setup/displays")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "12345")
	assert.NotContains(t, recorder.Body.String(), "selected")

	recorder = web.postHttpResponse("/setup/displays", "displayId=12345&allianceStation=B1")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/displays")
	assert.Contains(t, recorder.Body.String(), "12345")
	assert.Contains(t, recorder.Body.String(), "selected")
}
