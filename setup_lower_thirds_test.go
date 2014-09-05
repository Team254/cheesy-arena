// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupLowerThirds(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	mainArena.Setup()

	db.CreateLowerThird(&LowerThird{0, "Top Text 1", "Bottom Text 1"})
	db.CreateLowerThird(&LowerThird{0, "Top Text 2", "Bottom Text 2"})

	recorder := getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Top Text 1")
	assert.Contains(t, recorder.Body.String(), "Bottom Text 2")

	recorder = postHttpResponse("/setup/lower_thirds", "action=delete&id=1")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Top Text 1")
	assert.Contains(t, recorder.Body.String(), "Bottom Text 2")

	recorder = postHttpResponse("/setup/lower_thirds", "action=save&topText=Text 3&bottomText=")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Text 3")
	lowerThird, _ := db.GetLowerThirdById(3)
	assert.NotNil(t, lowerThird)

	recorder = postHttpResponse("/setup/lower_thirds", "action=show&topText=Text 4&bottomText=&id=3")
	assert.Equal(t, 302, recorder.Code)
	recorder = getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Text 4")
	lowerThird, _ = db.GetLowerThirdById(3)
	assert.Equal(t, "Text 4", lowerThird.TopText)
	assert.Equal(t, "", lowerThird.BottomText)

	recorder = postHttpResponse("/setup/lower_thirds", "action=hide&id=3")
	assert.Equal(t, 302, recorder.Code)
}
