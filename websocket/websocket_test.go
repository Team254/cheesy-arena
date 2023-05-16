// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebsocket(t *testing.T) {
	// Set up some fake notifiers.
	notifier1 := NewNotifier("messageType1", func() any { return "test message" })
	notifier2 := NewNotifier("messageType2", nil)
	changingValue := 123.45
	notifier3 := NewNotifier("messageType3", func() any { return changingValue })

	// Start up a fake server with a trivial websocket handler.
	testWebsocketHandler := func(w http.ResponseWriter, r *http.Request) {
		ws, err := NewWebsocket(w, r)
		assert.Nil(t, err)
		defer ws.Close()

		// Subscribe the websocket to the notifiers whose messages will be passed on, in a separate goroutine.
		go ws.HandleNotifiers(notifier3, notifier2, notifier1)

		// Loop, waiting for commands and responding to them, until the client closes the connection.
		for {
			messageType, data, err := ws.Read()
			if err != nil {
				if err == io.EOF {
					// Client has closed the connection; nothing to do here.
					return
				} else {
					assert.Fail(t, err.Error())
					return
				}
			}

			switch messageType {
			case "sendMessageType1":
				ws.WriteNotifier(notifier1)
			case "sendError":
				ws.WriteError("error message")
			default:
				// Echo the commands back out.
				err = ws.Write(messageType, data)
				assert.Nil(t, err)
			}
		}
	}
	handler := http.NewServeMux()
	handler.HandleFunc("/", testWebsocketHandler)
	server := httptest.NewServer(handler)
	defer server.Close()
	wsUrl := "ws" + server.URL[len("http"):]

	// Create a client connection to the websocket handler on the server.
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	assert.Nil(t, err)
	ws := NewTestWebsocket(conn)

	// Ensure the initial messages are sent upon connection.
	assertMessage(t, ws, "messageType3", changingValue)
	assertMessage(t, ws, "messageType1", "test message")

	// Trigger and read notifications.
	notifier2.Notify()
	assertMessage(t, ws, "messageType2", nil)
	notifier1.Notify()
	assertMessage(t, ws, "messageType1", "test message")
	notifier3.Notify()
	assertMessage(t, ws, "messageType3", changingValue)
	changingValue = 254.254
	notifier3.Notify()
	assertMessage(t, ws, "messageType3", changingValue)
	notifier1.NotifyWithMessage("test message 2")
	assertMessage(t, ws, "messageType1", "test message 2")
	notifier3.NotifyWithMessage("test message 3")
	assertMessage(t, ws, "messageType3", "test message 3")

	// Test sending commands back.
	ws.Write("sendMessageType1", nil)
	assertMessage(t, ws, "messageType1", "test message")
	ws.Write("messageType4", "test message 4")
	assertMessage(t, ws, "messageType4", "test message 4")
	ws.Write("sendError", nil)
	assertMessage(t, ws, "error", "error message")

	// Ensure the read times out if there is nothing to read.
	_, _, err = ws.ReadWithTimeout(time.Millisecond)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "timed out")
	}

	// Test that closing the connection eliminates the listeners once another message is sent.
	assert.Nil(t, ws.Close())
	time.Sleep(time.Millisecond)
	notifier1.Notify()
	time.Sleep(time.Millisecond)
	notifier1.Notify()
	assert.Equal(t, 0, len(notifier1.listeners))
}

func assertMessage(t *testing.T, ws *Websocket, expectedMessageType string, expectedMessageBody any) {
	messageType, messageBody, err := ws.ReadWithTimeout(time.Second)
	if assert.Nil(t, err) {
		assert.Equal(t, expectedMessageType, messageType)
		assert.Equal(t, expectedMessageBody, messageBody)
	}
}
