// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Publish-subscribe model for nonblocking notification of server events to websocket clients.

package websocket

import (
	"log"
	"sync"
)

// Allow the listeners to buffer a small number of notifications to streamline delivery.
const notifyBufferSize = 5

type Notifier struct {
	messageType     string
	messageProducer func() any
	listeners       map[chan messageEnvelope]struct{} // The map is essentially a set; the value is ignored.
	mutex           sync.Mutex
}

type messageEnvelope struct {
	messageType string
	messageBody any
}

func NewNotifier(messageType string, messageProducer func() any) *Notifier {
	notifier := &Notifier{messageType: messageType, messageProducer: messageProducer}
	notifier.listeners = make(map[chan messageEnvelope]struct{})
	return notifier
}

// Calls the messageProducer function and sends a message containing the results to all registered listeners, and cleans
// up any listeners that have closed.
func (notifier *Notifier) Notify() {
	notifier.NotifyWithMessage(notifier.getMessageBody())
}

// Sends the given message to all registered listeners, and cleans up any listeners that have closed. If there is a
// messageProducer function defined it is ignored.
func (notifier *Notifier) NotifyWithMessage(messageBody any) {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()

	message := messageEnvelope{messageType: notifier.messageType, messageBody: messageBody}
	for listener := range notifier.listeners {
		notifier.notifyListener(listener, message)
	}
}

func (notifier *Notifier) notifyListener(listener chan messageEnvelope, message messageEnvelope) {

	// 2) Otherwise proceed as normal:
	defer func() {
		if r := recover(); r != nil {
			delete(notifier.listeners, listener)
		}
	}()

	select {
	case listener <- message:
		// sent successfully
	default:
		// 1) If this notifier is the stationTrip topic, just drop the message entirely.
		if notifier.messageType == "stationTrip" {
			return
		}
		// only log blocked‐listener warnings for non‐stationTrip topics
		log.Printf("Failed to send a '%s' notification due to blocked listener.", notifier.messageType)
	}
}

// Registers and returns a channel that can be read from to receive notification messages. The caller is
// responsible for closing the channel, which will cause it to be reaped from the list of listeners.
func (notifier *Notifier) listen() chan messageEnvelope {
	notifier.mutex.Lock()
	defer notifier.mutex.Unlock()

	listener := make(chan messageEnvelope, notifyBufferSize)
	notifier.listeners[listener] = struct{}{}
	return listener
}

// Invokes the message producer to get the message, or returns nil if no producer is defined.
func (notifier *Notifier) getMessageBody() any {
	if notifier.messageProducer == nil {
		return nil
	} else {
		return notifier.messageProducer()
	}
}
