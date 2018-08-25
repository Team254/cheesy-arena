// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for the server side of handling websockets.

package web

import (
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"sync"
)

// Wraps the Gorilla Websocket module so that we can define additional functions on it.
type Websocket struct {
	conn       *websocket.Conn
	writeMutex *sync.Mutex
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var websocketUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 2014}

// Upgrades the given HTTP request to a websocket connection.
func NewWebsocket(w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Websocket{conn, new(sync.Mutex)}, nil
}

func (ws *Websocket) Close() {
	ws.conn.Close()
}

func (ws *Websocket) Read() (string, interface{}, error) {
	var message WebsocketMessage
	err := ws.conn.ReadJSON(&message)
	if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) {
		// This error indicates that the browser terminated the connection normally; rewwrite it so that clients don't
		// log it.
		return "", nil, io.EOF
	}
	return message.Type, message.Data, err
}

func (ws *Websocket) Write(messageType string, data interface{}) error {
	ws.writeMutex.Lock()
	defer ws.writeMutex.Unlock()
	return ws.conn.WriteJSON(WebsocketMessage{messageType, data})
}

func (ws *Websocket) WriteError(errorMessage string) error {
	ws.writeMutex.Lock()
	defer ws.writeMutex.Unlock()
	return ws.conn.WriteJSON(WebsocketMessage{"error", errorMessage})
}

func (ws *Websocket) ShowDialog(message string) error {
	ws.writeMutex.Lock()
	defer ws.writeMutex.Unlock()
	return ws.conn.WriteJSON(WebsocketMessage{"dialog", message})
}
