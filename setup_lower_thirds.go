// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for managing lower thirds.

package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"html/template"
	"io"
	"log"
	"net/http"
)

// Shows the lower third configuration page.
func LowerThirdsGetHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/setup_lower_thirds.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	lowerThirds, err := db.GetAllLowerThirds()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		LowerThirds []LowerThird
	}{eventSettings, lowerThirds}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the lower thirds client to send control commands.
func LowerThirdsWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}

		switch messageType {
		case "saveLowerThird":
			var lowerThird LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			saveLowerThird(&lowerThird)
		case "deleteLowerThird":
			var lowerThird LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = db.DeleteLowerThird(&lowerThird)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "showLowerThird":
			var lowerThird LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			saveLowerThird(&lowerThird)
			mainArena.lowerThirdNotifier.Notify(lowerThird)
			mainArena.audienceDisplayScreen = "lowerThird"
			mainArena.audienceDisplayNotifier.Notify(nil)
			continue
		case "hideLowerThird":
			var lowerThird LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			saveLowerThird(&lowerThird)
			mainArena.audienceDisplayScreen = "blank"
			mainArena.audienceDisplayNotifier.Notify(nil)
			continue
		case "reorderLowerThird":
			args := struct {
				Id     int
				MoveUp bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = reorderLowerThird(args.Id, args.MoveUp)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Force a reload of the client to render the updated lower thirds list.
		err = websocket.Write("reload", nil)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

func saveLowerThird(lowerThird *LowerThird) error {
	oldLowerThird, err := db.GetLowerThirdById(lowerThird.Id)
	if err != nil {
		return err
	}

	// Create or update lower third.
	if oldLowerThird == nil {
		err = db.CreateLowerThird(lowerThird)
	} else {
		err = db.SaveLowerThird(lowerThird)
	}
	if err != nil {
		return err
	}
	return nil
}

func reorderLowerThird(id int, moveUp bool) error {
	lowerThird, err := db.GetLowerThirdById(id)
	if err != nil {
		return err
	}

	// Get the lower third to swap positions with.
	lowerThirds, err := db.GetAllLowerThirds()
	if err != nil {
		return err
	}
	var lowerThirdIndex int
	for i, third := range lowerThirds {
		if third.Id == lowerThird.Id {
			lowerThirdIndex = i
			break
		}
	}
	if moveUp {
		lowerThirdIndex--
	} else {
		lowerThirdIndex++
	}
	if lowerThirdIndex < 0 || lowerThirdIndex == len(lowerThirds) {
		// The one to move is already at the limit; return an error to prevent a page reload.
		return fmt.Errorf("Already at the limit.")
	}
	adjacentLowerThird, err := db.GetLowerThirdById(lowerThirds[lowerThirdIndex].Id)
	if err != nil {
		return err
	}

	// Swap their display orders and save.
	lowerThird.DisplayOrder, adjacentLowerThird.DisplayOrder =
		adjacentLowerThird.DisplayOrder, lowerThird.DisplayOrder
	err = db.SaveLowerThird(lowerThird)
	if err != nil {
		return err
	}
	err = db.SaveLowerThird(adjacentLowerThird)
	if err != nil {
		return err
	}

	return nil
}
