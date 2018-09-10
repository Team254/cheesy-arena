// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFtaDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/fta?displayId=1")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Field Monitor - Untitled Event - Cheesy Arena")
}
