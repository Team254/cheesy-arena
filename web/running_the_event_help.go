// Author: justfishin@icloud.com (Justin Fischer)
//
// Web handler for the alliance station display help page.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/model"
)

// Renders the alliance station display help page.
func (web *Web) runningTheEventHelpHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := web.parseFiles(
		"templates/running_the_event_help.html",
		"templates/base.html",
	)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{
		web.arena.EventSettings,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		handleWebErr(w, err)
	}
}
