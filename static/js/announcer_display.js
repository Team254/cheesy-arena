// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
var blinkTimeout;

var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
    if (matchState == "PRE_MATCH" || matchState == "POST_MATCH") {
      $("#savedMatchResult").show();
    }
  });
};

var postMatchResult = function(data) {
  clearTimeout(blinkTimeout);
  $("#savedMatchResult").attr("data-blink", false);
  websocket.send("setAudienceDisplay", "score");
}

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/announcer/websocket", {
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); }
  });

  // Make the score blink.
  blinkTimeout = setInterval(function() {
    var blinkOn = $("#savedMatchResult").attr("data-blink") == "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
