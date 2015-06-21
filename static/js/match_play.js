// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
var scoreIsReady;

// Sends a websocket message to load a team into an alliance station.
var substituteTeam = function(team, position) {
  websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

// Sends a websocket message to toggle the bypass status for an alliance station.
var toggleBypass = function(station) {
  websocket.send("toggleBypass", station);
};

// Sends a websocket message to start the match.
var startMatch = function() {
  websocket.send("startMatch", { muteMatchSounds: $("#muteMatchSounds").prop("checked") });
};

// Sends a websocket message to abort the match.
var abortMatch = function() {
  websocket.send("abortMatch");
};

// Sends a websocket message to commit the match score and load the next match.
var commitResults = function() {
  websocket.send("commitResults");
};

// Sends a websocket message to discard the match score and load the next match.
var discardResults = function() {
  websocket.send("discardResults");
};

// Sends a websocket message to change what the audience display is showing.
var setAudienceDisplay = function() {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

// Sends a websocket message to change what the alliance station display is showing.
var setAllianceStationDisplay = function() {
  websocket.send("setAllianceStationDisplay", $("input[name=allianceStationDisplay]:checked").val());
};

var confirmCommit = function(isReplay) {
  if (isReplay || !scoreIsReady) {
    // Show the appropriate message(s) in the confirmation dialog.
    $("#confirmCommitReplay").css("display", isReplay ? "block" : "none");
    $("#confirmCommitNotReady").css("display", scoreIsReady ? "none" : "block");
    $("#confirmCommitResults").modal("show");
  } else {
    commitResults();
  }
};

// Handles a websocket message to update the team connection status.
var handleStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    if (stationStatus.DsConn) {
      var dsStatus = stationStatus.DsConn.DriverStationStatus;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsStatus.DsLinked);
      $("#status" + station + " .ds-status").text(dsStatus.MBpsToRobot.toFixed(1) + "/" + dsStatus.MBpsFromRobot.toFixed(1));
      $("#status" + station + " .robot-status").attr("data-status-ok", dsStatus.RobotLinked);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        $("#status" + station + " .robot-status").text("");
      }
      var lowBatteryThreshold = 6;
      if (matchStates[data.MatchState] == "PRE_MATCH") {
        lowBatteryThreshold = 12;
      }
      $("#status" + station + " .battery-status").attr("data-status-ok",
          dsStatus.BatteryVoltage > lowBatteryThreshold && dsStatus.RobotLinked);
      $("#status" + station + " .battery-status").text(dsStatus.BatteryVoltage.toFixed(1) + "V");
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .ds-status").text("");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");
      $("#status" + station + " .battery-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").text("");
    }

    if (stationStatus.EmergencyStop) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("ES");
    } else if (stationStatus.Bypass) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("B");
    } else {
      $("#status" + station + " .bypass-status").attr("data-status-ok", true);
      $("#status" + station + " .bypass-status").text("");
    }
  });

  // Enable/disable the buttons based on the current match state.
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      $("#startMatch").prop("disabled", !data.CanStartMatch);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      break;
    case "START_MATCH":
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
    case "ENDGAME_PERIOD":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      break;
    case "POST_MATCH":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", false);
      $("#discardResults").prop("disabled", false);
      break;
  }
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.RedScore);
  $("#blueScore").text(data.BlueScore);
};

// Handles a websocket message to update the audience display screen selector.
var handleSetAudienceDisplay = function(data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to signal whether the referee and scorers have committed after the match.
var handleScoringStatus = function(data) {
  scoreIsReady = data.RefereeScoreReady && data.RedScoreReady && data.BlueScoreReady;
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  $("#redScoreStatus").attr("data-ready", data.RedScoreReady);
  $("#blueScoreStatus").attr("data-ready", data.BlueScoreReady);
};

// Handles a websocket message to update the alliance station display screen selector.
var handleSetAllianceStationDisplay = function(data) {
  $("input[name=allianceStationDisplay]:checked").prop("checked", false);
  $("input[name=allianceStationDisplay][value=" + data + "]").prop("checked", true);
};

$(function() {
  // Activate tooltips above the status headers.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/match_play/websocket", {
    status: function(event) { handleStatus(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    setAudienceDisplay: function(event) { handleSetAudienceDisplay(event.data); },
    scoringStatus: function(event) { handleScoringStatus(event.data); },
    setAllianceStationDisplay: function(event) { handleSetAllianceStationDisplay(event.data); }
  });
});
