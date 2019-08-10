// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupAwards(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateAward(&model.Award{0, model.JudgedAward, "Spirit Award", 0, ""})
	web.arena.Database.CreateAward(&model.Award{0, model.JudgedAward, "Saftey Award", 0, ""})

	recorder := web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Spirit Award")
	assert.Contains(t, recorder.Body.String(), "Saftey Award")

	recorder = web.postHttpResponse("/setup/awards", "action=delete&id=1")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Spirit Award")
	assert.Contains(t, recorder.Body.String(), "Saftey Award")

	recorder = web.postHttpResponse("/setup/awards", "awardId=2&awardName=Saftey+Award&personName=Englebert")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Englebert")
}
