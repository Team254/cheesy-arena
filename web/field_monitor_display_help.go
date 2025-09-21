// Author: capplefamily@gmail.com (Corey Applegate)
//
// Web handlers for the field monitor display help file.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
)

// Renders the field monitor display.
func (web *Web) fieldMonitorDisplayHelpHandler(w http.ResponseWriter, r *http.Request) {

	template, err := web.parseFiles("templates/field_monitor_display_help.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		MatchSounds []*game.MatchSound
	}{web.arena.EventSettings, game.MatchSounds}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}

}
