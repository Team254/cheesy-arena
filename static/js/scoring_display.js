// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var scoreCommitted = false;

// Handles a websocket message to update the realtime scoring fields.
var handleScore = function(data) {
  // Update autonomous period values.
  var score = data.Score.CurrentScore;
  $("#autoMobility").text(score.AutoMobility);
  $("#autoGears").text(score.AutoGears);
  $("#autoRotors").text(data.ScoreSummary.Rotors);

  // Update teleoperated period values.
  $("#teleopAutoGears").text(score.AutoGears);
  $("#totalGears").text(score.AutoGears + score.Gears);
  $("#totalRotors").text(data.ScoreSummary.Rotors);

  // Update component visibility.
  if (!data.AutoCommitted) {
    $("#autoScoring").fadeTo(0, 1);
    $("#teleopScoring").hide();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else if (!data.Score.TeleopCommitted) {
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
  var key = String.fromCharCode(event.keyCode);
  switch (key) {
    case "m":
      websocket.send("mobility");
      break;
    case "M":
      websocket.send("undoMobility");
      break;
    case "g":
      websocket.send("gear");
      break;
    case "G":
      websocket.send("undoGear");
      break;
    case "\r":
      websocket.send("commit");
      break;
    case "a":
      websocket.send("uncommitAuto");
      break;
  }
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  if (matchStates[data.MatchState] == "POST_MATCH" && !scoreCommitted) {
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
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/scoring/" + alliance + "/websocket", {
    score: function(event) { handleScore(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); }
  });

  $(document).keypress(handleKeyPress);
});
