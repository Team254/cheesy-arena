// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the bracket display.

var websocket;

// Handles a websocket message to load a new match.
const handleMatchLoad = function(data) {
  fetch("/api/bracket/svg?activeMatch=current")
    .then(response => response.text())
    .then(svg => $("#bracket").html(svg));
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/bracket/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
  });
});
