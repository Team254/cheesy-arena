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
	"strconv"
	"github.com/Team254/cheesy-arena/field"
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

// Handles the request to start the match.
func (web *Web) startMatchPostHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request is a POST request.
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Start the match.
	web.arena.StartMatch()

	// Respond with success.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Field stack light state updated successfully."))
}

type fieldStackLight struct {
	Red    bool `json:"redStackLight"`
	Blue   bool `json:"blueStackLight"`
	Orange bool `json:"orangeStackLight"`
	Green  bool `json:"greenStackLight"`
}

func (web *Web) fieldStackLightGetHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request is a GET request.
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the current state of the field stack light.
	var stackLight fieldStackLight
	stackLight.Red, stackLight.Blue, stackLight.Orange, stackLight.Green = web.arena.Plc.GetFieldStackLight()

	// Marshal the response payload.
	response, err := json.Marshal(stackLight)
	if err != nil {
		http.Error(w, "Failed to marshal eStop state", http.StatusInternalServerError)
		return
	}

	// Send the response.
	w.Write(response)
}


type lightState struct {
	Color string `json:"color"`
	Blink bool `json:"blink"`
}
// Structure representing one light fixture
// Each lightState represents one light in the stack.
type teamStackLight struct {
	LightStates [2]lightState `json:"lightStates"`
}
// Structure that represents all of the team stack lights
type allStackLights struct {
	Red [3]teamStackLight `json:"red"`
	Blue [3]teamStackLight `json:"blue"`
}
func (web *Web) teamStackLightGetHandler(w http.ResponseWriter, r *http.Request) {
		// Ensure the request is a GET request.
		// See the team_sign.go method: generateTeamNumberTexts for the template
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var stackLights allStackLights
	// stackLights.Red = [3]teamStackLight
	// stackLights.Blue = [3]teamStackLight

	for team, allianceStation := range web.arena.AllianceStations { 
		var teamStackLights = &stackLights.Blue
		var allianceColor = "blue"
		if team[0] == 'R'	{
			teamStackLights = &stackLights.Red
			allianceColor = "red"
		}
		dsN,_ := strconv.Atoi(string(team[1]))
		teamStackLight := &teamStackLights[dsN-1]
		//  The lights are as follows:
		//  L2: Blue/Red
		//     off: Connection est. to robot
		//     solid: Robot enabled
		//     flash: no connection to robot or bypassed
		//  L1: Amber
		//     off: Estop not pressed/disabled
		//     solid: Estop pressed/enabled
		//     flash: Astop pressed/enabled during autonomous period


		// Light/Layer 1 - Stop States
		if allianceStation.EStop {
			teamStackLight.LightStates[0] = lightState{Color: "orangered", Blink: false}
		} else if allianceStation.AStop && web.arena.MatchState == field.AutoPeriod {
			teamStackLight.LightStates[0] = lightState{Color: "orangered", Blink: true}
		} else {
			teamStackLight.LightStates[0] = lightState{Color: "black", Blink: false}
		}

		// Light/Layer 2 - Robot States
		// Blink with any problem 
		// Solid during the match if all is good.
		// Off off-match if all is good.
		var ok = true;
		if allianceStation.Bypass {
			ok = false
			// This is always false for some reason
		// } else if !allianceStation.Ethernet {
		// 	ok = false
		} else if allianceStation.DsConn == nil {
			ok = false
		} else if allianceStation.DsConn.WrongStation != "" {
			ok = false
		} else if !allianceStation.DsConn.RadioLinked {
			ok = false
		} else if !allianceStation.DsConn.RioLinked {
			ok = false
		} else if !allianceStation.DsConn.RobotLinked {
			ok = false
		}

		if ok { 
			if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod || web.arena.MatchState == field.TeleopPeriod {
				// Robot enabled during match
				teamStackLight.LightStates[1] = lightState{Color: allianceColor, Blink: false}
			} else {
				// Robot connected outside of the match
				teamStackLight.LightStates[1] = lightState{Color: "black", Blink: false}
			}
		} else {
			teamStackLight.LightStates[1] = lightState{Color: allianceColor, Blink: true}		
		}
	}

	// Marshal the response payload.
	response, err := json.Marshal(stackLights)
	if err != nil {
		http.Error(w, "Failed to marshal team stacklights state", http.StatusInternalServerError)
		return
	}

	// Send the response.
	w.Write(response)
}