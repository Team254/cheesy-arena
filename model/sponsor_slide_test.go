// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentSponsorSlide(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	sponsorSlide, err := db.GetSponsorSlideById(1114)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide)
}

func TestSponsorSlideCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	assert.Equal(t, 1, db.GetNextSponsorSlideDisplayOrder())

	sponsorSlide := SponsorSlide{0, "Subtitle", "Line 1", "Line 2", "", 10, 1}
	assert.Nil(t, db.CreateSponsorSlide(&sponsorSlide))
	sponsorSlide2, err := db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Equal(t, sponsorSlide, *sponsorSlide2)
	assert.Equal(t, 2, db.GetNextSponsorSlideDisplayOrder())

	sponsorSlide.Line1 = "Blorpy"
	db.UpdateSponsorSlide(&sponsorSlide)
	sponsorSlide2, err = db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Equal(t, sponsorSlide.Line1, sponsorSlide2.Line1)

	db.DeleteSponsorSlide(sponsorSlide.Id)
	sponsorSlide2, err = db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide2)
}

func TestTruncateSponsorSlides(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	sponsorSlide := SponsorSlide{0, "Subtitle", "Line 1", "Line 2", "", 10, 0}
	db.CreateSponsorSlide(&sponsorSlide)
	db.TruncateSponsorSlides()
	sponsorSlide2, err := db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide2)
	assert.Equal(t, 1, db.GetNextSponsorSlideDisplayOrder())
}
