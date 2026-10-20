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

	web.arena.Database.CreateAward(&model.Award{Type: model.JudgedAward, AwardName: "Spirit Award"})
	web.arena.Database.CreateAward(&model.Award{Type: model.JudgedAward, AwardName: "Saftey Award"})

	recorder := web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Spirit Award")
	assert.Contains(t, recorder.Body.String(), "Saftey Award")
	assert.NotContains(t, recorder.Body.String(), "tbaTypeEnum")
	assert.NotContains(
		t, recorder.Body.String(),
		"https://github.com/the-blue-alliance/the-blue-alliance/blob/main/src/backend/common/consts/award_type.py",
	)

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

	// Enable TBA publishing and verify the TBA Award Type field is shown and saved.
	web.arena.EventSettings.TbaPublishingEnabled = true
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "tbaTypeEnum")
	assert.Contains(
		t, recorder.Body.String(),
		"https://github.com/the-blue-alliance/the-blue-alliance/blob/main/src/backend/common/consts/award_type.py",
	)

	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&personName=Englebert&tbaTypeEnum=18")
	assert.Equal(t, 303, recorder.Code)
	award, err := web.arena.Database.GetAwardById(2)
	assert.Nil(t, err)
	if assert.NotNil(t, award.TbaTypeEnum) {
		assert.Equal(t, 18, *award.TbaTypeEnum)
	}
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `name="tbaTypeEnum" value="18"`)

	// Saving tbaTypeEnum=0 (e.g. Chairman's Award) persists 0 rather than nil.
	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&personName=Englebert&tbaTypeEnum=0")
	assert.Equal(t, 303, recorder.Code)
	award, err = web.arena.Database.GetAwardById(2)
	assert.Nil(t, err)
	if assert.NotNil(t, award.TbaTypeEnum) {
		assert.Equal(t, 0, *award.TbaTypeEnum)
	}
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `name="tbaTypeEnum" value="0"`)

	// Updating the award without tbaTypeEnum in the form (e.g. when TBA publishing is disabled) preserves TbaTypeEnum.
	web.arena.EventSettings.TbaPublishingEnabled = false
	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&personName=Humperdinck")
	assert.Equal(t, 303, recorder.Code)
	award, err = web.arena.Database.GetAwardById(2)
	assert.Nil(t, err)
	assert.Equal(t, "Humperdinck", award.PersonName)
	if assert.NotNil(t, award.TbaTypeEnum) {
		assert.Equal(t, 0, *award.TbaTypeEnum)
	}

	// Clearing the TBA Award Type resets it to nil.
	web.arena.EventSettings.TbaPublishingEnabled = true
	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&personName=Humperdinck&tbaTypeEnum=")
	assert.Equal(t, 303, recorder.Code)
	award, err = web.arena.Database.GetAwardById(2)
	assert.Nil(t, err)
	assert.Nil(t, award.TbaTypeEnum)

	// Invalid or negative TBA Award Type values return an error.
	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&tbaTypeEnum=abc")
	assert.Equal(t, 500, recorder.Code)
	recorder = web.postHttpResponse("/setup/awards", "id=2&awardName=Saftey+Award&tbaTypeEnum=-1")
	assert.Equal(t, 500, recorder.Code)
}
