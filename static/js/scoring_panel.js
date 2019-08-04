// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var alliance;

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#matchName").text(data.MatchType + " " + data.Match.DisplayName);
  if (alliance === "red") {
    $("#team1").text(data.Match.Red1);
    $("#team2").text(data.Match.Red2);
    $("#team3").text(data.Match.Red3);
  } else {
    $("#team1").text(data.Match.Blue1);
    $("#team2").text(data.Match.Blue2);
    $("#team3").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the realtime scoring fields.
var handleRealtimeScore = function(data) {
  var realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  var score = realtimeScore.Score;

  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#robotStartLevel" + i1 + ">.value").text(getRobotStartLevelText(score.RobotStartLevels[i]));
    $("#robotStartLevel" + i1).attr("data-value", score.RobotStartLevels[i]);
    $("#sandstormBonus" + i1 + ">.value").text(score.SandstormBonuses[i] ? "Yes" : "No");
    $("#sandstormBonus" + i1).attr("data-value", score.SandstormBonuses[i]);
    $("#robotEndLevel" + i1 + ">.value").text(getRobotEndLevelText(score.RobotEndLevels[i]));
    $("#robotEndLevel" + i1).attr("data-value", score.RobotEndLevels[i]);
    getBay("rocketNearLeft", i).attr("data-value", score.RocketNearLeftBays[i]);
    getBay("rocketNearRight", i).attr("data-value", score.RocketNearRightBays[i]);
    getBay("rocketFarLeft", i).attr("data-value", score.RocketFarLeftBays[i]);
    getBay("rocketFarRight", i).attr("data-value", score.RocketFarRightBays[i]);
  }
  for (var i = 0; i < 8; i++) {
    getBay("cargoShip", i).attr("data-value", score.CargoBays[i]);
  }

  if (matchStates[data.MatchState] === "PRE_MATCH") {
    if (realtimeScore.IsPreMatchScoreReady) {
      $("#preMatchMessage").hide();
    } else {
      $("#preMatchMessage").css("display", "flex");
    }
  }
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      // Pre-match message state is set in handleRealtimeScore().
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
      break;
    case "POST_MATCH":
      $("#preMatchMessage").hide();
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
      $("#preMatchMessage").hide();
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  websocket.send(String.fromCharCode(event.keyCode));
};

// Handles an element click and sends the appropriate websocket message.
var handleClick = function(shortcut) {
  websocket.send(shortcut);
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

// Returns the display text corresponding to the given integer start level value.
var getRobotStartLevelText = function(level) {
  switch (level) {
    case 1:
      return "1";
    case 2:
      return "2";
    case 3:
      return "No-Show";
    default:
      return " ";
  }
};

// Returns the display text corresponding to the given integer end level value.
var getRobotEndLevelText = function(level) {
  switch (level) {
    case 1:
      return "1";
    case 2:
      return "2";
    case 3:
      return "3";
    default:
      return "Not On";
  }
};

// Returns the bay element matching the given parameters.
var getBay = function(type, index) {
  return $("#bay" + bayMappings[type][index]);
}

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $(".alliance-color").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); }
  });

  $(document).keypress(handleKeyPress);
});
