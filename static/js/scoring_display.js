// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var scoreCommitted = false;
var alliance;

// Handles a websocket message to update the realtime scoring fields.
var handleRealtimeScore = function(data) {
  var realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red.RealtimeScore;
  } else {
    realtimeScore = data.Blue.RealtimeScore;
  }

  // Update autonomous period values.
  var score = realtimeScore.CurrentScore;
  $("#autoRuns").text(score.AutoRuns);
  $("#climbs").text(score.Climbs);
  $("#parks").text(score.Parks);

  // Update component visibility.
  if (!realtimeScore.AutoCommitted) {
    $("#autoScoring").fadeTo(0, 1);
    $("#teleopScoring").hide();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else if (!realtimeScore.TeleopCommitted) {
    $("#autoScoring").fadeTo(0, 0.25);
    $("#teleopScoring").show();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else {
    $("#autoScoring").hide();
    $("#teleopScoring").hide();
    $("#commitMatchScore").hide();
    $("#waitingMessage").show();
    scoreCommitted = true;
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  websocket.send(String.fromCharCode(event.keyCode));
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  if (matchStates[data.MatchState] === "POST_MATCH" && !scoreCommitted) {
    $("#commitMatchScore").show();
  } else {
    $("#commitMatchScore").hide();
  }
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commitMatch");
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/scoring/" + alliance + "/websocket", {
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); }
  });

  $(document).keypress(handleKeyPress);
});
