// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNonexistentSponsorSlide(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	sponsorSlide, err := db.GetSponsorSlideById(1114)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide)
}

func TestSponsorSlideCrud(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	sponsorSlide := SponsorSlide{0, "Subtitle", "Line 1", "Line 2", "", 10}
	db.CreateSponsorSlide(&sponsorSlide)
	sponsorSlide2, err := db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Equal(t, sponsorSlide, *sponsorSlide2)

	sponsorSlide.Line1 = "Blorpy"
	db.SaveSponsorSlide(&sponsorSlide)
	sponsorSlide2, err = db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Equal(t, sponsorSlide.Line1, sponsorSlide2.Line1)

	db.DeleteSponsorSlide(&sponsorSlide)
	sponsorSlide2, err = db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide2)
}

func TestTruncateSponsorSlides(t *testing.T) {
	clearDb()
	defer clearDb()
	db, err := OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()

	sponsorSlide := SponsorSlide{0, "Subtitle", "Line 1", "Line 2", "", 10}
	db.CreateSponsorSlide(&sponsorSlide)
	db.TruncateSponsorSlides()
	sponsorSlide2, err := db.GetSponsorSlideById(1)
	assert.Nil(t, err)
	assert.Nil(t, sponsorSlide2)
}
