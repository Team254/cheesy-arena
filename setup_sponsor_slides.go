// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Web routes for managing sponsor slides.

package main

import (
	"html/template"
	"net/http"
	"strconv"
)

// Shows the sponsor slides configuration page.
func SponsorSlidesGetHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/sponsor_slides.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	sponsorSlides, err := db.GetAllSponsorSlides()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		SponsorSlides []SponsorSlide
	}{eventSettings, sponsorSlides}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the new or modified sponsor slides to the database.
func SponsorSlidesPostHandler(w http.ResponseWriter, r *http.Request) {
	sponsorSlideId, _ := strconv.Atoi(r.PostFormValue("id"))
	sponsorSlide, err := db.GetSponsorSlideById(sponsorSlideId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if r.PostFormValue("action") == "delete" {
		err := db.DeleteSponsorSlide(sponsorSlide)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		displayTimeSec, _ := strconv.Atoi(r.PostFormValue("displayTimeSec"))
		if sponsorSlide == nil {
			sponsorSlide = &SponsorSlide{Subtitle: r.PostFormValue("subtitle"),
				Line1: r.PostFormValue("line1"), Line2: r.PostFormValue("line2"),
				Image: r.PostFormValue("image"), DisplayTimeSec: displayTimeSec}
			err = db.CreateSponsorSlide(sponsorSlide)
		} else {
			sponsorSlide.Subtitle = r.PostFormValue("subtitle")
			sponsorSlide.Line1 = r.PostFormValue("line1")
			sponsorSlide.Line2 = r.PostFormValue("line2")
			sponsorSlide.Image = r.PostFormValue("image")
			sponsorSlide.DisplayTimeSec = displayTimeSec
			err = db.SaveSponsorSlide(sponsorSlide)
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/sponsor_slides", 302)
}
