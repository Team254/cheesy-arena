// Author: capplefamily@gmail.com (Corey Applegate)
//
// Web handlers for the field monitor display help file.

package web

import (
	"fmt"
	"net/http"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/Team254/cheesy-arena/model"
)

// Renders the Estop Control display.
func (web *Web) estopContolDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	alliance := r.PathValue("alliance")
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/alliance_station_estop_control.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the Estop Control client to receive status updates.
func (web *Web) estopContolDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(
		web.arena.PlcCoilsNotifier,  // Notifier for PLC coil state changes with Descriptions
		web.arena.Plc.IoChangeNotifier(), // Notifier for PLC IO changes includs All PLC coils Inputs and Fegisters
	)
}

