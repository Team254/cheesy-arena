// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the alliance selection page.

var websocket;

// Sends a websocket message to set the timer to the given time limit.
const setTimer = function (timeLimitInput) {
  document.getElementById("timeLimitSecInput").value = timeLimitInput.value;
  websocket.send("setTimer", parseInt(timeLimitInput.value));
}

// Sends a websocket message to start and show the timer.
const startTimer = function () {
  websocket.send("startTimer");
};

// Sends a websocket message to hide the timer.
const hideTimer = function() {
  websocket.send("hideTimer");
}

// Sends a websocket message to stop the timer.
const stopTimer = function () {
  websocket.send("stopTimer");
}

// Sends a websocket message to restart the timer. 
const restartTimer = function () {
  websocket.send("restartTimer");
};

// Handles a websocket message to update the alliance selection status.
const handleAllianceSelection = function (data) {
  $("#timer").text(getCountdownString(data.TimeRemainingSec));
};

// Handles a websocket message to update the audience display screen selector.
const handleAudienceDisplayMode = function (data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Sends a websocket message to change what the audience display is showing.
const setAudienceDisplay = function () {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

$(function () {
  // Activate playoff tournament datetime picker.
  const startTime = moment(new Date()).hour(13).minute(0).second(0);
  newDateTimePicker("startTimePicker", startTime.toDate());

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/alliance_selection/websocket", {
    allianceSelection: function (event) {
      handleAllianceSelection(event.data);
    },
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
  });
});
