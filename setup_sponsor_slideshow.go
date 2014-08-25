// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Web routes for managing sponsor slideshow.

package main

import (
	"html/template"
	"net/http"
	"strconv"
)

// Shows the lower third configuration page.
func SponsorSlideshowGetHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/sponsor_slideshow.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	sponsorSlideshow, err := db.GetAllSponsorSlideshows()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		SponsorSlideshow []SponsorSlideshow
	}{eventSettings, sponsorSlideshow}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the new or modified lower third to the database and triggers showing it on the audience display.
func SponsorSlideshowPostHandler(w http.ResponseWriter, r *http.Request) {
	sponsorSlideshowId, _ := strconv.Atoi(r.PostFormValue("id"))
	sponsorSlideshow, err := db.GetSponsorSlideshowById(sponsorSlideshowId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if r.PostFormValue("action") == "delete" {
		err := db.DeleteSponsorSlideshow(sponsorSlideshow)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		if sponsorSlideshow == nil {
			sponsorSlideshow = &SponsorSlideshow{Subtitle: r.PostFormValue("subtitle"),
												 Line1: r.PostFormValue("line1"),
												 Line2: r.PostFormValue("line2"),
												 Image: r.PostFormValue("image"),
												 Priority: r.PostFormValue("priority")}
			err = db.CreateSponsorSlideshow(sponsorSlideshow)
		} else {
			sponsorSlideshow.Subtitle = r.PostFormValue("subtitle")
			sponsorSlideshow.Line1 = r.PostFormValue("line1")
			sponsorSlideshow.Line2 = r.PostFormValue("line2")
			sponsorSlideshow.Image = r.PostFormValue("image")
			sponsorSlideshow.Priority = r.PostFormValue("priority")
			err = db.SaveSponsorSlideshow(sponsorSlideshow)
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/sponsor_slideshow", 302)
}
