// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the alliance selection page.

var websocket;

// Sends a websocket message to show the timer.
const showTimer = function() {
  websocket.send("showTimer");
};

// Sends a websocket message to hide the timer.
const hideTimer = function() {
  websocket.send("hideTimer");
}

// Handles a websocket message to update the alliance selection status.
const handleAllianceSelection = function(data) {
  $("#timer").text(getCountdownString(data.TimeRemainingSec));
};

$(function() {
  // Activate playoff tournament datetime picker.
  const startTime = moment(new Date()).hour(13).minute(0).second(0);
  newDateTimePicker("startTimePicker", startTime.toDate());

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/alliance_selection/websocket", {
    allianceSelection: function(event) { handleAllianceSelection(event.data); },
  });
});
