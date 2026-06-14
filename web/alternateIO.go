// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the field monitor display showing robot connection status.

package web

import (
	//"github.com/Team254/cheesy-arena/game"
	//"github.com/Team254/cheesy-arena/model"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"io"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
)

// RequestPayload represents the structure of the incoming POST data.
type RequestPayload struct {
	Channel int  `json:"channel"`
	State   bool `json:"state"`
}
type RequestPayloadPLCRegister struct {
	Register int  `json:"register"`
	CValue   uint16  `json:"cvalue"`
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
func (web *Web) setPLCRegister(w http.ResponseWriter, r *http.Request) {
	// Ensure the request is a POST request.
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body.
	var payload []RequestPayloadPLCRegister
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
    	http.Error(w, "Invalid request payload", http.StatusBadRequest)
    	return
	}

	for _, item := range payload {
    	web.arena.Plc.SetRegisterValue(item.Register, item.CValue)
	}

	// Respond with success.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("eStop state updated successfully."))

}

func (web *Web) xplcWebsocketHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := websocket.NewWebsocket(w, r)
    if err != nil {
        log.Printf("Websocket upgrade error: %v", err)
        return
    }
    defer ws.Close()

    // Subscribe to the PLC I/O + registers notifier
    arena := web.arena // assuming you have access to the arena/field
    plc := arena.Plc   // adjust based on actual field/arena struct

    if plc == nil || web.arena.Plc.IoChangeNotifier() == nil {
        ws.WriteError("PLC not configured")
        return
    }

    // Handle the notifier (sends initial state + live updates)
    ws.HandleNotifiers(web.arena.Plc.IoChangeNotifier())
}

// Main WebSocket handler
func (web *Web) plcWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// Subscribe to live PLC updates (registers, inputs, coils, etc.)
	if web.arena.Plc != nil {
		if notifier := web.arena.Plc.IoChangeNotifier(); notifier != nil {
			go ws.HandleNotifiers(notifier, web.arena.LedChangeNotifier)
		}
	}

	// Handle incoming messages (set registers)
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err != io.EOF {
				log.Printf("PLC WS read error: %v", err)
			}
			return
		}

		web.handlePLCWebSocketMessage(ws, messageType, data)
	}
}

func (web *Web) handlePLCWebSocketMessage(ws *websocket.Websocket, messageType string, data any) {
    switch messageType {

    case "setPLCRegister", "setRegisters":
        if messageType != "setPLCRegister" && messageType != "setRegisters" {
		return
		}

		var payloads []RequestPayloadPLCRegister

		switch v := data.(type) {
		case []any: // Array of payloads
			for _, item := range v {
				if p, ok := web.parsePLCRegisterPayload(item); ok {
					payloads = append(payloads, p)
				}
			}
		default: // Single payload
			if p, ok := web.parsePLCRegisterPayload(data); ok {
				payloads = append(payloads, p)
			}
		}

		if len(payloads) == 0 {
			ws.WriteError("Invalid or empty payload")
			return
		}

		// Apply changes exactly like your HTTP handler
		for _, item := range payloads {
			web.arena.Plc.SetRegisterValue(item.Register, uint16(item.CValue))  // Convert int16 → uint16
			//log.Printf("WebSocket: Set PLC register %d to %d", item.Register, item.CValue)
		}

		// Send success back to client
		ws.Write("plcRegisterSetSuccess", map[string]any{
			"count":   len(payloads),
			"success": true,
		})

    case "setInput":
        var payloads []RequestPayload

        switch v := data.(type) {
        case []any:
            for _, item := range v {
                if m, ok := item.(map[string]any); ok {
                    ch, _ := m["channel"].(float64)
                    st, _ := m["state"].(bool)
                    payloads = append(payloads, RequestPayload{Channel: int(ch), State: st})
                }
            }
        default:
            if m, ok := data.(map[string]any); ok {
                ch, _ := m["channel"].(float64)
                st, _ := m["state"].(bool)
                payloads = append(payloads, RequestPayload{Channel: int(ch), State: st})
            }
        }

        if len(payloads) == 0 {
            ws.WriteError("Invalid or empty input payload")
            return
        }

        for _, item := range payloads {
            web.arena.Plc.SetAlternateIOStopState(item.Channel, item.State)
        }

        ws.Write("plcInputSetSuccess", map[string]any{
            "count":   len(payloads),
            "success": true,
        })
    }
}
	

// Safe parser
func (web *Web) parsePLCRegisterPayload(data any) (RequestPayloadPLCRegister, bool) {
	m, ok := data.(map[string]any)
	if !ok {
		return RequestPayloadPLCRegister{}, false
	}

	regFloat, _ := m["register"].(float64)
	valFloat, _ := m["cValue"].(float64)   // JSON numbers come as float64

	return RequestPayloadPLCRegister{
		Register: int(regFloat),
		CValue:   uint16(valFloat),
	}, true
}