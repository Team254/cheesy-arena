// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the web page display.

var websocket;

$(function() {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  $("#webpageFrame").attr("src", urlParams.get("url"));

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/webpage/websocket", {
  });
});
