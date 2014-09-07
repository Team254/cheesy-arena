// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;

// Handles a websocket message to update the realtime scoring fields.
var handleScore = function(data) {
  // Update autonomous period values.
  var score = data.CurrentScore;
  $("#autoPreloadedBalls").text(data.AutoPreloadedBalls);
  $("#autoMobilityBonuses").text(score.AutoMobilityBonuses);
  $("#autoHighHot").text(score.AutoHighHot);
  $("#autoHigh").text(score.AutoHigh);
  $("#autoLowHot").text(score.AutoLowHot);
  $("#autoLow").text(score.AutoLow);
  var unscoredBalls = data.AutoPreloadedBalls - score.AutoHighHot - score.AutoHigh - score.AutoLowHot -
      score.AutoLow;
  $("#autoUnscoredBalls").text(unscoredBalls);

  // Update teleoperated period current cycle values.
  var cycle = data.CurrentCycle;
  $("#assists").text(cycle.Assists);
  $("#truss").text(cycle.Truss ? "X" : "");
  $("#catch").text(cycle.Catch ? "X" : "");
  $("#scoredHigh").text(cycle.ScoredHigh ? "X" : "");
  $("#scoredLow").text(cycle.ScoredLow ? "X" : "");
  $("#deadBall").text(cycle.DeadBall ? "X" : "");
  if (cycle.ScoredHigh || cycle.ScoredLow || cycle.DeadBall) {
    $("#teleopMessage").html("Press Enter to commit cycle.<br />This cannot be undone.");
  } else if (data.AutoLeftoverBalls > 0) {
    $("#teleopMessage").html(data.AutoLeftoverBalls + " leftover preload" + ((data.AutoLeftoverBalls > 1) ? "s" : ""));
  } else {
    $("#teleopMessage").text("");
  }

  // Update component visibility.
  if (!data.AutoCommitted) {
    $("#autoCommands").show();
    $("#autoScore").show();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#commitMatchScore").show();
    $("#waitingMessage").hide();
  } else if (!data.TeleopCommitted) {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").show();
    $("#teleopScore").show();
    $("#commitMatchScore").show();
    $("#waitingMessage").hide();
  } else {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#commitMatchScore").hide();
    $("#waitingMessage").show();
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  var key = String.fromCharCode(event.keyCode);
  switch(key) {
    case "0":
    case "1":
    case "2":
    case "3":
    case "4":
    case "5":
    case "6":
    case "7":
    case "8":
    case "9":
      websocket.send("preload", key);
      break;
    case "m":
      websocket.send("mobility");
      break;
    case "H":
      websocket.send("scoredHighHot");
      break;
    case "h":
      websocket.send("scoredHigh");
      break;
    case "L":
      websocket.send("scoredLowHot");
      break;
    case "l":
      websocket.send("scoredLow");
      break;
    case "a":
      websocket.send("assist");
      break;
    case "t":
      websocket.send("truss");
      break;
    case "c":
      websocket.send("catch");
      break;
    case "d":
      websocket.send("deadBall");
      break;
    case "\r":
      websocket.send("commit");
      break;
    case "u":
      websocket.send("undo");
      break;
  }
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commitMatch");
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/scoring/" + alliance + "/websocket", {
    score: function(event) { handleScore(event.data); }
  });

  $(document).keypress(handleKeyPress);
});
