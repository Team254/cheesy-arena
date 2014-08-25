// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Model and datastore CRUD methods for the sponsor slideshow.

package main

type SponsorSlideshow struct {
	Id 			int
	Subtitle    string
	Line1 		string
	Line2 		string
	Image 		string
	Priority	string
}

func (database *Database) CreateSponsorSlideshow(sponsorSlideshow *SponsorSlideshow) error {
	return database.sponsorSlideshowMap.Insert(sponsorSlideshow)
}

func (database *Database) GetSponsorSlideshowById(id int) (*SponsorSlideshow, error) {
	sponsorSlideshow := new(SponsorSlideshow)
	err := database.sponsorSlideshowMap.Get(sponsorSlideshow, id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		sponsorSlideshow = nil
		err = nil
	}
	return sponsorSlideshow, err
}

func (database *Database) SaveSponsorSlideshow(sponsorSlideshow *SponsorSlideshow) error {
	_, err := database.sponsorSlideshowMap.Update(sponsorSlideshow)
	return err
}

func (database *Database) DeleteSponsorSlideshow(sponsorSlideshow *SponsorSlideshow) error {
	_, err := database.sponsorSlideshowMap.Delete(sponsorSlideshow)
	return err
}

func (database *Database) TruncateSponsorSlideshows() error {
	return database.sponsorSlideshowMap.TruncateTables()
}

func (database *Database) GetAllSponsorSlideshows() ([]SponsorSlideshow, error) {
	var sponsorSlideshows []SponsorSlideshow
	err := database.teamMap.Select(&sponsorSlideshows, "SELECT * FROM sponsor_slideshow ORDER BY id")
	return sponsorSlideshows, err
}
