// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the queueing display.

var websocket;
var firstMatchLoad = true;

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  // Since the server always sends a matchLoad message upon establishing the websocket connection, ignore the first one.
  if (!firstMatchLoad) {
    location.reload();
  }
  firstMatchLoad = false;
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);
  });
};

// Handles a websocket message to update the event status message.
var handleEventStatus = function(data) {
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/queueing/websocket", {
    eventStatus: function(event) { handleEventStatus(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
  });
});
