// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
//
// Web routes for managing sponsor slides.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"net/http"
	"strconv"
)

// Shows the sponsor slides configuration page.
func (web *Web) sponsorSlidesGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_sponsor_slides.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	sponsorSlides, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Append a blank slide to the end that can be used to add a new one.
	sponsorSlides = append(sponsorSlides, model.SponsorSlide{DisplayTimeSec: 10})

	data := struct {
		*model.EventSettings
		SponsorSlides []model.SponsorSlide
	}{web.arena.EventSettings, sponsorSlides}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Saves the new or modified sponsor slides to the database.
func (web *Web) sponsorSlidesPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	sponsorSlideId, _ := strconv.Atoi(r.PostFormValue("id"))
	sponsorSlide, err := web.arena.Database.GetSponsorSlideById(sponsorSlideId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	switch r.PostFormValue("action") {
	case "delete":
		err := web.arena.Database.DeleteSponsorSlide(sponsorSlide.Id)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	case "save":
		displayTimeSec, _ := strconv.Atoi(r.PostFormValue("displayTimeSec"))
		if sponsorSlide == nil {
			sponsorSlide = &model.SponsorSlide{Subtitle: r.PostFormValue("subtitle"),
				Line1: r.PostFormValue("line1"), Line2: r.PostFormValue("line2"),
				Image: r.PostFormValue("image"), DisplayTimeSec: displayTimeSec,
				DisplayOrder: web.arena.Database.GetNextSponsorSlideDisplayOrder(),
			}
			err = web.arena.Database.CreateSponsorSlide(sponsorSlide)
		} else {
			sponsorSlide.Subtitle = r.PostFormValue("subtitle")
			sponsorSlide.Line1 = r.PostFormValue("line1")
			sponsorSlide.Line2 = r.PostFormValue("line2")
			sponsorSlide.Image = r.PostFormValue("image")
			sponsorSlide.DisplayTimeSec = displayTimeSec
			err = web.arena.Database.UpdateSponsorSlide(sponsorSlide)
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}
	case "reorderUp":
		if err = web.reorderSponsorSlide(sponsorSlideId, true); err != nil {
			handleWebErr(w, err)
			return
		}
	case "reorderDown":
		if err = web.reorderSponsorSlide(sponsorSlideId, false); err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/sponsor_slides", 303)
}

// Swaps the sponsor slide having the given ID with the one immediately above or below it.
func (web *Web) reorderSponsorSlide(id int, moveUp bool) error {
	sponsorSlide, err := web.arena.Database.GetSponsorSlideById(id)
	if err != nil {
		return err
	}

	// Get the sponsor slide to swap positions with.
	sponsorSlides, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		return err
	}
	var sponsorSlideIndex int
	for i, slide := range sponsorSlides {
		if slide.Id == sponsorSlide.Id {
			sponsorSlideIndex = i
			break
		}
	}
	if moveUp {
		sponsorSlideIndex--
	} else {
		sponsorSlideIndex++
	}
	if sponsorSlideIndex < 0 || sponsorSlideIndex == len(sponsorSlides) {
		// The one to move is already at the limit; do nothing.
		return nil
	}
	adjacentSponsorSlide := &sponsorSlides[sponsorSlideIndex]
	if err != nil {
		return err
	}

	// Swap their display orders and save.
	sponsorSlide.DisplayOrder, adjacentSponsorSlide.DisplayOrder =
		adjacentSponsorSlide.DisplayOrder, sponsorSlide.DisplayOrder
	err = web.arena.Database.UpdateSponsorSlide(sponsorSlide)
	if err != nil {
		return err
	}
	err = web.arena.Database.UpdateSponsorSlide(adjacentSponsorSlide)
	if err != nil {
		return err
	}

	return nil
}
