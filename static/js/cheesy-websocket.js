// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared code for initiating websocket connections back to the server for full-duplex communication.

var CheesyWebsocket = function(path, events) {
  var that = this;
  var protocol = "ws://";
  if (window.location.protocol === "https:") {
    protocol = "wss://";
  }
  var url = protocol + window.location.hostname;
  if (window.location.port !== "") {
    url += ":" + window.location.port;
  }
  url += path;

  // Append the page's query string to the websocket URL.
  url += window.location.search;

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function(event) {
      // Data is just an error string.
      console.log(event.data);
      alert(event.data);
    };
  }

  // Parse the display parameters that will be present in the query string if this is a display.
  var displayId = new URLSearchParams(window.location.search).get("displayId");

  // Insert an event to allow the server to force-reload the client for any display.
  events.reload = function(event) {
    if (event.data === null || event.data === displayId) {
      location.reload();
    }
  };

  // Insert an event to allow reconfiguration if this is a display.
  if (!events.hasOwnProperty("displayConfiguration")) {
    events.displayConfiguration = function (event) {
      var newUrl = event.data;

      // Reload the display if the configuration has changed.
      if (newUrl !== window.location.pathname + window.location.search) {
        window.location = newUrl;
      }
    };
  }

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
