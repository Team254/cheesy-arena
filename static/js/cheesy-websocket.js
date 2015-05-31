// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared code for initiating websocket connections back to the server for full-duplex communication.

var CheesyWebsocket = function(path, events) {
  var that = this;

  var url = "ws://" + window.location.hostname;
  if (window.location.port != "") {
    url += ":" + window.location.port;
  }
  url += path;

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function(event) {
      // Data is just an error string.
      console.log(event.data);
    }
  }

  // Insert an event to show a dialog when the server wishes it.
  events.dialog = function(event) {
    alert(event.data);
  }

  // Insert an event to allow the server to force-reload the client for any display.
  events.reload = function(event) {
    location.reload();
  };

  this.connect = function() {
    this.websocket = $.websocket(url, {
      open: function() {
        console.log("Websocket connected to the server at " + url + ".")
      },
      close: function() {
        console.log("Websocket lost connection to the server. Reconnecting in 3 seconds...");
        setTimeout(that.connect, 3000);
      },
      events: events
    });
  };

  this.send = function(type, data) {
    this.websocket.send(type, data);
  };

  this.connect();
};
