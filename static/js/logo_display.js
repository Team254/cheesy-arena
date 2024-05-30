// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the logo display.

var websocket;

$(function() {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  const message = urlParams.get("message");
  const messageDiv = $("#message");
  messageDiv.text(message);
  messageDiv.toggle(message !== "");

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/logo/websocket", {});
});
