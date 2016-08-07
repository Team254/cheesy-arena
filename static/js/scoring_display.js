// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var scoreCommitted = false;

// Handles a websocket message to update the realtime scoring fields.
var handleScore = function(data) {
  // Update autonomous period values.
  var score = data.CurrentScore;
  $("#autoDefense1Crossings").text(score.AutoDefensesCrossed[0]);
  $("#autoDefense2Crossings").text(score.AutoDefensesCrossed[1]);
  $("#autoDefense3Crossings").text(score.AutoDefensesCrossed[2]);
  $("#autoDefense4Crossings").text(score.AutoDefensesCrossed[3]);
  $("#autoDefense5Crossings").text(score.AutoDefensesCrossed[4]);
  $("#autoDefensesReached").text(score.AutoDefensesReached);
  $("#autoHighGoals").text(score.AutoHighGoals);
  $("#autoLowGoals").text(score.AutoLowGoals);

  // Update teleoperated period values.
  $("#defense1Crossings").text(score.DefensesCrossed[0] + " (" + score.AutoDefensesCrossed[0] + " in auto)");
  $("#defense2Crossings").text(score.DefensesCrossed[1] + " (" + score.AutoDefensesCrossed[1] + " in auto)");
  $("#defense3Crossings").text(score.DefensesCrossed[2] + " (" + score.AutoDefensesCrossed[2] + " in auto)");
  $("#defense4Crossings").text(score.DefensesCrossed[3] + " (" + score.AutoDefensesCrossed[3] + " in auto)");
  $("#defense5Crossings").text(score.DefensesCrossed[4] + " (" + score.AutoDefensesCrossed[4] + " in auto)");
  $("#highGoals").text(score.HighGoals);
  $("#lowGoals").text(score.LowGoals);
  $("#challenges").text(score.Challenges);
  $("#scales").text(score.Scales);

  // Update component visibility.
  if (!data.AutoCommitted) {
    $("#autoCommands").show();
    $("#autoScore").show();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else if (!data.TeleopCommitted) {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").show();
    $("#teleopScore").show();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#commitMatchScore").hide();
    $("#waitingMessage").show();
    scoreCommitted = true;
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  var key = String.fromCharCode(event.keyCode);
  switch (key) {
    case "1":
    case "2":
    case "3":
    case "4":
    case "5":
      websocket.send("defenseCrossed", key);
      break;
    case "!":
      websocket.send("undoDefenseCrossed", "1");
      break;
    case "@":
      websocket.send("undoDefenseCrossed", "2");
      break;
    case "#":
      websocket.send("undoDefenseCrossed", "3");
      break;
    case "$":
      websocket.send("undoDefenseCrossed", "4");
      break;
    case "%":
      websocket.send("undoDefenseCrossed", "5");
      break;
    case "r":
      websocket.send("autoDefenseReached");
      break;
    case "R":
      websocket.send("undoAutoDefenseReached");
      break;
    case "h":
      websocket.send("highGoal");
      break;
    case "H":
      websocket.send("undoHighGoal");
      break;
    case "l":
      websocket.send("lowGoal");
      break;
    case "L":
      websocket.send("undoLowGoal");
      break;
    case "c":
      websocket.send("challenge");
      break;
    case "C":
      websocket.send("undoChallenge");
      break;
    case "s":
      websocket.send("scale");
      break;
    case "S":
      websocket.send("undoScale");
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
