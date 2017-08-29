// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
)

func TestNotifier(t *testing.T) {
	notifier := NewNotifier()

	// Should do nothing when there are no listeners.
	notifier.Notify("test message")
	notifier.Notify(12345)
	notifier.Notify(struct{}{})

	listener := notifier.Listen()
	notifier.Notify("test message")
	assert.Equal(t, "test message", <-listener)
	notifier.Notify(12345)
	assert.Equal(t, 12345, <-listener)

	// Should allow multiple messages without blocking.
	notifier.Notify("message1")
	notifier.Notify("message2")
	notifier.Notify("message3")
	assert.Equal(t, "message1", <-listener)
	assert.Equal(t, "message2", <-listener)
	assert.Equal(t, "message3", <-listener)

	// Should stop sending messages and not block once the buffer is full.
	log.SetOutput(ioutil.Discard) // Silence noisy log output.
	for i := 0; i < 20; i++ {
		notifier.Notify(i)
	}
	var value interface{}
	var lastValue interface{}
	for lastValue == nil {
		select {
		case value = <-listener:
		default:
			lastValue = value
			return
		}
	}
	notifier.Notify("next message")
	assert.True(t, lastValue.(int) < 10)
	assert.Equal(t, "next message", <-listener)
}

func TestNotifyMultipleListeners(t *testing.T) {
	notifier := NewNotifier()
	listeners := [50]chan interface{}{}
	for i := 0; i < len(listeners); i++ {
		listeners[i] = notifier.Listen()
	}

	notifier.Notify("test message")
	notifier.Notify(12345)
	for listener, _ := range notifier.listeners {
		assert.Equal(t, "test message", <-listener)
		assert.Equal(t, 12345, <-listener)
	}

	// Should reap closed channels automatically.
	close(listeners[4])
	notifier.Notify("message1")
	assert.Equal(t, 49, len(notifier.listeners))
	for listener, _ := range notifier.listeners {
		assert.Equal(t, "message1", <-listener)
	}
	close(listeners[16])
	close(listeners[21])
	close(listeners[49])
	notifier.Notify("message2")
	assert.Equal(t, 46, len(notifier.listeners))
	for listener, _ := range notifier.listeners {
		assert.Equal(t, "message2", <-listener)
	}
}
