// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for the server side of handling websockets.

package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const pingInterval = time.Second * 10

// Wraps the Gorilla Websocket module so that we can define additional functions on it.
type Websocket struct {
	conn       *websocket.Conn
	writeMutex *sync.Mutex
}

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
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

func NewTestWebsocket(conn *websocket.Conn) *Websocket {
	return &Websocket{conn, new(sync.Mutex)}
}

func (ws *Websocket) Close() error {
	return ws.conn.Close()
}

func (ws *Websocket) Read() (string, any, error) {
	var message Message
	err := ws.conn.ReadJSON(&message)
	if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived) {
		// This error indicates that the browser terminated the connection normally; rewrite it so that clients don't
		// log it.
		return "", nil, io.EOF
	}
	if err != nil {
		// Include the caller of this method in the error message.
		_, file, line, _ := runtime.Caller(1)
		filePathParts := strings.Split(file, "/")
		return "", nil, fmt.Errorf("[%s:%d] Websocket read error: %v", filePathParts[len(filePathParts)-1], line, err)
	}
	return message.Type, message.Data, nil
}

func (ws *Websocket) ReadWithTimeout(timeout time.Duration) (string, any, error) {
	type wsReadResult struct {
		messageType string
		message     any
		err         error
	}
	readChan := make(chan wsReadResult, 1)
	go func() {
		messageType, message, err := ws.Read()
		readChan <- wsReadResult{messageType, message, err}
	}()

	select {
	case result := <-readChan:
		return result.messageType, result.message, result.err
	case <-time.After(timeout):
		return "", nil, fmt.Errorf("Websocket read timed out after waiting for %v", timeout)
	}
}

func (ws *Websocket) Write(messageType string, data any) error {
	ws.writeMutex.Lock()
	defer ws.writeMutex.Unlock()
	err := ws.conn.WriteJSON(Message{messageType, data})
	if err != nil {
		// Include the caller of this method in the error message.
		_, file, line, _ := runtime.Caller(1)
		filePathParts := strings.Split(file, "/")
		return fmt.Errorf("[%s:%d] Websocket write error: %v", filePathParts[len(filePathParts)-1], line, err)
	}
	return nil
}

func (ws *Websocket) WriteNotifier(notifier *Notifier) error {
	return ws.Write(notifier.messageType, notifier.getMessageBody())
}

func (ws *Websocket) WriteError(errorMessage string) error {
	return ws.Write("error", errorMessage)
}

// Creates listeners for the given notifiers and loops forever to pass their output directly through to the websocket.
func (ws *Websocket) HandleNotifiers(notifiers ...*Notifier) {
	// Use reflection to dynamically build a select/case structure for all the notifiers.
	listeners := make([]reflect.SelectCase, len(notifiers))
	for i, notifier := range notifiers {
		listener := notifier.listen()
		defer close(listener)
		listeners[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(listener)}

		// Send each notifier's respective data immediately upon connection to bootstrap the client state.
		if notifier.messageProducer != nil {
			err := ws.WriteNotifier(notifier)
			if err != nil {
				log.Printf("Websocket error writing inital value for notifier %v: %v", notifier, err)
				return
			}
		}
	}

	// Add an additional case to periodically ping the websocket to detect whether the client has closed it.
	pingCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(time.Tick(pingInterval))}
	pingIndex := len(listeners)
	listeners = append(listeners, pingCase)

	for {
		// Block until a message is available on any of the channels.
		chosenIndex, value, ok := reflect.Select(listeners)
		if ok && chosenIndex == pingIndex {
			err := ws.Write("ping", nil)
			if err != nil {
				// The client has probably closed the connection; bail out of the loop.
				return
			}
			continue
		}
		if !ok {
			log.Printf("Channel for notifier %v closed unexpectedly.", notifiers[chosenIndex])
			return
		}
		message, ok := value.Interface().(messageEnvelope)
		if !ok {
			log.Printf("Channel for notifier %v sent unexpected value %v.", notifiers[chosenIndex], value)
			continue
		}

		// Forward the message verbatim on to the websocket.
		err := ws.Write(message.messageType, message.messageBody)
		if err != nil {
			// The client has probably closed the connection; bail out of the loop.
			return
		}
	}
}
