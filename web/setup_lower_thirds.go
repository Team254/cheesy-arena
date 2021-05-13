// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for managing lower thirds.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

// Shows the lower third configuration page.
func (web *Web) lowerThirdsGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_lower_thirds.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	lowerThirds, err := web.arena.Database.GetAllLowerThirds()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		LowerThirds []model.LowerThird
	}{web.arena.EventSettings, lowerThirds}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the lower thirds client to send control commands.
func (web *Web) lowerThirdsWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "saveLowerThird":
			var lowerThird model.LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			web.saveLowerThird(&lowerThird)
		case "deleteLowerThird":
			var lowerThird model.LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.Database.DeleteLowerThird(lowerThird.Id)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "showLowerThird":
			var lowerThird model.LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			web.saveLowerThird(&lowerThird)
			web.arena.LowerThird = &lowerThird
			web.arena.ShowLowerThird = true
			web.arena.LowerThirdNotifier.Notify()
			continue
		case "hideLowerThird":
			var lowerThird model.LowerThird
			err = mapstructure.Decode(data, &lowerThird)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			web.saveLowerThird(&lowerThird)
			web.arena.ShowLowerThird = false
			web.arena.LowerThirdNotifier.Notify()
			continue
		case "reorderLowerThird":
			args := struct {
				Id     int
				MoveUp bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.reorderLowerThird(args.Id, args.MoveUp)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Force a reload of the client to render the updated lower thirds list.
		err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (web *Web) saveLowerThird(lowerThird *model.LowerThird) error {
	oldLowerThird, err := web.arena.Database.GetLowerThirdById(lowerThird.Id)
	if err != nil {
		return err
	}

	// Create or update lower third.
	if oldLowerThird == nil {
		lowerThird.DisplayOrder = web.arena.Database.GetNextLowerThirdDisplayOrder()
		err = web.arena.Database.CreateLowerThird(lowerThird)
	} else {
		err = web.arena.Database.UpdateLowerThird(lowerThird)
	}
	if err != nil {
		return err
	}
	return nil
}

// Swaps the lower third having the given ID with the one immediately above or below it.
func (web *Web) reorderLowerThird(id int, moveUp bool) error {
	lowerThird, err := web.arena.Database.GetLowerThirdById(id)
	if err != nil {
		return err
	}

	// Get the lower third to swap positions with.
	lowerThirds, err := web.arena.Database.GetAllLowerThirds()
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
	adjacentLowerThird := &lowerThirds[lowerThirdIndex]
	if err != nil {
		return err
	}

	// Swap their display orders and save.
	lowerThird.DisplayOrder, adjacentLowerThird.DisplayOrder =
		adjacentLowerThird.DisplayOrder, lowerThird.DisplayOrder
	err = web.arena.Database.UpdateLowerThird(lowerThird)
	if err != nil {
		return err
	}
	err = web.arena.Database.UpdateLowerThird(adjacentLowerThird)
	if err != nil {
		return err
	}

	return nil
}
