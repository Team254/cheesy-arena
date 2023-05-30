// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);
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

// Handles a websocket message to update the match status.
const handleMatchTime = function(data) {
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

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function(data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    $(`#mobilityStatus${i1}>.value`).text(score.MobilityStatuses[i] ? "Yes" : "No");
    $("#mobilityStatus" + i1).attr("data-value", score.MobilityStatuses[i]);
    $("#autoDockStatus" + i1 + ">.value").text(score.AutoDockStatuses[i] ? "Yes" : "No");
    $("#autoDockStatus" + i1).attr("data-value", score.AutoDockStatuses[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score.EndgameStatuses[i]);
  }

  $("#autoChargeStationLevel>.value").text(score.AutoChargeStationLevel ? "Level" : "Not Level");
  $("#autoChargeStationLevel").attr("data-value", score.AutoChargeStationLevel);
  $("#endgameChargeStationLevel>.value").text(score.EndgameChargeStationLevel ? "Level" : "Not Level");
  $("#endgameChargeStationLevel").attr("data-value", score.EndgameChargeStationLevel);

  for (let i = 0; i < 3; i++) {
    for (let j = 0; j < 9; j++) {
      $(`#gridAutoScoringRow${i}Node${j}`).attr("data-value", score.Grid.AutoScoring[i][j]);
      $(`#gridNodeStatesRow${i}Node${j}`).children().each(function() {
        const element = $(this);
        element.attr("data-value", element.attr("data-node-state") === score.Grid.Nodes[i][j].toString());
      });
    }
  }
};

// Handles an element click and sends the appropriate websocket message.
const handleClick = function(command, teamPosition = 0, gridRow = 0, gridNode = 0, nodeState = 0) {
  websocket.send(command, {TeamPosition: teamPosition, GridRow: gridRow, GridNode: gridNode, NodeState: nodeState});
};

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

// Returns the display text corresponding to the given integer endgame status value.
const getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Park";
    case 2:
      return "Dock";
    default:
      return "None";
  }
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});
