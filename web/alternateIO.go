// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the field monitor display showing robot connection status.

package web

import (
	//"github.com/Team254/cheesy-arena/game"
	//"github.com/Team254/cheesy-arena/model"
	"encoding/json"
	"net/http"
)

// RequestPayload represents the structure of the incoming POST data.
type RequestPayload struct {
	Channel int  `json:"channel"`
	State   bool `json:"state"`
}

// Renders the field monitor display.
func (web *Web) eStopStatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request is a POST request.
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body.
	var payload []RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
    	http.Error(w, "Invalid request payload", http.StatusBadRequest)
    	return
	}

	for _, item := range payload {
    	web.arena.Plc.SetAlternateIOStopState(item.Channel, item.State)
	}

	// Respond with success.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("eStop state updated successfully."))

}

func (web *Web) getAllPlcCoilsGetHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request is a GET request.
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the current state of all PLC coils.
    coilsArray := web.arena.Plc.GetAllCoils()
    coilsArrayNames := web.arena.Plc.GetCoilNames()

	// Build a map pairing coil names with their values.
    coilsMap := make(map[string]bool)
    for i, name := range coilsArrayNames {
        if i < len(coilsArray) {
            coilsMap[name] = coilsArray[i]
        }
    }
	
	// Marshal the response payload.
	response, err := json.Marshal(coilsMap)
	if err != nil {
		http.Error(w, "Failed to marshal PLC Coils state", http.StatusInternalServerError)
		return
	}

	// Send the response.
	w.Write(response)
}
