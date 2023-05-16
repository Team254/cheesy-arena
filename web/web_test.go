// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIndex(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Home - Untitled Event - Cheesy Arena")
}

func (web *Web) getHttpResponse(path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	web.newHandler().ServeHTTP(recorder, req)
	return recorder
}

func (web *Web) getHttpResponseWithHeaders(path string, headers map[string]string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	web.newHandler().ServeHTTP(recorder, req)
	return recorder
}

func (web *Web) postHttpResponse(path string, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	web.newHandler().ServeHTTP(recorder, req)
	return recorder
}

// Starts a real local HTTP server that can be used by more sophisticated tests.
func (web *Web) startTestServer() (*httptest.Server, string) {
	server := httptest.NewServer(web.newHandler())
	return server, "ws" + server.URL[len("http"):]
}

// Receives the next websocket message and asserts that it is an error.
func readWebsocketError(t *testing.T, ws *websocket.Websocket) string {
	messageType, data, err := ws.Read()
	if assert.Nil(t, err) && assert.Equal(t, "error", messageType) {
		return data.(string)
	}
	return "error"
}

// Receives the next websocket message and asserts that it is of the given type.
func readWebsocketType(t *testing.T, ws *websocket.Websocket, expectedMessageType string) any {
	messageType, message, err := ws.ReadWithTimeout(time.Second)
	if assert.Nil(t, err) {
		assert.Equal(t, expectedMessageType, messageType)
	}
	return message
}

func readWebsocketMultiple(t *testing.T, ws *websocket.Websocket, count int) map[string]any {
	messages := make(map[string]any)
	for i := 0; i < count; i++ {
		messageType, message, err := ws.ReadWithTimeout(time.Second)
		if assert.Nil(t, err) {
			messages[messageType] = message
		}
	}
	return messages
}

func setupTestWeb(t *testing.T) *Web {
	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.PauseDurationSec = 2
	arena := field.SetupTestArena(t, "web")
	return NewWeb(arena)
}
