package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"net/http"
	"strconv"
)

// Prevent MIME sniffing in browsers.
func addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func (web *Web) awardsGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	addSecurityHeaders(w)

	template, err := web.parseFiles("templates/setup_awards.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	awards, err := web.arena.Database.GetAllAwards()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	awards = append(awards, model.Award{})

	data := struct {
		*model.EventSettings
		Awards []model.Award
		Teams  []model.Team
	}{web.arena.EventSettings, awards, teams}
	if err := template.ExecuteTemplate(w, "base", data); err != nil {
		handleWebErr(w, err)
		return
	}
}

func (web *Web) awardsPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	addSecurityHeaders(w)

	awardId, _ := strconv.Atoi(r.PostFormValue("id"))
	if r.PostFormValue("action") == "delete" {
		if err := tournament.DeleteAward(web.arena.Database, awardId); err != nil {
			handleWebErr(w, err)
			return
		}
	} else {
		teamId, _ := strconv.Atoi(r.PostFormValue("teamId"))
		award := model.Award{
			Id:         awardId,
			Type:       model.JudgedAward,
			AwardName:  r.PostFormValue("awardName"),
			TeamId:     teamId,
			PersonName: r.PostFormValue("personName"),
		}
		if err := tournament.CreateOrUpdateAward(web.arena.Database, &award, true); err != nil {
			handleWebErr(w, err)
			return
		}
	}

	http.Redirect(w, r, "/setup/awards", http.StatusSeeOther)
}
