// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupSponsorSlides(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateSponsorSlide(&model.SponsorSlide{0, "Subtitle", "Sponsor Line 1", "Sponsor Line 2", "", 10})
	web.arena.Database.CreateSponsorSlide(&model.SponsorSlide{0, "Subtitle", "", "", "Image.gif", 10})

	recorder := web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Sponsor Line 1")
	assert.Contains(t, recorder.Body.String(), "Image.gif")

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=delete&id=1")
	assert.Equal(t, 302, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Sponsor Line 1")
	assert.Contains(t, recorder.Body.String(), "Image.gif")

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=save&line2=Sponsor Line 2 revised")
	assert.Equal(t, 302, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Sponsor Line 2 revised")
	sponsorSlide, _ := web.arena.Database.GetSponsorSlideById(3)
	assert.NotNil(t, sponsorSlide)

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=save&image=Image2.gif&id=2")
	assert.Equal(t, 302, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Image.gif")
	assert.Contains(t, recorder.Body.String(), "Image2.gif")
	sponsorSlide, _ = web.arena.Database.GetSponsorSlideById(3)
	assert.NotNil(t, sponsorSlide)
}
