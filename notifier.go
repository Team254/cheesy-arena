// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Publish-subscribe model for nonblocking notification of server events to websocket clients.

package main

import (
	"log"
)

// Allow the listeners to buffer a small number of notifications to streamline delivery.
const notifyBufferSize = 3

type Notifier struct {
	// The map is essentially a set; the value is ignored.
	listeners map[chan interface{}]struct{}
}

func NewNotifier() *Notifier {
	notifier := new(Notifier)
	notifier.listeners = make(map[chan interface{}]struct{})
	return notifier
}

// Registers and returns a channel that can be read from to receive notification messages. The caller is
// responsible for closing the channel, which will cause it to be reaped from the list of listeners.
func (notifier *Notifier) Listen() chan interface{} {
	listener := make(chan interface{}, notifyBufferSize)
	notifier.listeners[listener] = struct{}{}
	return listener
}

// Sends the given message to all registered listeners, and cleans up any listeners that have closed.
func (notifier *Notifier) Notify(message interface{}) {
	for listener, _ := range notifier.listeners {
		notifier.notifyListener(listener, message)
	}
}

func (notifier *Notifier) notifyListener(listener chan interface{}, message interface{}) {
	defer func() {
		// If channel is closed sending to it will cause a panic; recover and remove it from the list.
		if r := recover(); r != nil {
			delete(notifier.listeners, listener)
		}
	}()

	// Do a non-blocking send. This guarantees that sending notifications won't interrupt the main event loop,
	// at the risk of clients missing some messages if they don't read them all promptly.
	select {
	case listener <- message:
		// The notification was sent and received successfully.
	default:
		log.Println("Failed to send a notification due to blocked listener.")
	}
}
