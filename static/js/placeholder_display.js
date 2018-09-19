// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the placeholder display.

var websocket;

$(function() {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  $("#displayId").text(urlParams.get("displayId"));
  var nickname = urlParams.get("nickname");
  if (nickname !== null) {
    $("#displayNickname").text(nickname);
  }

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/display/websocket", {
  });
});
