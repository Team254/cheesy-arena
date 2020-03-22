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
  var summary = realtimeScore.ScoreSummary;

  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#exitedInitiationLine" + i1 + ">.value").text(score.ExitedInitiationLine[i] ? "Yes" : "No");
    $("#exitedInitiationLine" + i1).attr("data-value", score.ExitedInitiationLine[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score.EndgameStatuses[i]);
    setGoalValue($("#autoCellsInner"), score.AutoCellsInner);
    setGoalValue($("#autoCellsOuter"), score.AutoCellsOuter);
    setGoalValue($("#autoCellsBottom"), score.AutoCellsBottom);
    setGoalValue($("#teleopCellsInner"), score.TeleopCellsInner);
    setGoalValue($("#teleopCellsOuter"), score.TeleopCellsOuter);
    setGoalValue($("#teleopCellsBottom"), score.TeleopCellsBottom);
  }

  if (score.ControlPanelStatus >= 1) {
    $("#rotationControl>.value").text("Yes");
    $("#rotationControl").attr("data-value", true);
  } else if (summary.StagePowerCellsRemaining[1] === 0) {
    $("#rotationControl>.value").text("Unlocked");
    $("#rotationControl").attr("data-value", false);
  } else {
    $("#rotationControl>.value").text("Disabled (" + summary.StagePowerCellsRemaining[1] + " left)");
    $("#rotationControl").attr("data-value", "disabled");
  }
  if (score.ControlPanelStatus === 2) {
    $("#positionControl>.value").text("Yes");
    $("#positionControl").attr("data-value", true);
  } else if (summary.StagePowerCellsRemaining[2] === 0) {
    $("#positionControl>.value").text("Unlocked");
    $("#positionControl").attr("data-value", false);
  } else {
    $("#positionControl>.value").text("Disabled (" + summary.StagePowerCellsRemaining[2] + " left)");
    $("#positionControl").attr("data-value", "disabled");
  }
  $("#rungIsLevel>.value").text(score.RungIsLevel ? "Yes" : "No");
  $("#rungIsLevel").attr("data-value", score.RungIsLevel);
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
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
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

// Returns the display text corresponding to the given integer endgame status value.
var getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Park";
    case 2:
      return "Hang";
    default:
      return "None";
  }
};

// Updates the power cell count for a goal, given the element and score values.
var setGoalValue = function(element, powerCells) {
  var total = 0;
  $.each(powerCells, function(k, v) {
    total += v;
  });
  element.text(total);
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); }
  });

  $(document).keypress(handleKeyPress);
});
