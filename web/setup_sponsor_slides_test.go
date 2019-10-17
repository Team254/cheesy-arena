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

	web.arena.Database.CreateSponsorSlide(&model.SponsorSlide{0, "Subtitle", "Sponsor Line 1", "Sponsor Line 2", "", 10,
		0})
	web.arena.Database.CreateSponsorSlide(&model.SponsorSlide{0, "Subtitle", "", "", "Image.gif", 10,
		1})

	recorder := web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Sponsor Line 1")
	assert.Contains(t, recorder.Body.String(), "Image.gif")

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=delete&id=1")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Sponsor Line 1")
	assert.Contains(t, recorder.Body.String(), "Image.gif")

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=save&line2=Sponsor Line 2 revised")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Sponsor Line 2 revised")
	sponsorSlide, _ := web.arena.Database.GetSponsorSlideById(3)
	assert.NotNil(t, sponsorSlide)

	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=save&image=Image2.gif&id=2")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/sponsor_slides")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Image.gif")
	assert.Contains(t, recorder.Body.String(), "Image2.gif")
	sponsorSlide, _ = web.arena.Database.GetSponsorSlideById(3)
	assert.NotNil(t, sponsorSlide)

	sponsorSlides1, _ := web.arena.Database.GetAllSponsorSlides()
	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=reorderUp&id=3")
	assert.Equal(t, 303, recorder.Code)
	sponsorSlides2, _ := web.arena.Database.GetAllSponsorSlides()
	if assert.Equal(t, 2, len(sponsorSlides1)) && assert.Equal(t, 2, len(sponsorSlides2)) {
		assert.Equal(t, sponsorSlides1[0].Id, sponsorSlides2[1].Id)
		assert.Equal(t, sponsorSlides1[1].Id, sponsorSlides2[0].Id)
	}
	recorder = web.postHttpResponse("/setup/sponsor_slides", "action=reorderDown&id=3")
	assert.Equal(t, 303, recorder.Code)
	sponsorSlides3, _ := web.arena.Database.GetAllSponsorSlides()
	if assert.Equal(t, 2, len(sponsorSlides1)) && assert.Equal(t, 2, len(sponsorSlides2)) {
		assert.Equal(t, sponsorSlides1[0].Id, sponsorSlides3[0].Id)
		assert.Equal(t, sponsorSlides1[1].Id, sponsorSlides3[1].Id)
	}
}
