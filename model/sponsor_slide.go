// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Model and datastore CRUD methods for the sponsor slideshow.

package model

import "sort"

type SponsorSlide struct {
	Id             int `db:"id"`
	Subtitle       string
	Line1          string
	Line2          string
	Image          string
	DisplayTimeSec int
	DisplayOrder   int
}

func (database *Database) CreateSponsorSlide(sponsorSlide *SponsorSlide) error {
	return database.sponsorSlideTable.create(sponsorSlide)
}

func (database *Database) GetSponsorSlideById(id int) (*SponsorSlide, error) {
	return database.sponsorSlideTable.getById(id)
}

func (database *Database) UpdateSponsorSlide(sponsorSlide *SponsorSlide) error {
	return database.sponsorSlideTable.update(sponsorSlide)
}

func (database *Database) DeleteSponsorSlide(id int) error {
	return database.sponsorSlideTable.delete(id)
}

func (database *Database) TruncateSponsorSlides() error {
	return database.sponsorSlideTable.truncate()
}

func (database *Database) GetAllSponsorSlides() ([]SponsorSlide, error) {
	sponsorSlides, err := database.sponsorSlideTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(sponsorSlides, func(i, j int) bool {
		return sponsorSlides[i].DisplayOrder < sponsorSlides[j].DisplayOrder
	})
	return sponsorSlides, nil
}

func (database *Database) GetNextSponsorSlideDisplayOrder() int {
	sponsorSlides, err := database.GetAllSponsorSlides()
	if err != nil {
		return 0
	}
	if len(sponsorSlides) == 0 {
		return 1
	}
	return sponsorSlides[len(sponsorSlides)-1].DisplayOrder + 1
}
