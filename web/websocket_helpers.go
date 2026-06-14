// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/websocket"
	"log"
)

func writeWebsocketError(ws *websocket.Websocket, errorMessage string) {
	if err := ws.WriteError(errorMessage); err != nil {
		log.Println(err)
	}
}

func writeWebsocketMessage(ws *websocket.Websocket, messageType string, data any) {
	if err := ws.Write(messageType, data); err != nil {
		log.Println(err)
	}
}

func closeWebsocket(ws *websocket.Websocket) {
	if err := ws.Close(); err != nil {
		log.Println(err)
	}
}
