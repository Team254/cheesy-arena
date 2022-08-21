// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the bracket display.

var websocket;

// Handles a websocket message to populate the final score data, which also triggers a bracket update.
const handleScorePosted = function(data) {
  $("#bracketSvg").attr("src", "/api/bracket/svg?v=" + new Date().getTime());
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/bracket/websocket", {
    scorePosted: function(event) { handleScorePosted(event.data); },
  });
});
