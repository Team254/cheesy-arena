// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
var scoreIsReady;

var substituteTeam = function(team, position) {
  websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

var toggleBypass = function(station) {
  websocket.send("toggleBypass", station);
};

var startMatch = function() {
  websocket.send("startMatch");
};

var abortMatch = function() {
  websocket.send("abortMatch");
};

var setAudienceDisplay = function() {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

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

var handleStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    if (stationStatus.DsConn) {
      var dsStatus = stationStatus.DsConn.DriverStationStatus;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsStatus.DsLinked);
      $("#status" + station + " .robot-status").attr("data-status-ok", dsStatus.RobotLinked);
      if (stationStatus.DsConn.SecondsSinceLastRobotConnection > 1 && stationStatus.DsConn.SecondsSinceLastRobotConnection < 1000) {
        $("#status" + station + " .robot-status").text(
          stationStatus.DsConn.SecondsSinceLastRobotConnection.toFixed());
      } else {
        $("#status" + station + " .robot-status").text("");
      }
      $("#status" + station + " .battery-status").attr("data-status-ok",
                                                       dsStatus.BatteryVoltage > 6 && dsStatus.RobotLinked);
      $("#status" + station + " .battery-status").text(dsStatus.BatteryVoltage.toFixed(1) + "V");
      $("#status" + station + " .trip-time").attr("data-status-ok", true);
      $("#status" + station + " .trip-time").text(dsStatus.DsRobotTripTimeMs.toFixed(1) + "ms");
      $("#status" + station + " .packet-loss").attr("data-status-ok", true);
      $("#status" + station + " .packet-loss").text((dsStatus.MissedPacketCount - dsStatus.MissedOffset).toFixed() + "p");
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");
      $("#status" + station + " .battery-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").text("");
      $("#status" + station + " .trip-time").attr("data-status-ok", "");
      $("#status" + station + " .trip-time").text("");
      $("#status" + station + " .packet-loss").attr("data-status-ok", "");
      $("#status" + station + " .packet-loss").text("");
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

var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

var handleSetAudienceDisplay = function(data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

var handleScoringStatus = function(data) {
  scoreIsReady = data.RefereeScoreReady && data.RedScoreReady && data.BlueScoreReady;
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  $("#redScoreStatus").attr("data-ready", data.RedScoreReady);
  $("#blueScoreStatus").attr("data-ready", data.BlueScoreReady);
};

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
    setAudienceDisplay: function(event) { handleSetAudienceDisplay(event.data); },
    scoringStatus: function(event) { handleScoringStatus(event.data); },
    setAllianceStationDisplay: function(event) { handleSetAllianceStationDisplay(event.data); }
  });
});
