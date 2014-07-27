// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()

	recorder := getHttpResponse("/")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Home - Untitled Event - Cheesy Arena")
}

func getHttpResponse(path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	newHandler().ServeHTTP(recorder, req)
	return recorder
}

func postHttpResponse(path string, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	newHandler().ServeHTTP(recorder, req)
	return recorder
}

// Starts a real local HTTP server that can be used by more sophisticated tests.
func startTestServer() (*httptest.Server, string) {
	server := httptest.NewServer(newHandler())
	return server, "ws" + server.URL[len("http"):]
}

// Receives the next websocket message and asserts that it is an error.
func readWebsocketError(t *testing.T, ws *Websocket) string {
	messageType, data, err := ws.Read()
	if assert.Nil(t, err) && assert.Equal(t, "error", messageType) {
		return data.(string)
	}
	return "error"
}

// Receives the next websocket message and asserts that it is of the given type.
func readWebsocketType(t *testing.T, ws *Websocket, expectedMessageType string) interface{} {
	messageType, message, err := ws.Read()
	if assert.Nil(t, err) {
		assert.Equal(t, expectedMessageType, messageType)
	}
	return message
}
