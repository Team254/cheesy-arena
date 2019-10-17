// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Model and datastore CRUD methods for the sponsor slideshow.

package model

type SponsorSlide struct {
	Id             int
	Subtitle       string
	Line1          string
	Line2          string
	Image          string
	DisplayTimeSec int
	DisplayOrder   int
}

func (database *Database) CreateSponsorSlide(sponsorSlide *SponsorSlide) error {
	return database.sponsorSlideMap.Insert(sponsorSlide)
}

func (database *Database) GetSponsorSlideById(id int) (*SponsorSlide, error) {
	sponsorSlide := new(SponsorSlide)
	err := database.sponsorSlideMap.Get(sponsorSlide, id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		sponsorSlide = nil
		err = nil
	}
	return sponsorSlide, err
}

func (database *Database) SaveSponsorSlide(sponsorSlide *SponsorSlide) error {
	_, err := database.sponsorSlideMap.Update(sponsorSlide)
	return err
}

func (database *Database) DeleteSponsorSlide(sponsorSlide *SponsorSlide) error {
	_, err := database.sponsorSlideMap.Delete(sponsorSlide)
	return err
}

func (database *Database) TruncateSponsorSlides() error {
	return database.sponsorSlideMap.TruncateTables()
}

func (database *Database) GetAllSponsorSlides() ([]SponsorSlide, error) {
	var sponsorSlides []SponsorSlide
	err := database.sponsorSlideMap.Select(&sponsorSlides, "SELECT * FROM sponsor_slides ORDER BY displayorder")
	return sponsorSlides, err
}

func (database *Database) GetNextSponsorSlideDisplayOrder() int {
	var count int
	_ = database.sponsorSlideMap.SelectOne(&count, "SELECT MAX(displayorder) + 1 FROM sponsor_slides")
	return count
}
